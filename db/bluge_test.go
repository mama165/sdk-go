package db

import (
	"context"
	"github.com/blugelabs/bluge"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLoadBluge_SanityCheck(t *testing.T) {
	req := require.New(t)
	ctx := context.Background()
	writer, err := LoadBluge(t)
	req.NoError(err)
	req.NotNil(writer)
	defer writer.Close()

	// Given: A very simple document
	doc := bluge.NewDocument("test-id").
		AddField(bluge.NewTextField("content", "hello world").StoreValue())

	// When: Updating the index
	err = writer.Update(doc.ID(), doc)
	req.NoError(err)

	// Then: We try to find it immediately
	reader, err := writer.Reader()
	req.NoError(err)
	defer reader.Close()

	query := bluge.NewMatchQuery("hello").SetField("content")
	request := bluge.NewTopNSearch(1, query)

	dmi, err := reader.Search(ctx, request)
	req.NoError(err)

	match, err := dmi.Next()
	req.NoError(err)

	// If match is nil here, then your LoadBluge config is too slow for unit tests
	req.NotNil(match, "Document should be searchable immediately after update")
}
