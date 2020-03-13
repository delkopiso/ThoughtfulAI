package lru

type Cache interface {
	Put(key int, value string)
	Get(key int) string
}
