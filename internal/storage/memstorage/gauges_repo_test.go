package memstorage

import (
	"testing"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGaugesRepo_Get(t *testing.T) {
	s := New()

	gauge := struct {
		name  string
		value model.Gauge
	}{
		name:  "TestCounter",
		value: model.Gauge(42.420),
	}

	t.Run("found", func(t *testing.T) {
		err := s.Gauges().Set(gauge.name, gauge.value)
		require.NoError(t, err)

		got, err := s.Gauges().Get(gauge.name)
		require.NotErrorIs(t, err, storage.ErrNotFound, "unexpected ErrNotFound when error must be nil")
		require.NoError(t, err)
		assert.Equal(t, gauge.value, got)
	})

	t.Run("not found", func(t *testing.T) {
		got, err := s.Gauges().Get("unknown-test-gauge")
		require.Errorf(t, err, "expected nothing (ErrNotFound), but found something - name: %s, gauge: %d", gauge.name, got)
		require.ErrorIs(t, err, storage.ErrNotFound, "expected ErrNotFound")
		assert.EqualValues(t, 0, got)
	})
}

func TestGaugesRepo_GetAll(t *testing.T) {
	counters := []struct {
		name  string
		value model.Gauge
	}{
		{
			name:  "TestCounter1",
			value: model.Gauge(42.420),
		},
		{
			name:  "TestCounter2",
			value: model.Gauge(42.420),
		},
	}

	t.Run("get", func(t *testing.T) {
		s := New()

		for _, c := range counters {
			err := s.Gauges().Set(c.name, c.value)
			require.NoError(t, err)
		}

		got, err := s.Gauges().GetAll()
		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.Len(t, got, len(counters))
	})

	t.Run("get empty", func(t *testing.T) {
		s := New()

		for _, c := range counters {
			err := s.Gauges().Delete(c.name)
			require.NoError(t, err)
		}

		got, err := s.Gauges().GetAll()
		require.NoError(t, err)
		assert.Empty(t, got)
		assert.Len(t, got, 0)
	})
}

func TestGaugesRepo_Set(t *testing.T) {
	s := New()

	gauge := struct {
		name  string
		value model.Gauge
	}{
		name:  "TestCounter",
		value: model.Gauge(42.420),
	}

	err := s.Gauges().Set(gauge.name, gauge.value)
	require.NoError(t, err)

	got, err := s.Gauges().Get(gauge.name)
	require.NotErrorIs(t, err, storage.ErrNotFound, "ErrNotFound: nothing found after metric update attempt")
	require.NoError(t, err)
	assert.Equal(t, gauge.value, got)
}

func TestGaugesRepo_Delete(t *testing.T) {
	s := New()

	gauge := struct {
		name  string
		value model.Gauge
	}{
		name:  "TestCounter",
		value: model.Gauge(42.420),
	}

	err := s.Gauges().Set(gauge.name, gauge.value)
	require.NoError(t, err)

	got, err := s.Gauges().Get(gauge.name)
	require.NotErrorIs(t, err, storage.ErrNotFound, "ErrNotFound: nothing found after metric update attempt")
	require.NoError(t, err)
	assert.Equal(t, gauge.value, got)

	err = s.Gauges().Delete(gauge.name)
	require.NoError(t, err)

	got, err = s.Gauges().Get(gauge.name)
	require.Errorf(t, err, "expected nothing (ErrNotFound), but found something - name: %s, gauge: %d", gauge.name, got)
	require.ErrorIs(t, err, storage.ErrNotFound, "expected ErrNotFound")
}
