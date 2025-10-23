package provisioning

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dip-software/go-dip-api/internal"
	"github.com/go-playground/validator/v10"
)

type SubscribersService struct {
	*Client
}

var (
	orgConfiguratioAPIVersion = "1"
)

type OrgConfiguration struct {
	ResourceType       string             `json:"resourceType" validate:"required" enum:"OrgConfiguration"`
	ID                 string             `json:"id,omitempty"`
	OrganizationGuid   string             `json:"organizationGuid" validate:"required"`
	ServiceAccount     ServiceAccount     `json:"serviceAccount" validate:"required"`
	BootstrapSignature BootstrapSignature `json:"bootstrapSignature" validate:"required"`
	Meta               *Meta              `json:"meta,omitempty"`
}

type ServiceAccount struct {
	ServiceAccountId  string `json:"serviceAccountId,omitempty" validate:"required"`
	ServiceAccountKey string `json:"serviceAccountKey,omitempty" validate:"required"`
}

type BootstrapSignature struct {
	PublicKey string                   `json:"publicKey,omitempty" validate:"required"`
	Algorithm string                   `json:"algorithm,omitempty" validate:"required" enum:"SHA256|SHA512|RSA-MD5|RSA-RIPEMD160|RSA-SHA1|RSA-SHA1-2|RSA-SHA224|RSA-SHA256|RSA-SHA3-224|RSA-SHA3-256|RSA-SHA3-384|RSA-SHA3-512|RSA-SHA384|RSA-SHA512"`
	Config    BootStrapSignatureConfig `json:"config,omitempty"`
}

type BootStrapSignatureConfig struct {
	Type       string `json:"type" enum:"RSA|ECC|DSA"`
	Padding    string `json:"padding" enum:"RSA_PKCS1_PSS_PADDING"`
	SaltLength string `json:"saltLength" enum:"RSA_PSS_SALTLEN_DIGEST|RSA_PSS_SALTLEN_MAX_SIGN|RSA_PSS_SALTLEN_AUTO"`
}

// GetOrgConfiguration struct describes search criteria for looking up OrgConfiguration
type GetOrgConfiguration struct {
	ID               *string `url:"_id,omitempty"`
	OrganizationGuid *string `url:"_organizationGuid,omitempty"`
}

type Meta struct {
	LastUpdated time.Time `json:"lastUpdated,omitempty"`
	VersionID   int       `json:"versionId,omitempty"`
}
type OrgConfigurationsService struct {
	*Client
	validate *validator.Validate
}

func (b *OrgConfigurationsService) CreateOrganizationConfiguration(orgConfig OrgConfiguration) (*OrgConfiguration, *Response, error) {
	orgConfig.ResourceType = "OrgConfiguration"
	if err := b.validate.Struct(orgConfig); err != nil {
		return nil, nil, err
	}

	req, _ := b.NewRequest(http.MethodPost, "/OrgConfiguration", orgConfig, nil)
	req.Header.Set("api-version", orgConfiguratioAPIVersion)
	req.Header.Set("Content-Type", "application/json")

	var created OrgConfiguration

	resp, err := b.Do(req, &created)

	if err != nil {
		return nil, resp, err
	}
	if created.ID == "" {
		return nil, resp, fmt.Errorf("the 'ID' field is missing")
	}
	return &created, resp, nil
}

// UpdateOrgConfiguration updates a OrgConfiguration
func (b *OrgConfigurationsService) UpdateOrganizationConfiguration(orgConfig OrgConfiguration) (*OrgConfiguration, *Response, error) {
	orgConfig.ResourceType = "OrgConfiguration"
	id := orgConfig.ID
	orgConfig.ID = "" // Server does not like a value here
	if err := b.validate.Struct(orgConfig); err != nil {
		return nil, nil, err
	}
	req, _ := b.NewRequest(http.MethodPut, "/OrgConfiguration/"+id, orgConfig, nil)
	req.Header.Set("api-version", orgConfiguratioAPIVersion)
	req.Header.Set("Content-Type", "application/json")

	var updated OrgConfiguration

	resp, err := b.Do(req, &updated)
	if err != nil {
		return nil, resp, err
	}
	return &updated, resp, nil
}

func (b *OrgConfigurationsService) DeleteOrganizationConfiguration(orgConfig OrgConfiguration) (bool, *Response, error) {
	req, err := b.NewRequest(http.MethodDelete, "/OrgConfiguration/"+orgConfig.ID, nil, nil)
	if err != nil {
		return false, nil, err
	}
	req.Header.Set("api-version", orgConfiguratioAPIVersion)

	var deleteResponse interface{}

	resp, err := b.Do(req, &deleteResponse)
	if resp == nil || resp.StatusCode() != http.StatusNoContent {
		return false, resp, err
	}
	return true, resp, nil
}

func (b *OrgConfigurationsService) GetOrganizationConfigurationByID(id string) (*OrgConfiguration, *Response, error) {
	req, err := b.NewRequest(http.MethodGet, "/OrgConfiguration/"+id, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("api-version", orgConfiguratioAPIVersion)
	req.Header.Set("Content-Type", "application/json")

	var resource OrgConfiguration

	resp, err := b.Do(req, &resource)
	if err != nil {
		return nil, resp, err
	}
	err = internal.CheckResponse(resp.Response)
	if err != nil {
		return nil, resp, fmt.Errorf("GetByID: %w", err)
	}
	if resource.ID != id {
		return nil, nil, fmt.Errorf("returned resource does not match")
	}
	return &resource, resp, nil
}

func (b *OrgConfigurationsService) FindOrgConfiguration(opt *GetOrgConfiguration, options ...OptionFunc) (*[]OrgConfiguration, *Response, error) {
	req, err := b.NewRequest(http.MethodGet, "/OrgConfiguration", opt, options...)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("api-version", orgConfiguratioAPIVersion)
	req.Header.Set("Content-Type", "application/json")

	var bundleResponse internal.Bundle

	resp, err := b.Do(req, &bundleResponse)
	if err != nil {
		return nil, resp, err
	}
	var resources []OrgConfiguration
	for _, c := range bundleResponse.Entry {
		var resource OrgConfiguration
		if err := json.Unmarshal(c.Resource, &resource); err == nil {
			resources = append(resources, resource)
		}
	}

	return &resources, resp, err
}
