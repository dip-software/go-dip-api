package dicom_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/philips-software/go-hsdp-api/dicom"
	"github.com/stretchr/testify/assert"
)

func TestMoveServiceGetSet(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	serviceID := "f5a1e608-6787-4af1-a94a-4cbda7677a9c"

	muxDICOM.HandleFunc("/store/dicom/config/dicom/production/moveService", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/fhir+json")
		switch r.Method {
		case "POST":
			if !assert.Equal(t, "application/json", r.Header.Get("Content-Type")) {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}
			if !assert.Equal(t, dicom.APIVersion, r.Header.Get("API-Version")) {
				w.WriteHeader(http.StatusPreconditionFailed)
				return
			}
			var received dicom.SCPConfig
			err := json.NewDecoder(r.Body).Decode(&received)
			if !assert.Nil(t, err) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			received.ID = serviceID
			resp, err := json.Marshal(&received)
			if !assert.Nil(t, err) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, _ = w.Write(resp)
		case "GET":
			stores := []dicom.SCPConfig{
				{
					ID:          serviceID,
					Description: "Some description",
				},
			}
			resp, err := json.Marshal(&stores)
			if !assert.Nil(t, err) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(resp)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	created, resp, err := dicomClient.Config.SetQueryRetrieveService(dicom.BrokenSCPConfig{
		Title:       "A title here",
		Description: "A description here",
		ApplicationEntities: []dicom.ApplicationEntity{
			{
				AeTitle:  "Foo",
				AllowAny: true,
			},
		},
	}, nil)
	if !assert.Equal(t, nil, err) {
		return
	}
	if !assert.NotNil(t, resp) {
		return
	}
	if !assert.NotNil(t, created) {
		return
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, created.ID, serviceID)

	services, resp, err := dicomClient.Config.GetQueryRetrieveService(nil)
	if !assert.Nil(t, err) {
		return
	}
	if !assert.NotNil(t, resp) {
		return
	}
	if !assert.NotNil(t, services) {
		return
	}
	assert.Equal(t, services.ID, serviceID)
}
