# Mijn Host for `libdns`

This package implements the libdns interfaces for the [Mijn Host API](https://mijn.host/api/doc/)

## Authenticating

To authenticate, you need to create am api key [here](https://mijn.host/cp/account/api/).

## Example

Here's a minimal example of how to get all your DNS records using this `libdns` provider

```go
package main

import (
	"context"
	"fmt"
	"text/tabwriter"

	mijn_host "github.com/pbergman/libdns-mijn-host"
)

func main() {
	provider := mijn_host.NewProvider()
	provider.SetApiKey("***************************")
	//provider.SetDebug(os.Stdout)

	zones, err := provider.ListZones(context.Background())

	if err != nil {
		panic(err)
	}

	var writer = tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)

	for _, zone := range zones {
		records, err := provider.GetRecords(context.Background(), zone.Name)

		if err != nil {
			panic(err)
		}

		for record := range mijn_host.RecordIterator(records).Iterate() {
			_, _ = fmt.Fprintf(writer, "%s\t%v\t%s\t%s\n", record.Name, record.TTL.Seconds(), record.Type, record.Data)
		}

	}

	_ = writer.Flush()
}
```