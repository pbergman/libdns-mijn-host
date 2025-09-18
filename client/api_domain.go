package client

import (
	"context"
)

type domainsData struct {
	Domains []*Domain `json:"domains"`
}

type Domain struct {
	Id          int      `json:"id"`
	Domain      string   `json:"domain"`
	RenewalDate string   `json:"renewal_date"`
	Status      string   `json:"status"`
	StatusId    int      `json:"status_id"`
	Tags        []string `json:"tags"`
}

func (a *ApiClient) GetDomains(ctx context.Context) ([]*Domain, error) {

	var object struct {
		status
		Data *domainsData `json:"data"`
	}

	if err := a.fetch(ctx, "domains/", "GET", nil, &object); err != nil {
		return nil, err
	}

	return object.Data.Domains, nil
}
