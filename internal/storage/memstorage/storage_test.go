// Package memstorage implements storage repository using map as data storage.

package memstorage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorage_Ping(t *testing.T) {
	s := New()
	require.NotNil(t, s, "nil *Storage")

	err := s.Ping(context.Background())
	require.NoError(t, err)
}

func TestStorage_Close(t *testing.T) {
	s := New()
	require.NotNil(t, s, "nil *Storage")

	err := s.Close(context.Background())
	require.NoError(t, err)
}
