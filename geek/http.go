package geek

import (
	"fmt"
	"geek-cache/geek/consistenthash"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geek_cache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self        string                 // self ip
	basePath    string                 // prefix path for communicating
	mu          sync.Mutex             // guards
	peers       *consistenthash.Map    // stores the list of peers, selected by specific key
	httpGetters map[string]*httpGetter // keyed by e.g. "http://10.0.0.2:8009"
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log info
func (p *HTTPPool) Log(format string, path ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, path...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO：直接panic是否不太友好
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// required /<base_url>/<group_name>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]

	g := GetGroup(groupName)
	if g == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}

	view, err := g.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.b)
}

// -------------Client---------------

type httpGetter struct {
	baseURL string // the base URL of remote server
}

// resure implemented
var _ PeerGetter = (*httpGetter)(nil)

// Get send the url for getting specific group and key,
// and return the result
func (h *httpGetter) Get(group, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key))
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}

// add peer to cluster, create the httpGetter function for every peer
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{
			baseURL: peer + p.basePath,
		}
	}
}

// resure implemented
var _ PeerPicker = (*HTTPPool)(nil)

// PickPeer pick a peer with the consistenthash algorithm
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}
