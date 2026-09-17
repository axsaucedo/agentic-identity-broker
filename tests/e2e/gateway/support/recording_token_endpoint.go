package support

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"sync"
)

// TokenEndpointRecorder records token forms while allowing production routing to handle them.
type TokenEndpointRecorder struct {
	mu    sync.RWMutex
	forms []url.Values
}

// Wrap records POST /oauth2/token form values before delegating to next unchanged.
func (r *TokenEndpointRecorder) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost && request.URL.Path == "/oauth2/token" {
			r.record(request)
		}

		next.ServeHTTP(response, request)
	})
}

func (r *TokenEndpointRecorder) record(request *http.Request) {
	if request.Body == nil {
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return
	}
	request.Body = io.NopCloser(bytes.NewReader(body))

	recordedRequest := request.Clone(request.Context())
	recordedRequest.Body = io.NopCloser(bytes.NewReader(body))
	if err := recordedRequest.ParseForm(); err != nil {
		return
	}

	r.mu.Lock()
	r.forms = append(r.forms, cloneTokenEndpointForm(recordedRequest.PostForm))
	r.mu.Unlock()
}

// Forms returns copied form values in request order.
func (r *TokenEndpointRecorder) Forms() []url.Values {
	r.mu.RLock()
	defer r.mu.RUnlock()

	forms := make([]url.Values, len(r.forms))
	for index, form := range r.forms {
		forms[index] = cloneTokenEndpointForm(form)
	}
	return forms
}

// Reset discards every recorded token form.
func (r *TokenEndpointRecorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.forms = nil
}

func cloneTokenEndpointForm(form url.Values) url.Values {
	clone := make(url.Values, len(form))
	for key, values := range form {
		clone[key] = append([]string(nil), values...)
	}
	return clone
}
