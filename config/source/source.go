package source

import "time"

// KeyValue is config key value.
type KeyValue struct {
	Key       string
	Value     []byte
	Metadata  map[string]string
	Timestamp time.Time
}

// Source is config source.
type Source interface {
	Load() ([]*KeyValue, error)
	Watch() (Watcher, error)
}

// Watcher watches a source for changes.
type Watcher interface {
	Next() ([]*KeyValue, error)
	Close() error
}
