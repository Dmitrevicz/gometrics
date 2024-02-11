package staticchecks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticcheck(t *testing.T) {
	// get all analyzers
	z := Staticcheck()
	assert.NotEmpty(t, z, "expected all analyzers to be retrieved")

	exclude := make([]string, 0, len(z)+1)
	for _, alz := range z {
		exclude = append(exclude, alz.Name)
	}

	// exclude all
	z = Staticcheck(exclude...)
	assert.Empty(t, z, "expected no analyzers to be retrieved")

	// exclude all + 1
	exclude = append(exclude, "onemore")
	z = Staticcheck(exclude...)
	assert.Empty(t, z, "expected no analyzers to be retrieved")
}

func TestSimple(t *testing.T) {
	// get all analyzers
	z := Simple()
	assert.NotEmpty(t, z, "expected all analyzers to be retrieved")

	exclude := make([]string, 0, len(z)+1)
	for _, alz := range z {
		exclude = append(exclude, alz.Name)
	}

	// exclude all
	z = Simple(exclude...)
	assert.Empty(t, z, "expected no analyzers to be retrieved")

	// exclude all + 1
	exclude = append(exclude, "onemore")
	z = Simple(exclude...)
	assert.Empty(t, z, "expected no analyzers to be retrieved")
}

func TestStylecheck(t *testing.T) {
	// get all analyzers
	z := Stylecheck()
	assert.NotEmpty(t, z, "expected all analyzers to be retrieved")

	exclude := make([]string, 0, len(z)+1)
	for _, alz := range z {
		exclude = append(exclude, alz.Name)
	}

	// exclude all
	z = Stylecheck(exclude...)
	assert.Empty(t, z, "expected no analyzers to be retrieved")

	// exclude all + 1
	exclude = append(exclude, "onemore")
	z = Stylecheck(exclude...)
	assert.Empty(t, z, "expected no analyzers to be retrieved")
}

func TestQuickfix(t *testing.T) {
	// get all analyzers
	z := Quickfix()
	assert.NotEmpty(t, z, "expected all analyzers to be retrieved")

	exclude := make([]string, 0, len(z)+1)
	for _, alz := range z {
		exclude = append(exclude, alz.Name)
	}

	// exclude all
	z = Quickfix(exclude...)
	assert.Empty(t, z, "expected no analyzers to be retrieved")

	// exclude all + 1
	exclude = append(exclude, "onemore")
	z = Quickfix(exclude...)
	assert.Empty(t, z, "expected no analyzers to be retrieved")
}
