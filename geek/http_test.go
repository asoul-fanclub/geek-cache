package geek

import (
	"fmt"
	"io"
	"log"
	"net/http/httptest"
	"testing"
)

var http_test_db = map[string]string{
	"Tom":  "630",
	"Tom2": "631",
	"Tom3": "632",
}

func TestHTTP(t *testing.T) {
	base_test_url := "http://localhost:8080" + defaultBasePath

	req1 := httptest.NewRequest("GET", base_test_url+"scores/Tom", nil)
	req2 := httptest.NewRequest("GET", base_test_url+"scoress/Tom", nil)
	req3 := httptest.NewRequest("GET", base_test_url+"scores/Tom4", nil)
	req4 := httptest.NewRequest("GET", base_test_url+"scores/Tom3", nil)

	w := httptest.NewRecorder()
	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()
	w3 := httptest.NewRecorder()
	NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := http_test_db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not found", key)
		}))
	addr := "localhost:8080"
	peers := NewHTTPPool(addr)
	log.Println("geek-cache is running at", addr)
	peers.ServeHTTP(w, req1)
	bytes, _ := io.ReadAll(w.Result().Body)
	if string(bytes) != http_test_db["Tom"] {
		t.Fatal("expected 630, but got", string(bytes))
	}
	peers.ServeHTTP(w1, req2)
	bytes, _ = io.ReadAll(w1.Result().Body)
	if string(bytes) == "630" {
		t.Fatal("expected not got 630, but got", string(bytes))
	}
	peers.ServeHTTP(w2, req3)
	bytes, _ = io.ReadAll(w2.Result().Body)
	if string(bytes) == http_test_db["Tom4"] {
		t.Fatal("expected got Tom4 not found, but got", string(bytes))
	}
	peers.ServeHTTP(w3, req4)
	bytes, _ = io.ReadAll(w3.Result().Body)
	if string(bytes) != http_test_db["Tom3"] {
		t.Fatal("expected 632, but got", string(bytes))
	}
}
