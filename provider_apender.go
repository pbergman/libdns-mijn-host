package mijn_host

import (
	"context"

	"github.com/libdns/libdns"
)

func (p *Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	records, err := p.client.GetDNSRecords(ctx, zone)

	if err != nil {
		return nil, err
	}

	var size = len(records)

	for x := range MarshallIterator(zone, RecordIterator(recs)) {
		if x.Value != "" || x.Name != "" || x.Type != "" {
			records = append(records, x)
		}
	}

	if len(records) > size {
		if err := p.client.SetDNSRecords(ctx, zone, records); err != nil {
			return nil, err
		}
	}

	return p.marshallClientRecords(zone, records[size:])
}
