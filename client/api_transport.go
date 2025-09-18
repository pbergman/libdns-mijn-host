package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Status struct {
	Code        int    `json:"status"`
	Description string `json:"status_description"`
}

type apiTransport struct {
	baseUri *url.URL
	apiKey  string
	debug   io.Writer
	inner   http.RoundTripper
}

func (a *apiTransport) RoundTrip(request *http.Request) (*http.Response, error) {

	if nil != a.baseUri {
		request.URL = a.baseUri.ResolveReference(request.URL)
	}

	request.Header.Set("accept", "application/json")
	request.Header.Set("content-type", "application/json")
	request.Header.Set("api-key", a.apiKey)
	request.Header.Set("user-agent", "libdns-client/1.0")

	if nil != a.debug {
		a.dumpRequest(request, a.debug)
	}

	response, err := a.inner.RoundTrip(request)

	if nil != a.debug {
		a.dumpResponse(response, a.debug)
	}

	return response, err
}

func (a *apiTransport) dumpRequest(r *http.Request, x io.Writer) {
	if out, err := httputil.DumpRequest(r, true); err == nil {
		scanner := bufio.NewScanner(bytes.NewReader(out))
		for scanner.Scan() {
			_, _ = fmt.Fprintf(x, "[>] %s\n", scanner.Text())
		}
	}
}

func (a *apiTransport) dumpResponse(r *http.Response, x io.Writer) {

	if nil == r {
		return
	}

	if out, err := httputil.DumpResponse(r, true); err == nil {
		scanner := bufio.NewScanner(bytes.NewReader(out))
		for scanner.Scan() {
			_, _ = fmt.Fprintf(x, "[<] %s\n", scanner.Text())
		}
	}
}
