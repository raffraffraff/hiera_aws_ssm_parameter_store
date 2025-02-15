package main

type Cache struct {
	data map[string]string
}

func NewCache() *Cache {
	return &Cache{data: make(map[string]string)}
}

func (c *Cache) Put(key string, value string) {
	c.data[key] = value
}

func (c *Cache) Get(key string) (string, bool) {
	val, exists := c.data[key]
	return val, exists
}

var cache *Cache

func init() {
	cache = NewCache()
}
