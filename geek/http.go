package geek

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_geek_cache/"

type HTTPPool struct {
	self     string // self ip
	basePath string // prefix path for communicating
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
