package utils

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSet(t *testing.T) {
	s := NewSet[int]()
	s.Add(1)
	require.True(t, s.Contains(1))
	require.False(t, s.Contains(2))
	s.Remove(1)
	require.False(t, s.Contains(1))
	s.Add(1)
	s.Add(1)
	s.Add(1)
	s.Add(1)
	s.Remove(1)
	require.False(t, s.Contains(1))

}
