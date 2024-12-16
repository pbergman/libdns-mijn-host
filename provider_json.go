package mijn_host

import (
	"bytes"
	"encoding/json"
	"os"
)

func (p *Provider) UnmarshalJSON(b []byte) error {

	var data struct {
		ApiKey  string `json:"api_key"`
		Debug   bool   `json:"debug"`
		BaseUri string `json:"base_uri"`
	}

	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&data); err != nil {
		return err
	}

	if transport := p.client.getTransport(); nil != transport {

		transport.apiKey = data.ApiKey

		if data.Debug {
			transport.debug = os.Stdout
		}

		if "" != data.BaseUri {
			if err := p.SetBaseUrl(data.BaseUri); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Provider) MarshalJSON() ([]byte, error) {
	var data = make(map[string]interface{})

	if transport := p.client.getTransport(); nil != transport {
		data["api_key"] = transport.apiKey
		data["debug"] = transport.debug != nil

		if nil != transport.baseUri {
			data["base_uri"] = transport.baseUri.String()
		}
	}

	return json.Marshal(data)
}
