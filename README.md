# Mijn Host for `libdns`

This package implements the libdns interfaces for the [Mijn Host API](https://mijn.host/api/doc/)

## Authenticating

To authenticate you need to create am api key [here](https://mijn.host/cp/account/api/).

## Example

Here's a minimal example of how to get all your DNS records using this `libdns` provider

```go
package main

import (
	"context"
	"fmt"
	
	mijn_host "github.com/pbergman/libdns-mijn-host"
)

func main() {
	provider := mijn_host.NewProvider("api_key")

	records, err := provider.GetRecords(context.Background(), "example.com")
	
	if err != nil {
		panic(err)
	}

	for _, record := range records {
		fmt.Printf("%s %v %s %s\n", record.Name, record.TTL.Seconds(), record.Type, record.Value)
	}
}
```