package db

import (
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/require"
)

func TestLoadBadger_SanityCheck(t *testing.T) {
	req := require.New(t)
	db, err := LoadBadger(t)
	req.NoError(err)
	defer db.Close()

	key := []byte("test-key")
	val := []byte("test-value")

	err = db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
	req.NoError(err)

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		req.NoError(err)
		return item.Value(func(v []byte) error {
			req.Equal(val, v)
			return nil
		})
	})
	req.NoError(err)
}
