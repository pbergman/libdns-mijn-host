package mijn_host

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/url"
	"os"
	"sync"

	"github.com/libdns/libdns"
	"github.com/pbergman/libdns-mijn-host/client"
)

func NewProvider() *Provider {
	return &Provider{
		client: client.NewApiClient("", nil),
	}
}

type clientConfig struct {
	ApiKey  string `json:"api_key,omitempty"`
	Debug   bool   `json:"debug"`
	BaseUri string `json:"base_uri,omitempty"`
}

type Provider struct {
	client   *client.ApiClient
	mutex    sync.RWMutex
	resolver *net.Resolver
}

func (p *Provider) UnmarshalJSON(b []byte) error {

	var data *clientConfig

	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&data); err != nil {
		return err
	}

	if data.Debug {
		p.SetDebug(os.Stdout)
	}

	if "" != data.BaseUri {
		if err := p.SetBaseUrl(data.BaseUri); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) MarshalJSON() ([]byte, error) {

	var config = &clientConfig{
		ApiKey: p.client.GetApiKey(),
		Debug:  p.client.GetDebug() == nil,
	}

	if nil != p.client.GetBaseUrl() {
		config.BaseUri = p.client.GetBaseUrl().String()
	}

	return json.Marshal(config)
}

func (p *Provider) SetBaseUrl(base string) error {
	return p.client.SetBaseUrl(base)
}

func (p *Provider) GetBaseUrl() *url.URL {
	return p.client.GetBaseUrl()
}

func (p *Provider) SetApiKey(key string) {
	p.client.SetApiKey(key)
}

func (p *Provider) GetApiKey() string {
	return p.client.GetApiKey()
}

func (p *Provider) SetDebug(writer io.Writer) {
	p.client.SetDebug(writer)
}

func (p *Provider) IsDebug() bool {
	return nil != p.client.GetDebug()
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
	_ libdns.ZoneLister     = (*Provider)(nil)
)
