package mijn_host

import (
	"iter"

	"github.com/libdns/libdns"
	"github.com/pbergman/libdns-mijn-host/client"
)

type ResourceRecordsIterator interface {
	Iterate() iter.Seq[libdns.RR]
}

func MarshallRecord(zone string, record *libdns.RR) *client.DNSRecord {
	var x = &client.DNSRecord{
		Type:  record.Type,
		Value: record.Data,
		Name:  libdns.AbsoluteName(record.Name, zone) + ".",
	}

	switch record.TTL.Seconds() {
	case 300, 900, 3600, 10800, 21600, 43200, 86400:
		x.TTL = int(record.TTL.Seconds())
	default:
		x.TTL = 900
	}

	return x
}

func MarshallIterator(zone string, iter ResourceRecordsIterator) iter.Seq[*client.DNSRecord] {
	return func(yield func(*client.DNSRecord) bool) {
		for record := range iter.Iterate() {
			if !yield(MarshallRecord(zone, &record)) {
				return
			}
		}
	}
}

func FilteredIterator(items []libdns.Record, exclude []libdns.Record) ResourceRecordsIterator {
	return FilteredRecordIterator{
		list:     items,
		excludes: exclude,
	}
}

type FilteredRecordIterator struct {
	list     []libdns.Record
	excludes []libdns.Record
}

func (l FilteredRecordIterator) inExcludes(a libdns.Record) bool {
	for _, v := range l.excludes {
		if v == a {
			return true
		}
	}
	return false
}

func (l FilteredRecordIterator) Iterate() iter.Seq[libdns.RR] {
	return func(yield func(libdns.RR) bool) {
		for _, record := range l.list {
			if false == l.inExcludes(record) {
				if !yield(record.RR()) {
					return
				}
			}
		}
	}
}

type RecordIterator []libdns.Record

func (l RecordIterator) Iterate() iter.Seq[libdns.RR] {
	return func(yield func(libdns.RR) bool) {
		for _, record := range l {
			if !yield(record.RR()) {
				return
			}
		}
	}
}
