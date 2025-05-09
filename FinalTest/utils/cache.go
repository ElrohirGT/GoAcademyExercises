package utils

import "time"

type Cache[T any] struct {
	Data       []T
	Length     uint
	ExpireTime time.Time
}

func (c *Cache[T]) IsValid() bool {
	return time.Now().After(c.ExpireTime)
}
