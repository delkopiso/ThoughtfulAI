package list_test

import (
	"testing"

	"better.com/list"
	"github.com/stretchr/testify/require"
)

func TestDoublyLinkedList(t *testing.T) {
	l := list.NewDoublyLinkedList()
	require.Equal(t, 0, l.Count())

	l.Push(1)
	require.Equal(t, 1, l.Count())

	l.Push(2)
	require.Equal(t, 2, l.Count())

	l.Push(1)
	require.Equal(t, 3, l.Count())

	i := l.Pop()
	require.Equal(t, 2, l.Count())
	require.Equal(t, 1, i)

	l.Push(3)
	require.Equal(t, 3, l.Count())

	i = l.Pop()
	require.Equal(t, 2, l.Count())
	require.Equal(t, 2, i)

	l.Push(4)
	require.Equal(t, 3, l.Count())

	i = l.Pop()
	require.Equal(t, 2, l.Count())
	require.Equal(t, 1, i)

	l.Push(3)
	require.Equal(t, 3, l.Count())

	i = l.Pop()
	require.Equal(t, 2, l.Count())
	require.Equal(t, 3, i)

	l.Push(4)
	require.Equal(t, 3, l.Count())

	i = l.Pop()
	require.Equal(t, 2, l.Count())
	require.Equal(t, 4, i)
}
