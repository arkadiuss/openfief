package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/arkadiuss/openfief/plugins/model-openai/openaichat"
)

// fakeKey is a synthetic API key.
const fakeKey = "synthetic-key-0000"

// fakeProvider is an OpenAI-compatible server: it records the last request
// and answers with status and body.
type fakeProvider struct {
	mu     sync.Mutex
	auth   string
	path   string
	body   map[string]any
	status int
	reply  string
}

func (f *fakeProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)
	var body map[string]any
	if err != nil || json.Unmarshal(raw, &body) != nil {
		http.Error(w, "unreadable request", http.StatusBadRequest)
		return
	}
	f.mu.Lock()
	f.auth, f.path, f.body = r.Header.Get("Authorization"), r.URL.Path, body
	status, reply := f.status, f.reply
	f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := io.WriteString(w, reply); err != nil {
		return // the client went away; the test sees its own error
	}
}

// newFake starts a fake provider that answers 200 with reply.
func newFake(t *testing.T, reply string) (*fakeProvider, *openaichat.Client) {
	t.Helper()
	f := &fakeProvider{status: http.StatusOK, reply: reply}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	return f, &openaichat.Client{HTTP: srv.Client(), BaseURL: srv.URL + "/v1", Key: func() (string, error) { return fakeKey, nil }}
}

const textReply = `{"choices":[{"message":{"role":"assistant","content":"synthetic answer"},"finish_reason":"stop"}],
"usage":{"prompt_tokens":11,"completion_tokens":7}}`
