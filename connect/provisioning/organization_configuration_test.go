package provisioning_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/dip-software/go-dip-api/connect/provisioning"

	"github.com/stretchr/testify/assert"
)

func OrgConfigurationBody(orgConfigID, organizationGuid, serviceAccountID, serviceAccountKey, publicKey string) string {
	body := map[string]interface{}{
		"resourceType":     "OrgConfiguration",
		"id":               orgConfigID,
		"organizationGuid": organizationGuid,
		"serviceAccount": map[string]string{
			"serviceAccountId":  serviceAccountID,
			"serviceAccountKey": serviceAccountKey,
		},
		"bootstrapSignature": map[string]interface{}{
			"algorithm": "RSA-SHA256",
			"publicKey": publicKey,
			"config": map[string]string{
				"type":       "RSA",
				"padding":    "RSA_PKCS1_PSS_PADDING",
				"saltLength": "RSA_PSS_SALTLEN_MAX_SIGN",
			},
		},
	}
	b, _ := json.MarshalIndent(body, "", "  ")
	return string(b)
}

func TestOrgConfigCRUD(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	organizationGuid := "dbf1d779-ab9f-4c27-b4aa-ea75f9efbbc1"
	serviceAccountID := "service-demo.test-account-id.iot__connect__sandbox.apmplatform.philips-healthsuite.com"
	serviceAccountKey := "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBlahBlahBlahncryI/0qC019+ihpI1KExUm\n-----END RSA PRIVATE KEY-----"
	publicKey := "-----BEGIN PUBLIC KEY-----MIIBblahBlahBlahDAQAB-----END PUBLIC KEY-----"
	orgConfigID := "786b71ad-50e2-4e5d-bdb9-087140f3806d"
	muxProvisioning.HandleFunc("/client-test/connect/provisioning/OrgConfiguration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case "POST":
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, OrgConfigurationBody(orgConfigID, organizationGuid, serviceAccountID, serviceAccountKey, publicKey))
		}
	})
	muxProvisioning.HandleFunc("/client-test/connect/provisioning/OrgConfiguration/"+orgConfigID, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case "GET":
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, OrgConfigurationBody(orgConfigID, organizationGuid, serviceAccountID, serviceAccountKey, publicKey))
		case "PUT":
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, OrgConfigurationBody(orgConfigID, organizationGuid, serviceAccountID, serviceAccountKey, publicKey))
		case "DELETE":
			w.WriteHeader(http.StatusNoContent)
		}
	})

	created, resp, err := provisioningClient.OrgConfigurationsService.CreateOrganizationConfiguration(provisioning.OrgConfiguration{
		ResourceType:     "OrgConfiguration",
		OrganizationGuid: organizationGuid,
		ServiceAccount: provisioning.ServiceAccount{
			ServiceAccountId:  serviceAccountID,
			ServiceAccountKey: serviceAccountKey,
		},
		BootstrapSignature: provisioning.BootstrapSignature{
			PublicKey: publicKey,
			Algorithm: "RSA-SHA256",
			Config: provisioning.BootStrapSignatureConfig{
				Type:       "RSA",
				Padding:    "RSA_PKCS1_PSS_PADDING",
				SaltLength: "RSA_PSS_SALTLEN_MAX_SIGN",
			},
		},
	})
	if !assert.Nil(t, err) {
		return
	}
	if !assert.NotNil(t, resp) {
		return
	}
	if !assert.NotNil(t, created) {
		return
	}
	assert.Equal(t, organizationGuid, created.OrganizationGuid)
	assert.Equal(t, orgConfigID, created.ID)

	res, resp, err := provisioningClient.OrgConfigurationsService.DeleteOrganizationConfiguration(*created)
	if !assert.Nil(t, err) {
		return
	}
	if !assert.NotNil(t, resp) {
		return
	}
	assert.True(t, res)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode())
}
