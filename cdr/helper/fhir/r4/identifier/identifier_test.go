package identifier_test

import (
	r4dt "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/datatypes_go_proto"
	"testing"

	"github.com/dip-software/go-dip-api/cdr/helper/fhir/r4/identifier"
	r4gp "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/codes_go_proto"
	"github.com/stretchr/testify/assert"
)

func TestIdentifierToString(t *testing.T) {
	val := identifier.UseToString(&r4dt.Identifier_UseCode{
		Value: r4gp.IdentifierUseCode_TEMP,
	})
	assert.Equal(t, "TEMP", val)
}
