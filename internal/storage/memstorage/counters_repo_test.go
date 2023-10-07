package memstorage

import (
	"testing"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountersRepo_Get(t *testing.T) {
	s := New()

	counter := struct {
		name  string
		value model.Counter
	}{
		name:  "TestCounter",
		value: model.Counter(42),
	}

	t.Run("found", func(t *testing.T) {
		s.Counters().Set(counter.name, counter.value)

		got, ok := s.Counters().Get(counter.name)
		require.True(t, ok, "expected ok=true, but nothing was found")
		assert.Equal(t, counter.value, got)
	})

	t.Run("not found", func(t *testing.T) {
		got, ok := s.Counters().Get("unknown-test-counter")
		require.Falsef(t, ok, "expected ok=false, but counter was found - name: %s, counter: %d", counter.name, got)
		assert.EqualValues(t, 0, got)
	})
}

func TestCountersRepo_GetAll(t *testing.T) {
	counters := []struct {
		name  string
		value model.Counter
	}{
		{
			name:  "TestCounter1",
			value: model.Counter(42),
		},
		{
			name:  "TestCounter2",
			value: model.Counter(42),
		},
	}

	t.Run("get", func(t *testing.T) {
		s := New()

		for _, c := range counters {
			s.Counters().Set(c.name, c.value)
		}

		got := s.Counters().GetAll()
		require.NotEmpty(t, got)
		assert.Len(t, got, len(counters))
	})

	t.Run("get empty", func(t *testing.T) {
		s := New()

		for _, c := range counters {
			s.Counters().Delete(c.name)
		}

		got := s.Counters().GetAll()
		assert.Empty(t, got)
		assert.Len(t, got, 0)
	})
}

func TestCountersRepo_Set(t *testing.T) {
	s := New()

	counter := struct {
		name  string
		value model.Counter
	}{
		name:  "TestCounter",
		value: model.Counter(42),
	}

	s.Counters().Set(counter.name, counter.value)

	got, ok := s.Counters().Get(counter.name)
	require.True(t, ok, "expected ok=true, but nothing was found")
	assert.Equal(t, counter.value, got)
}

func TestCountersRepo_Delete(t *testing.T) {
	s := New()

	counter := struct {
		name  string
		value model.Counter
	}{
		name:  "TestCounter",
		value: model.Counter(42),
	}

	s.Counters().Set(counter.name, counter.value)

	got, ok := s.Counters().Get(counter.name)
	require.True(t, ok, "expected ok=true, but nothing was found")
	assert.Equal(t, counter.value, got)

	s.Counters().Delete(counter.name)

	got, ok = s.Counters().Get(counter.name)
	require.Falsef(t, ok, "expected ok=false, but counter was found - name: %s, counter: %d", counter.name, got)
}
