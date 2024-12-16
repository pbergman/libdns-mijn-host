package mijn_host

import (
	"io"
	"net/url"
)

func (p *Provider) SetBaseUrl(base string) error {
	return p.client.setBaseUrl(base)
}

func (p *Provider) GetBaseUrl() *url.URL {
	return p.client.getBaseUrl()
}

func (p *Provider) SetApiKey(key string) {
	p.client.setApiKey(key)
}

func (p *Provider) GetApiKey() string {
	return p.client.getApiKey()
}

func (p *Provider) SetDebug(writer io.Writer) {
	p.client.setDebug(writer)
}

func (p *Provider) IsDebug() bool {
	return nil != p.client.getDebug()
}
