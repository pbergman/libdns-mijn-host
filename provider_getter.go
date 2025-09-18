package mijn_host

import (
	"context"
	"time"

	"github.com/libdns/libdns"
	"github.com/pbergman/libdns-mijn-host/client"
)

func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	items, err := p.client.GetDNSRecords(ctx, zone)

	if err != nil {
		return nil, err
	}

	return p.marshallClientRecords(zone, items)
}

func (p *Provider) marshallClientRecords(zone string, items []*client.DNSRecord) ([]libdns.Record, error) {

	var records = make([]libdns.Record, len(items))
	var err error

	for idx, record := range items {

		rr := &libdns.RR{
			TTL:  time.Duration(record.TTL) * time.Second,
			Data: record.Value,
			Type: record.Type,
			Name: libdns.RelativeName(record.Name, zone),
		}

		if records[idx], err = rr.Parse(); err != nil {
			return nil, err
		}
	}

	return records, nil
}
