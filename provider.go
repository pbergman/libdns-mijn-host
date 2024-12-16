package mijn_host

import (
	"context"
	"net"
	"sync"

	"github.com/libdns/libdns"
)

func NewProvider() *Provider {
	return &Provider{
		client: NewApiClient("", nil),
	}
}

type Provider struct {
	client *ApiClient

	mutex    sync.RWMutex
	resolver *net.Resolver
}

func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.client.GetRecords(ctx, zone)
}

func (p *Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	existing, err := p.client.GetRecords(ctx, zone)

	if err != nil {
		return nil, err
	}

	if err := p.client.SetRecords(ctx, zone, append(existing, recs...)); err != nil {
		return nil, err
	}

	return recs, nil
}

func (p *Provider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	existing, err := p.client.GetRecords(ctx, zone)
	var update = make([]libdns.Record, 0)

	if err != nil {
		return nil, err
	}

outerLoop:

	for _, a := range existing {

		for _, b := range recs {
			if (b.ID != "" && a.ID == b.ID) || (a.Type == b.Type && a.Name == b.Name && a.Value == b.Value) {
				continue outerLoop
			}
		}

		update = append(update, a)
	}

	if err := p.client.SetRecords(ctx, zone, update); err != nil {
		return nil, err
	}

	return recs, nil
}

func (p *Provider) equals(a, b *libdns.Record) bool {
	return (a.Type == "CNAME" && b.Type == "CNAME" && a.Name == b.Name) || (a.ID != "" && a.ID == b.ID)
}

func (p *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	existing, err := p.client.GetRecords(ctx, zone)

	if err != nil {
		return nil, err
	}

outerLoop:

	for i, c := 0, len(recs); i < c; i++ {
		for x, y := 0, len(existing); x < y; x++ {
			if p.equals(&recs[i], &existing[x]) {
				existing[x] = recs[i]
				continue outerLoop
			}
		}

		if "" == recs[i].ID {
			recs[i].ID = p.client.makeId(&recs[i])
		}

		existing = append(existing, recs[i])
	}

	if err := p.client.SetRecords(ctx, zone, existing); err != nil {
		return nil, err
	}

	return recs, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
