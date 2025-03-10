package internal_test

import (
	"github.com/dip-software/go-dip-api/internal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVersion(t *testing.T) {
	assert.True(t, len(internal.LibraryVersion) > 0)
}
