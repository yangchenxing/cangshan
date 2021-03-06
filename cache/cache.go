package cache

type Cache interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, bool, error)
	Remove(key string) error
}
