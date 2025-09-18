package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func NewApiClient(key string, debug io.Writer) *ApiClient {

	client := &ApiClient{
		client: &http.Client{
			Transport: &apiTransport{
				inner:  http.DefaultTransport,
				apiKey: key,
				debug:  debug,
			},
		},
	}

	_ = client.SetBaseUrl("https://mijn.host/api/v2/")

	return client
}

type ApiClient struct {
	client *http.Client
}

func (p *ApiClient) getTransport() *apiTransport {
	if transport, ok := p.client.Transport.(*apiTransport); ok {
		return transport
	}
	return nil
}

func (p *ApiClient) SetDebug(writer io.Writer) {
	if transport := p.getTransport(); nil != transport {
		transport.debug = writer
	}
}

func (p *ApiClient) GetDebug() io.Writer {
	if transport := p.getTransport(); nil != transport {
		return transport.debug
	}
	return nil
}

func (p *ApiClient) SetApiKey(key string) {
	if transport := p.getTransport(); nil != transport {
		transport.apiKey = key
	}
}

func (p *ApiClient) GetApiKey() string {
	if transport := p.getTransport(); nil != transport {
		return transport.apiKey
	}
	return ""
}

func (p *ApiClient) GetBaseUrl() *url.URL {
	if transport := p.getTransport(); nil != transport {
		return transport.baseUri
	}
	return nil
}

func (p *ApiClient) SetBaseUrl(base string) error {
	if transport := p.getTransport(); nil != transport {
		uri, err := url.Parse(base)
		if err != nil {
			return err
		}
		transport.baseUri = uri
	}
	return nil
}

func (a *ApiClient) toDnsPath(domain string) string {
	return fmt.Sprintf("domains/%s/dns", url.PathEscape(strings.TrimSuffix(domain, ".")))
}

func (a *ApiClient) fetch(ctx context.Context, path string, method string, body io.Reader, object any) error {

	request, err := http.NewRequestWithContext(ctx, method, path, body)

	if err != nil {
		return err
	}

	response, err := a.client.Do(request)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if !strings.HasPrefix(response.Header.Get("content-type"), "application/json") {
		return fmt.Errorf("unexpected response type: %s", response.Header.Get("content-type"))
	}

	if nil != object {

		if err := json.NewDecoder(response.Body).Decode(object); err != nil {
			return err
		}

		if v, ok := object.(StatusResponse); ok {
			if err := v.Error(); err != nil {
				return err
			}
		}
	}

	return nil
}
