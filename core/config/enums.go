package config

// memory, redis
type CacheType string

const (
	Cache_Memory CacheType = "memory"
	Cache_Redis  CacheType = "redis"
)

// file, mongo
type StorageType string

const (
	Storage_File  StorageType = "file"
	Storage_Mongo StorageType = "mongo"
)
