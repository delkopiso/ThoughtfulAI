package lru

import (
	"better.com/list"
)

type DoublyLinkedListBacked struct {
	capacity   int
	values     map[int]string
	linkedList *list.DoublyLinkedList
}

func NewDoublyLinkedListBacked(capacity int) *DoublyLinkedListBacked {
	return &DoublyLinkedListBacked{
		capacity:   capacity,
		values:     make(map[int]string, capacity),
		linkedList: list.NewDoublyLinkedList(),
	}
}

func (c *DoublyLinkedListBacked) Put(key int, value string) {
	c.increment(key)
	c.values[key] = value
}

func (c *DoublyLinkedListBacked) Get(key int) string {
	if v, exists := c.values[key]; exists {
		c.increment(key)
		return v
	}
	return "-1"
}

func (c *DoublyLinkedListBacked) increment(key int) {
	c.linkedList.Push(key)
	if c.linkedList.Count() > c.capacity {
		e := c.linkedList.Pop()
		delete(c.values, e)
	}
}
