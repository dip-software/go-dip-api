package audit_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dip-software/go-dip-api/audit"

	"github.com/stretchr/testify/assert"
)

var (
	muxAudit    *http.ServeMux
	serverAudit *httptest.Server
	auditClient *audit.Client
)

func setup(t *testing.T) func() {
	var err error
	muxAudit = http.NewServeMux()
	serverAudit = httptest.NewServer(muxAudit)

	auditClient, err = audit.NewClient(nil, &audit.Config{
		AuditBaseURL: serverAudit.URL,
		SharedKey:    "foo",
		SharedSecret: "bar",
	})
	if !assert.Nil(t, err) {
		t.Fatalf("invalid client")
	}
	return func() {
		serverAudit.Close()
	}
}

func TestDebug(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	_, _, _ = auditClient.CreateAuditEvent(nil)

	defer auditClient.Close()
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}() // clean up

	// TODO: trigger call here

	fi, err := tmpfile.Stat()
	assert.Nil(t, err)
	assert.NotEqual(t, 0, fi.Size(), "Expected something to be written to DebugLog")
}
