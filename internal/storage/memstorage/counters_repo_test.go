package memstorage

import (
	"testing"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/storage"
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
		err := s.Counters().Set(counter.name, counter.value)
		require.NoError(t, err)

		got, err := s.Counters().Get(counter.name)
		require.NotErrorIs(t, err, storage.ErrNotFound, "unexpected ErrNotFound when error must be nil")
		require.NoError(t, err)
		assert.Equal(t, counter.value, got)
	})

	t.Run("not found", func(t *testing.T) {
		got, err := s.Counters().Get("unknown-test-counter")
		require.Errorf(t, err, "expected nothing (ErrNotFound), but found something - name: %s, counter: %d", counter.name, got)
		require.ErrorIs(t, err, storage.ErrNotFound, "expected ErrNotFound")
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
			err := s.Counters().Set(c.name, c.value)
			require.NoError(t, err)
		}

		got, err := s.Counters().GetAll()
		require.NoError(t, err)
		require.NotEmpty(t, got)
		assert.Len(t, got, len(counters))
	})

	t.Run("get empty", func(t *testing.T) {
		s := New()

		for _, c := range counters {
			err := s.Counters().Delete(c.name)
			require.NoError(t, err)
		}

		got, err := s.Counters().GetAll()
		require.NoError(t, err)
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

	err := s.Counters().Set(counter.name, counter.value)
	require.NoError(t, err)

	got, err := s.Counters().Get(counter.name)
	require.NotErrorIs(t, err, storage.ErrNotFound, "ErrNotFound: nothing found after metric update attempt")
	require.NoError(t, err)
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

	err := s.Counters().Set(counter.name, counter.value)
	require.NoError(t, err)

	got, err := s.Counters().Get(counter.name)
	require.NotErrorIs(t, err, storage.ErrNotFound, "ErrNotFound: nothing found after metric update attempt")
	require.NoError(t, err)
	assert.Equal(t, counter.value, got)

	err = s.Counters().Delete(counter.name)
	require.NoError(t, err)

	got, err = s.Counters().Get(counter.name)
	require.Errorf(t, err, "expected nothing (ErrNotFound), but found something - name: %s, counter: %d", counter.name, got)
	require.ErrorIs(t, err, storage.ErrNotFound, "expected ErrNotFound")
}
