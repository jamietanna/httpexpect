package httpexpect

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func createChunkedHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Proto != "HTTP/1.1" {
			w.WriteHeader(http.StatusBadRequest)
		} else if len(r.TransferEncoding) != 1 || r.TransferEncoding[0] != "chunked" {
			w.WriteHeader(http.StatusBadRequest)
		} else if r.PostFormValue("key") != "value" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[1, `))
			w.(http.Flusher).Flush()
			_, _ = w.Write([]byte(`2]`))
		}
	})

	return mux
}

func testChunkedHandler(e *Expect) {
	e.PUT("/").
		WithHeader("Content-Type", "application/x-www-form-urlencoded").
		WithChunked(strings.NewReader("key=value")).
		Expect().
		Status(http.StatusOK).
		ContentType("application/json").
		TransferEncoding("chunked").
		JSON().Array().Elements(1, 2)
}

func TestE2EChunkedLive(t *testing.T) {
	handler := createChunkedHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	testChunkedHandler(New(t, server.URL))
}

func TestE2EChunkedBinderStandard(t *testing.T) {
	handler := createChunkedHandler()

	testChunkedHandler(WithConfig(Config{
		BaseURL:  "http://example.com",
		Reporter: NewAssertReporter(t),
		Client: &http.Client{
			Transport: NewBinder(handler),
		},
	}))
}
