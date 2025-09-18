package client

import (
	"bytes"
	"context"
	"encoding/json"
)

type dnsRecordData struct {
	Domain  string `json:"domain"`
	Records []*DNSRecord
}

type DNSRecord struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

func (a *ApiClient) SetDNSRecords(ctx context.Context, domain string, records []*DNSRecord) error {

	var buf = new(bytes.Buffer)

	if err := json.NewEncoder(buf).Encode(map[string][]*DNSRecord{"records": records}); err != nil {
		return err
	}

	var object status

	if err := a.fetch(ctx, a.toDnsPath(domain), "PUT", buf, &object); err != nil {
		return err
	}

	if err := object.Error(); err != nil {
		return err
	}

	return nil
}

func (a *ApiClient) GetDNSRecords(ctx context.Context, domain string) ([]*DNSRecord, error) {

	var object struct {
		status
		Data *dnsRecordData `json:"data"`
	}

	if err := a.fetch(ctx, a.toDnsPath(domain), "GET", nil, &object); err != nil {
		return nil, err
	}

	return object.Data.Records, nil
}
