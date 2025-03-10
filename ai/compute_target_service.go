package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/dip-software/go-dip-api/internal"
	"github.com/go-playground/validator/v10"
)

type ComputeTargetService struct {
	client *Client

	validate *validator.Validate
}

type ComputeTarget struct {
	ID           string `json:"id,omitempty"`
	ResourceType string `json:"resourceType" Validate:"required"`
	Name         string `json:"name" Validate:"required"`
	Description  string `json:"description"`
	InstanceType string `json:"instanceType" Validate:"required"`
	Storage      int    `json:"storage,omitempty"`
	IsFactory    bool   `json:"isFactory,omitempty"`
	Created      string `json:"created,omitempty"`
	CreatedBy    string `json:"createdBy,omitempty"`
}

func (s *ComputeTargetService) path(components ...string) string {
	return path.Join(components...)
}

func (s *ComputeTargetService) CreateComputeTarget(target ComputeTarget) (*ComputeTarget, *Response, error) {
	if err := s.validate.Struct(target); err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewAIRequest("POST", s.path("ComputeTarget"), target, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Api-Version", APIVersion)

	var createdTarget ComputeTarget
	resp, err := s.client.Do(req, &createdTarget)
	if (err != nil && err != io.EOF) || resp == nil {
		if resp == nil && err != nil {
			err = fmt.Errorf("CreateComputeTarget: %w", ErrEmptyResult)
		}
		return nil, resp, err
	}
	return &createdTarget, resp, nil
}

func (s *ComputeTargetService) DeleteComputeTarget(target ComputeTarget) (*Response, error) {
	req, err := s.client.NewAIRequest("DELETE", s.path("ComputeTarget", target.ID), nil, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Api-Version", APIVersion)

	resp, err := s.client.Do(req, nil)
	if (err != nil && err != io.EOF) || resp == nil {
		if resp == nil && err != nil {
			err = fmt.Errorf("DeleteComputeTarget: %w", ErrEmptyResult)
		}
		return resp, err
	}
	return resp, nil
}

func (s *ComputeTargetService) GetComputeTargetByID(id string) (*ComputeTarget, *Response, error) {
	req, err := s.client.NewAIRequest("GET", s.path("ComputeTarget", id), nil, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Api-Version", APIVersion)

	var foundTarget ComputeTarget
	resp, err := s.client.Do(req, &foundTarget)
	if (err != nil && err != io.EOF) || resp == nil {
		if resp == nil && err != nil {
			err = fmt.Errorf("GetComputeTargetByID: %w", ErrEmptyResult)
		}
		return nil, resp, err
	}
	return &foundTarget, resp, nil
}

func (s *ComputeTargetService) GetComputeTargets(opt *GetOptions, options ...OptionFunc) ([]ComputeTarget, *Response, error) {
	req, err := s.client.NewAIRequest("GET", s.path("ComputeTarget"), opt, options...)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Api-Version", APIVersion)

	var bundleResponse struct {
		ResourceType string                 `json:"resourceType,omitempty"`
		Type         string                 `json:"type,omitempty"`
		Total        int                    `json:"total,omitempty"`
		Entry        []internal.BundleEntry `json:"entry"`
	}
	resp, err := s.client.Do(req, &bundleResponse)
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			return nil, resp, ErrEmptyResult
		}
		return nil, resp, err
	}
	var targets []ComputeTarget
	for _, e := range bundleResponse.Entry {
		var target ComputeTarget
		if err := json.Unmarshal(e.Resource, &target); err == nil {
			targets = append(targets, target)
		}
	}
	return targets, resp, err
}
