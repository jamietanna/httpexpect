package httpexpect

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createAutoTLSHandler(https string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/tls", func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			_, _ = w.Write([]byte(`no`))
		} else {
			_, _ = w.Write([]byte(`yes`))
		}
	})

	mux.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			http.Redirect(w, r, https+r.RequestURI, http.StatusFound)
		} else {
			_, _ = w.Write([]byte(`hello`))
		}
	})

	return mux
}

func testAutoTLSHandler(config Config) {
	e := WithConfig(config)

	tls := e.POST("/tls").
		Expect().
		Status(http.StatusOK).Body()

	if strings.HasPrefix(config.BaseURL, "https://") {
		tls.Equal(`yes`)
	} else {
		tls.Equal(`no`)
	}

	e.POST("/protected").
		Expect().
		Status(http.StatusOK).Body().Equal(`hello`)
}

func TestE2EAutoTLSLive(t *testing.T) {
	httpsServ := httptest.NewTLSServer(createAutoTLSHandler(""))
	defer httpsServ.Close()

	httpServ := httptest.NewServer(createAutoTLSHandler(httpsServ.URL))
	defer httpServ.Close()

	assert.True(t, strings.HasPrefix(httpsServ.URL, "https://"))
	assert.True(t, strings.HasPrefix(httpServ.URL, "http://"))

	for _, url := range []string{httpsServ.URL, httpServ.URL} {
		testAutoTLSHandler(Config{
			BaseURL:  url,
			Reporter: NewRequireReporter(t),
			Printers: []Printer{
				NewDebugPrinter(t, true),
			},
			Client: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			},
		})
	}
}

func TestE2EAutoTLSBinderStandard(t *testing.T) {
	handler := createAutoTLSHandler("https://example.com")

	for _, url := range []string{"https://example.com", "http://example.com"} {
		testAutoTLSHandler(Config{
			BaseURL:  url,
			Reporter: NewRequireReporter(t),
			Printers: []Printer{
				NewDebugPrinter(t, true),
			},
			Client: &http.Client{
				Transport: &Binder{
					Handler: handler,
					TLS:     &tls.ConnectionState{},
				},
			},
		})
	}
}
