package lru_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"interviews/better.com/lru"
)

func TestCache(t *testing.T) {
	capacity := 2
	tests := []struct {
		name  string
		cache lru.Cache
	}{
		{
			name:  "array backed",
			cache: lru.NewArrayBacked(capacity),
		},
		{
			name:  "doubly linked list backed",
			cache: lru.NewDoublyLinkedListBacked(capacity),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cache.Put(1, "one")
			tt.cache.Put(2, "two")
			require.Equal(t, "one", tt.cache.Get(1))

			tt.cache.Put(3, "three")
			require.Equal(t, "-1", tt.cache.Get(2))

			tt.cache.Put(4, "four")
			require.Equal(t, "-1", tt.cache.Get(1))

			require.Equal(t, "three", tt.cache.Get(3))

			require.Equal(t, "four", tt.cache.Get(4))
		})
	}
}
