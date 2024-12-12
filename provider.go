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

func NewProvider() *Provider {
	return &Provider{
		client: &http.Client{},
	}
}

type status struct {
	Code        int    `json:"status"`
	Description string `json:"status_description"`
}

type Provider struct {
	ApiKey string `json:"api_key"`
	client *http.Client
	mutex  sync.RWMutex
}

func (p *Provider) toPath(zone string) string {
	return fmt.Sprintf("https://mijn.host/api/v2/domains/%s/dns", url.PathEscape(strings.TrimSuffix(zone, ".")))
}

func (p *Provider) newRequest(ctx context.Context, method string, zone string, body io.Reader) (*http.Request, error) {

	request, err := http.NewRequestWithContext(ctx, method, p.toPath(zone), body)

	if err != nil {
		return nil, err
	}

	request.Header.Set("accept", "application/json")
	request.Header.Set("content-type", "application/json")
	request.Header.Set("API-Key", p.ApiKey)

	return request, nil
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

	request, err := p.newRequest(ctx, "PUT", zone, bytes.NewReader(out))

	if err != nil {
		return err
	}

	response, err := p.client.Do(request)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	var object status

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

	for i, c := 0, len(recs); i < c; i++ {
		recs[i].ID = p.generateId(&recs[i])
	}

	return recs, nil
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

	for i, c := 0, len(recs); i < c; i++ {
		recs[i].ID = p.generateId(&recs[i])
	}

	return recs, nil
}

func (p *Provider) generateId(record *libdns.Record) string {
	var hash = sha1.New()
	_, _ = fmt.Fprintf(hash, "[type:%s]", record.Type)
	_, _ = fmt.Fprintf(hash, "[name:%s]", record.Name)
	_, _ = fmt.Fprintf(hash, "[value:%s]", record.Value)
	_, _ = fmt.Fprintf(hash, "[ttl:%s]", record.TTL)

	return hex.EncodeToString(hash.Sum(nil))
}

func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	request, err := p.newRequest(ctx, "GET", zone, nil)

	if err != nil {
		return nil, err
	}

	response, err := p.client.Do(request)

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

	if err := json.NewDecoder(response.Body).Decode(&object); err != nil {
		return nil, err
	}

	if 200 != object.status.Code {
		return nil, errors.New(object.status.Description)
	}

	for i, c := 0, len(*object.Data.Records); i < c; i++ {
		(*object.Data.Records)[i].ID = p.generateId(&(*object.Data.Records)[i])
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
