package db

import (
	"testing"

	"github.com/dgraph-io/badger/v4"
)

// LoadBadger Reduced to 16 Mo for testing (avoid 20 Go of storage)
func LoadBadger(t *testing.T) (*badger.DB, error) {
	return badger.Open(badger.DefaultOptions(
		t.TempDir()).
		WithLoggingLevel(badger.ERROR).
		WithValueLogFileSize(16 << 20),
	)
}
