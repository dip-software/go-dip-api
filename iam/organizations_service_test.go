package iam

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testOrg = `{
    "schemas": [
        "urn:ietf:params:scim:schemas:core:philips:hsdp:2.0:Organization"
    ],
    "id": "c57b2625-eda3-4b27-a8e6-86f0a0e76afc",
    "externalId": "7da36044-d004-44d2-a0b5-9aac72f1e3ad",
    "name": "DCOrg",
    "displayName": "DC Organization",
    "description": "DC Hospital Organization",
    "parent": {
        "value": "cb8d013c-a32b-4bd8-92d6-ce2d2690349a",
        "$ref": "https://idm.host.com/authorize/scim/v2/Organizations/cb8d013c-a32b-4bd8-92d6-ce2d2690349a"
    },
    "type": "Hospital",
    "active": true,
    "inheritProperties": true,
    "address": {
        "formatted": "9780, Hall Street, Adams Boulevard, California, US 90001",
        "streetAddress": "9780, Hall Street",
        "locality": "Adams Boulevard",
        "region": "California",
        "postalCode": "90001",
        "country": "US"
    },
    "owners": [
        {
            "value": "1d725079-b351-4199-9fec-3e796cc82b37",
            "primary": true
        }
    ],
    "createdBy": {
        "value": "1d725079-b351-4199-9fec-3e796cc82b37"
    },
    "modifiedBy": {
        "value": "1d725079-b351-4199-9fec-3e796cc82b37"
    },
    "meta": {
        "resourceType": "Organization",
        "created": "2019-04-30T11:57:58.001Z",
        "lastModified": "2019-04-30T11:57:58.001Z",
        "location": "https://idm.host.com/authorize/scim/v2/Organizations/c57b2625-eda3-4b27-a8e6-86f0a0e76afc",
        "version": "W/\"550012545\""
    }
}`
)

func TestCreateOrganization(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	parentOrgID := "cb8d013c-a32b-4bd8-92d6-ce2d2690349a"
	orgName := "DCOrg"
	orgDescription := "DC Hospital Organization"
	orgID := "c57b2625-eda3-4b27-a8e6-86f0a0e76afc"

	muxIDM.HandleFunc("/authorize/scim/v2/Organizations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got ‘%s’", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Expected body to be read: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var newOrg Organization
		err = json.Unmarshal(body, &newOrg)
		if err != nil {
			t.Errorf("Expected orgnization in body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if newOrg.Parent.Value != parentOrgID {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, testOrg)
	})
	var newOrg Organization
	newOrg.Name = orgName
	newOrg.Description = orgDescription
	newOrg.Parent.Value = parentOrgID

	createdOrg, resp, err := client.Organizations.CreateOrganization(newOrg)
	assert.Nil(t, err)
	if !assert.NotNil(t, resp) {
		return
	}
	if !assert.NotNil(t, createdOrg) {
		return
	}
	assert.Equal(t, http.StatusCreated, resp.StatusCode())
	assert.Equal(t, orgName, createdOrg.Name)
	assert.Equal(t, orgID, createdOrg.ID)
	assert.Equal(t, parentOrgID, createdOrg.Parent.Value)
}

func TestGetOrganizationByID(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	orgUUID := "c57b2625-eda3-4b27-a8e6-86f0a0e76afc"
	orgName := "DCOrg"

	muxIDM.HandleFunc("/authorize/scim/v2/Organizations/"+orgUUID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected ‘GET’ request, got ‘%s’", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, testOrg)
	})

	foundOrg, resp, err := client.Organizations.GetOrganizationByID(orgUUID)
	assert.Nil(t, err)
	if !assert.NotNil(t, resp) {
		return
	}
	if !assert.NotNil(t, foundOrg) {
		return
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, orgName, foundOrg.Name)
	assert.Equal(t, orgUUID, foundOrg.ID)
}

func TestUpdateAndDeleteOrganization(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	orgUUID := "c57b2625-eda3-4b27-a8e6-86f0a0e76afc"
	ifMatch := "W/\"550012545\""

	muxIDM.HandleFunc("/authorize/scim/v2/Organizations/"+orgUUID, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, testOrg)
		case "PUT":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("Expected body to be read: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			var updatedOrg Organization
			err = json.Unmarshal(body, &updatedOrg)
			if err != nil {
				t.Errorf("Expected orgnization in body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if r.Header.Get("If-Match") != ifMatch {
				w.WriteHeader(http.StatusConflict)
				return
			}
			// Update here

			responseBody, err := json.Marshal(updatedOrg)
			if err != nil {
				t.Errorf("Expected orgnization in body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(responseBody)
		case "DELETE":
			if r.Header.Get("If-Method") != "DELETE" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusAccepted)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
	foundOrg, resp, err := client.Organizations.GetOrganizationByID(orgUUID)
	assert.Nil(t, err)
	if !assert.NotNil(t, resp) {
		return
	}
	if !assert.NotNil(t, foundOrg) {
		return
	}

	updatedOrg, resp, err := client.Organizations.UpdateOrganization(*foundOrg)
	if !assert.Nil(t, err) {
		return
	}
	if !assert.NotNil(t, resp) {
		return
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	if !assert.NotNil(t, updatedOrg) {
		return
	}
	assert.Nil(t, err)
	assert.Equal(t, orgUUID, updatedOrg.ID)

	deleted, resp, err := client.Organizations.DeleteOrganization(*updatedOrg)
	if !assert.NotNil(t, resp) {
		return
	}
	assert.True(t, deleted)
	assert.Nil(t, err)

}

func TestGetOrganization(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	orgUUID := "c57b2625-eda3-4b27-a8e6-86f0a0e76afc"
	orgName := "DCOrg"

	muxIDM.HandleFunc("/authorize/scim/v2/Organizations/"+orgUUID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected ‘GET’ request, got ‘%s’", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, testOrg)
	})

	muxIDM.HandleFunc("/authorize/scim/v2/Organizations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected ‘GET’ request, got ‘%s’", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{
  "schemas": [
    "urn:ietf:params:scim:api:messages:2.0:ListResponse"
  ],
  "totalResults": -1,
  "Resources": [
    {
      "schemas": [
        "urn:ietf:params:scim:schemas:core:philips:hsdp:2.0:Organization"
      ],
      "id": "`+orgUUID+`"
    }
  ],
  "startIndex": 1,
  "itemsPerPage": 1
}`)
	})

	foundOrg, resp, err := client.Organizations.GetOrganization(FilterOrgEq(orgUUID))
	assert.Nil(t, err)
	if !assert.NotNil(t, resp) {
		return
	}
	if !assert.NotNil(t, foundOrg) {
		return
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, orgName, foundOrg.Name)
	assert.Equal(t, orgUUID, foundOrg.ID)
}

func TestDeleteStatus(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	orgUUID := "c57b2625-eda3-4b27-a8e6-86f0a0e76afc"

	muxIDM.HandleFunc("/authorize/scim/v2/Organizations/"+orgUUID+"/deleteStatus", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{
  "schemas": [
    "urn:ietf:params:scim:api:messages:philips:hsdp:2.0:DeleteStatus"
  ],
  "id": "`+orgUUID+`",
  "status": "IN_PROGRESS",
  "totalResources": 100,
  "meta": {
    "resourceType": "Organization",
    "created": "2020-05-24T20:33:59.192Z",
    "lastModified": "2020-05-24T20:33:59.192Z",
    "location": "https://idm-xx.us-east.philips-healthsuite.com/authorize/scim/v2/Organizations/`+orgUUID+`",
    "version": "W/\"f250dd84f0671c3\""
  }
}`)
	})

	status, resp, err := client.Organizations.DeleteStatus(orgUUID)
	assert.Nil(t, err)
	if !assert.NotNil(t, resp) {
		return
	}
	if !assert.NotNil(t, status) {
		return
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "IN_PROGRESS", status.Status)
	assert.Equal(t, orgUUID, status.ID)
}

func TestFilters(t *testing.T) {
	opts := FilterParentEq("xxx")
	if !assert.NotNil(t, opts) {
		return
	}
	assert.Equal(t, `parent.value eq "xxx"`, *opts.Filter)
	opts = FilterOrgEq("yyy")
	if !assert.NotNil(t, opts) {
		return
	}
	assert.Equal(t, `id eq "yyy"`, *opts.Filter)
	opts = FilterNameEq("zzz")
	if !assert.NotNil(t, opts) {
		return
	}
	assert.Equal(t, `name eq "zzz"`, *opts.Filter)
}
