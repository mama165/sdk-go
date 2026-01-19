package db

import (
	"testing"

	"github.com/blugelabs/bluge"
)

func LoadBluge(t *testing.T) (*bluge.Writer, error) {
	return bluge.OpenWriter(bluge.DefaultConfig(t.TempDir()))
}
