package geek

// PeerPicker must be implemented to locate the peer that owns a specific key
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter must be implemented by a peer
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
