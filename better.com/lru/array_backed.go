package lru

type ArrayBacked struct {
	capacity int
	values   map[int]string
	stack    []int
}

func NewArrayBacked(capacity int) *ArrayBacked {
	return &ArrayBacked{
		capacity: capacity,
		values:   make(map[int]string, capacity),
		stack:    make([]int, 0, capacity),
	}
}

func (c *ArrayBacked) Put(key int, value string) {
	c.increment(key)
	c.values[key] = value
}

func (c *ArrayBacked) Get(key int) string {
	if v, exists := c.values[key]; exists {
		c.increment(key)
		return v
	}
	return "-1"
}

func (c *ArrayBacked) increment(key int) {
	c.stack = append(c.stack, key)
	if len(c.stack) > c.capacity {
		e := c.stack[0]
		c.stack = c.stack[1:]
		delete(c.values, e)
	}
}
