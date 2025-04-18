package tdr

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/dip-software/go-dip-api/internal"
)

// ContractsService provides operations on TDR contracts
type ContractsService struct {
	client *Client
}

const (
	APIVersion = "5"
)

// GetContractOptions describes the fields on which you can search for contracts
type GetContractOptions struct {
	Organization *string `url:"organization,omitempty"`
	DataType     *string `url:"dataType,omitempty"`
	Count        *int    `url:"_count,omitempty"`
}

// GetContract searches for contracts in TDR
func (c *ContractsService) GetContract(opt *GetContractOptions, options ...OptionFunc) ([]*Contract, *Response, error) {
	var contracts []*Contract

	req, err := c.client.newTDRRequest("GET", "store/tdr/Contract", opt, options)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Api-Version", APIVersion)

	var bundleResponse internal.Bundle

	resp, err := c.client.Do(req, &bundleResponse)
	if err != nil {
		return nil, resp, err
	}
	if bundleResponse.Total == 0 {
		return contracts, resp, ErrEmptyResult
	}
	for _, e := range bundleResponse.Entry {
		c := new(Contract)
		if err := json.Unmarshal(e.Resource, c); err == nil {
			contracts = append(contracts, c)
		} else {
			return nil, resp, err
		}
	}
	return contracts, resp, err
}

// CreateContract creates a new contract in TDR
func (c *ContractsService) CreateContract(contract Contract) (bool, *Response, error) {
	req, err := c.client.newTDRRequest("POST", "store/tdr/Contract", &contract, nil)
	if err != nil {
		return false, nil, err
	}
	req.Header.Set("Api-Version", APIVersion)

	var createResponse bytes.Buffer
	resp, err := c.client.Do(req, &createResponse)
	if err != nil {
		return false, resp, err
	}
	if resp.StatusCode() != http.StatusCreated {
		return false, resp, err
	}
	if location := resp.Header.Get("Location"); location == "" {
		return false, resp, ErrCouldNoReadResourceAfterCreate
	}
	return true, resp, nil
}
