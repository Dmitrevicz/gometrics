package memstorage

import (
	"testing"

	"github.com/Dmitrevicz/gometrics/internal/model"
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
		s.Gauges().Set(gauge.name, gauge.value)

		got, ok := s.Gauges().Get(gauge.name)
		require.True(t, ok, "expected ok=true, but nothing was found")
		assert.Equal(t, gauge.value, got)
	})

	t.Run("not found", func(t *testing.T) {
		got, ok := s.Gauges().Get("unknown-test-gauge")
		require.Falsef(t, ok, "expected ok=false, but gauge was found - name: %s, gauge: %d", gauge.name, got)
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
			s.Gauges().Set(c.name, c.value)
		}

		got := s.Gauges().GetAll()
		require.NotEmpty(t, got)
		assert.Len(t, got, len(counters))
	})

	t.Run("get empty", func(t *testing.T) {
		s := New()

		for _, c := range counters {
			s.Gauges().Delete(c.name)
		}

		got := s.Gauges().GetAll()
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

	s.Gauges().Set(gauge.name, gauge.value)

	got, ok := s.Gauges().Get(gauge.name)
	require.True(t, ok, "expected ok=true, but nothing was found")
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

	s.Gauges().Set(gauge.name, gauge.value)

	got, ok := s.Gauges().Get(gauge.name)
	require.True(t, ok, "expected ok=true, but nothing was found")
	assert.Equal(t, gauge.value, got)

	s.Gauges().Delete(gauge.name)

	got, ok = s.Gauges().Get(gauge.name)
	require.Falsef(t, ok, "expected ok=false, but gauge was found - name: %s, gauge: %d", gauge.name, got)
}
