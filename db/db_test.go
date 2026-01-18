package db

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLoadBluge(t *testing.T) {
	req := require.New(t)
	bluge, err := LoadBluge(t)
	defer bluge.Close()
	req.NoError(err)
}

func TestLoadBadger(t *testing.T) {
	req := require.New(t)
	badger, err := LoadBadger(t)
	defer badger.Close()
	req.NoError(err)
}
