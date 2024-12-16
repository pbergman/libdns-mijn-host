package mijn_host

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/libdns/libdns"
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

	_ = client.setBaseUrl("https://mijn.host/api/v2/")

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

func (p *ApiClient) setDebug(writer io.Writer) {
	if transport := p.getTransport(); nil != transport {
		transport.debug = writer
	}
}

func (p *ApiClient) getDebug() io.Writer {
	if transport := p.getTransport(); nil != transport {
		return transport.debug
	}
	return nil
}

func (p *ApiClient) setApiKey(key string) {
	if transport := p.getTransport(); nil != transport {
		transport.apiKey = key
	}
}

func (p *ApiClient) getApiKey() string {
	if transport := p.getTransport(); nil != transport {
		return transport.apiKey
	}
	return ""
}

func (p *ApiClient) getBaseUrl() *url.URL {
	if transport := p.getTransport(); nil != transport {
		return transport.baseUri
	}
	return nil
}

func (p *ApiClient) setBaseUrl(base string) error {
	if transport := p.getTransport(); nil != transport {
		uri, err := url.Parse(base)
		if err != nil {
			return err
		}
		transport.baseUri = uri
	}
	return nil
}

func (a *ApiClient) toPath(zone string) string {
	return fmt.Sprintf("domains/%s/dns", url.PathEscape(strings.TrimSuffix(zone, ".")))
}

func (a *ApiClient) getSuffix(zone string) string {
	return strings.TrimSuffix(zone, ".") + "."
}

func (a *ApiClient) SetRecords(ctx context.Context, zone string, records []libdns.Record) error {

	payload, err := a.toBody(zone, records)

	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, "PUT", a.toPath(zone), payload)

	if err != nil {
		return err
	}

	response, err := a.client.Do(request)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if !strings.HasPrefix(response.Header.Get("content-type"), "application/json") {
		return errors.New(response.Status)
	}

	var object status

	if err := json.NewDecoder(response.Body).Decode(&object); err != nil {
		return err
	}

	if 200 != object.Code {
		return errors.New(object.Description)
	}

	return nil
}

func (a *ApiClient) toBody(zone string, records []libdns.Record) (io.Reader, error) {
	var hostname = zone

	if zone[len(zone)] != '.' {
		hostname += "."
	}

	type record map[string]interface{}

	var data = map[string][]record{
		"records": make([]record, len(records)),
	}

	for i, c := 0, len(records); i < c; i++ {

		var object = record{
			"type":  records[i].Type,
			"value": records[i].Value,
			"ttl":   a.ttl(records[i].TTL),
		}

		if records[i].Name == "@" {
			object["name"] = hostname
		} else {
			object["name"] = records[i].Name + "." + hostname
		}

		data["records"][i] = object
	}

	payload, err := json.Marshal(data)

	if err != nil {
		return nil, err
	}

	return bytes.NewReader(payload), nil
}

func (a *ApiClient) ttl(t time.Duration) int {
	switch t.Seconds() {
	case 300, 900, 3600, 10800, 21600, 43200, 86400:
		return int(t.Seconds())
	default:
		return 900
	}
}

func (a *ApiClient) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {

	request, err := http.NewRequestWithContext(ctx, "GET", a.toPath(zone), nil)

	if err != nil {
		return nil, err
	}

	response, err := a.client.Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	var object = struct {
		status
		Data struct {
			Domain  string
			Records *[]libdns.Record
		}
	}{}

	if !strings.HasPrefix(response.Header.Get("content-type"), "application/json") {
		return nil, errors.New(response.Status)
	}

	if err := json.NewDecoder(response.Body).Decode(&object); err != nil {
		return nil, err
	}

	if 200 != object.status.Code {
		return nil, errors.New(object.status.Description)
	}

	for i, c := 0, len(*object.Data.Records); i < c; i++ {
		(*object.Data.Records)[i].TTL *= time.Second

		if idx := strings.Index((*object.Data.Records)[i].Name, zone); idx >= 0 {
			(*object.Data.Records)[i].Name = strings.TrimSuffix((*object.Data.Records)[i].Name[:idx], ".")
		}

		if (*object.Data.Records)[i].Name == "" {
			(*object.Data.Records)[i].Name = "@"
		}

		(*object.Data.Records)[i].ID = a.makeId(&(*object.Data.Records)[i])
	}

	return *object.Data.Records, nil
}

func (a *ApiClient) makeId(record *libdns.Record) string {
	var hash = sha1.New()
	_, _ = fmt.Fprintf(hash, "[type:%s]", record.Type)
	_, _ = fmt.Fprintf(hash, "[name:%s]", record.Name)
	_, _ = fmt.Fprintf(hash, "[value:%s]", record.Value)
	_, _ = fmt.Fprintf(hash, "[ttl:%s]", record.TTL)
	return hex.EncodeToString(hash.Sum(nil))
}
