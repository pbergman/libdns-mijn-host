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
	"sync"
	"time"

	"github.com/libdns/libdns"
)

func NewProvider(key string) *Provider {
	provider := &Provider{
		client: &http.Client{
			Transport: &apiTransport{
				apiKey: key,
				inner:  http.DefaultTransport,
			},
		},
	}

	_ = provider.SetBaseUrl("https://mijn.host/api/v2/")

	return provider
}

type Provider struct {
	client *http.Client
	mutex  sync.RWMutex
}

func (p *Provider) SetBaseUrl(base string) error {
	uri, err := url.Parse(base)

	if err != nil {
		return err
	}

	p.client.Transport.(*apiTransport).baseUri = uri

	return nil
}

func (p *Provider) SetApiKey(key string) {
	p.client.Transport.(*apiTransport).apiKey = key
}

func (p *Provider) GetApiKey() string {
	return p.client.Transport.(*apiTransport).apiKey
}

func (p *Provider) SetDebug(writer io.Writer) {
	p.client.Transport.(*apiTransport).debug = writer
}

func (p *Provider) toPath(zone string) string {
	return fmt.Sprintf("domains/%s/dns", url.PathEscape(strings.TrimSuffix(zone, ".")))
}

func (p *Provider) updateRecords(ctx context.Context, zone string, recs []libdns.Record) error {

	type record map[string]interface{}

	var data = map[string][]record{
		"records": make([]record, len(recs)),
	}

	for i, c := 0, len(recs); i < c; i++ {
		data["records"][i] = record{
			"type":  recs[i].Type,
			"name":  recs[i].Name,
			"value": recs[i].Value,
			"ttl":   recs[i].TTL.Seconds(),
		}
	}

	out, err := json.Marshal(data)

	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, "PUT", p.toPath(zone), bytes.NewReader(out))

	if err != nil {
		return err
	}

	response, err := p.client.Do(request)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	var object Status

	if err := json.NewDecoder(response.Body).Decode(&object); err != nil {
		return err
	}

	if 200 != object.Code {
		return errors.New(object.Description)
	}

	return nil
}

func (p *Provider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	records, err := p.GetRecords(ctx, zone)

	if err != nil {
		return nil, err
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

outerLoop:

	for _, remove := range recs {
		if remove.ID != "" {
			for idx, x := range records {
				if x.ID == remove.ID {
					records = append(records[:idx], records[idx+1:]...)
					continue outerLoop
				}
			}
		}
		for idx, x := range records {
			if x.Type == remove.Type && x.Name == remove.Name && x.Value == remove.Value && x.TTL == remove.TTL {
				records = append(records[:idx], records[idx+1:]...)
				continue outerLoop
			}
		}
	}

	if err := p.updateRecords(ctx, zone, records); err != nil {
		return nil, err
	}

	return records, nil
}

func (p *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	records, err := p.GetRecords(ctx, zone)

	if err != nil {
		return nil, err
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

next:
	for i, c := 0, len(recs); i < c; i++ {

		if recs[i].ID != "" {
			for x, y := 0, len(records); x < y; x++ {
				if records[x].ID == recs[i].ID {
					records[x].Name = recs[i].Name
					records[x].Type = recs[i].Type
					records[x].TTL = recs[i].TTL
					records[x].Value = recs[i].Value
				}
				continue next
			}
		}

		if recs[i].Type == "CNAME" {
			for x, y := 0, len(records); x < y; x++ {
				if records[x].Type == "CNAME" && records[x].Name == recs[i].Name {
					records[x].Name = recs[i].Name
					records[x].Type = recs[i].Type
					records[x].TTL = recs[i].TTL
					records[x].Value = recs[i].Value
				}
				continue next
			}
		}

		records = append(records, recs[i])
	}

	if err := p.updateRecords(ctx, zone, records); err != nil {
		return nil, err
	}

	return records, nil
}

func (p *Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	records, err := p.GetRecords(ctx, zone)

	if err != nil {
		return nil, err
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	for i, c := 0, len(recs); i < c; i++ {
		records = append(records, recs[i])
	}

	if err := p.updateRecords(ctx, zone, records); err != nil {
		return nil, err
	}

	return records, nil
}

func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	request, err := http.NewRequestWithContext(ctx, "GET", p.toPath(zone), nil)

	if err != nil {
		return nil, err
	}

	response, err := p.client.Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	var object = struct {
		Status
		Data struct {
			Domain  string
			Records *[]libdns.Record
		}
	}{}

	if err := json.NewDecoder(response.Body).Decode(&object); err != nil {
		return nil, err
	}

	if 200 != object.Status.Code {
		return nil, errors.New(object.Status.Description)
	}

	for i, c := 0, len(*object.Data.Records); i < c; i++ {
		var hash = sha1.New()
		_, _ = fmt.Fprintf(hash, "[type:%s]", (*object.Data.Records)[i].Type)
		_, _ = fmt.Fprintf(hash, "[name:%s]", (*object.Data.Records)[i].Name)
		_, _ = fmt.Fprintf(hash, "[value:%s]", (*object.Data.Records)[i].Value)
		_, _ = fmt.Fprintf(hash, "[ttl:%s]", (*object.Data.Records)[i].TTL)
		(*object.Data.Records)[i].ID = hex.EncodeToString(hash.Sum(nil))
		(*object.Data.Records)[i].TTL *= time.Second
	}

	return *object.Data.Records, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)