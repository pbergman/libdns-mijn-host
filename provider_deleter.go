package mijn_host

import (
	"context"

	"github.com/libdns/libdns"
	"github.com/pbergman/libdns-mijn-host/client"
)

func (p *Provider) DeleteRecords(ctx context.Context, zone string, deletes []libdns.Record) ([]libdns.Record, error) {

	records, err := p.GetRecords(ctx, zone)

	if err != nil {
		return nil, err
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	items, removes := make([]*client.DNSRecord, 0), make([]libdns.Record, 0)

	for _, item := range deletes {
		for _, record := range records {
			if shouldRemove(item, record) {
				removes = append(removes, record)
			}
		}
	}

	if len(removes) <= 0 {
		return nil, nil
	}

	for item := range MarshallIterator(zone, FilteredIterator(records, removes)) {
		items = append(items, item)
	}

	if err := p.client.SetDNSRecords(ctx, zone, items); err != nil {
		return nil, err
	}

	return removes, nil
}

// https://github.com/libdns/libdns/blob/master/libdns.go#L232
func shouldRemove(a, b libdns.Record) bool {

	c, d := a.RR(), b.RR()

	if c.Name != d.Name {
		return false
	}

	if c.Type != d.Type && c.Type != "" {
		return false
	}

	if c.Data != d.Data && c.Data != "" {
		return false
	}

	if c.TTL != d.TTL && c.TTL > 0 {
		return false
	}

	return true
}
