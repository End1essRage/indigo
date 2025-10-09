package cache

type Cache interface {
	GetString(key string) string
	SetString(key string, val string) error
	Exists(key string) bool
}
