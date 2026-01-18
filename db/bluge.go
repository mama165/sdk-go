package db

import (
	"github.com/blugelabs/bluge"
	"testing"
)

func LoadBluge(t *testing.T) (*bluge.Writer, error) {
	return bluge.OpenWriter(bluge.DefaultConfig(t.TempDir()))
}
