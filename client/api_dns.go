package client

import (
	"bytes"
	"context"
	"encoding/json"
	"time"
)

type dnsRecordData struct {
	Domain  string `json:"domain"`
	Records []*DNSRecord
}

//
//func (r *RecordList) UnmarshalJSON(p []byte) error {
//
//	var x struct {
//		Domain  string `json:"domain"`
//		Records []*Record
//	}
//
//	if err := json.Unmarshal(p, &x); err != nil {
//		return err
//	}
//
//	r.Records = x.Records
//	r.Domain = &Domain{
//		Domain: x.Domain,
//	}
//
//	for idx, _ := range x.Records {
//		x.Records[idx].domain = r.Domain
//	}
//
//	return nil
//}

type DNSRecord struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

//func (r *Record) RR() libdns.RR {
//
//	var name = r.Name
//
//	if r.domain != nil {
//		name = libdns.RelativeName(r.domain.Domain, name)
//	}
//
//	return libdns.RR{
//		TTL:  time.Duration(r.TTL) * time.Second,
//		Data: r.Value,
//		Type: r.Type,
//		Name: name,
//	}
//}

//type RecordIterator []libdns.Record
//
//func (l RecordIterator) ResourceRecord() iter.Seq[libdns.RR] {
//	return func(yield func(libdns.RR) bool) {
//		for i, c := 0, len(l); i < c; i++ {
//			if !yield(l[i].RR()) {
//				return
//			}
//		}
//	}
//}

func toSeconds(t time.Duration) int {
	switch t.Seconds() {
	case 300, 900, 3600, 10800, 21600, 43200, 86400:
		return int(t.Seconds())
	default:
		return 900
	}
}

//func MarshallRecords(domain *Domain, records []libdns.Record) []*Record {
//
//	var items = make([]*Record, 0)
//
//	for record := range RecordIterator(records).ResourceRecord() {
//		items = append(items, &Record{
//			Type:  record.Type,
//			Value: record.Data,
//			TTL:   toSeconds(record.TTL),
//			Name:  libdns.AbsoluteName(domain.Domain, record.Name),
//		})
//	}
//
//	return items
//}

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
