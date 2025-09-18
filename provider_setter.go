package mijn_host

import (
	"context"

	"github.com/libdns/libdns"
	"github.com/pbergman/libdns-mijn-host/client"
)

func (p *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {

	records, err := p.GetRecords(ctx, zone)

	if err != nil {
		return nil, err
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	var set = make([]*client.DNSRecord, 0)
	var req = RecordIterator(recs)

outerLoop:
	for record := range RecordIterator(records).Iterate() {
		for item := range req.Iterate() {
			if item.Name == record.Name && item.Type == record.Type {
				continue outerLoop
			}
		}
		set = append(set, MarshallRecord(zone, &record))
	}

	for item := range req.Iterate() {
		set = append(set, MarshallRecord(zone, &item))
	}

	if err := p.client.SetDNSRecords(ctx, zone, set); err != nil {
		return nil, err
	}

	return recs, nil
}

//	//
//	//existing, err := p.client.GetRecords(ctx, zone)
//	//
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	////outerLoop:
//	//
//	//for i, c := 0, len(recs); i < c; i++ {
//	//	//for x, y := 0, len(existing); x < y; x++ {
//	//	//	if p.equals(&recs[i], &existing[x]) {
//	//	//		existing[x] = recs[i]
//	//	//		continue outerLoop
//	//	//	}
//	//	//}
//	//
//	//	existing = append(existing, recs[i])
//	//}
//	//
//	//if err := p.client.SetRecords(ctx, zone, existing); err != nil {
//	//	return nil, err
//	//}
//
//	return recs, nil
//}
