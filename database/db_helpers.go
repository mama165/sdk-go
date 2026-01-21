package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/blugelabs/bluge"
	"github.com/dgraph-io/badger/v4"
	"github.com/mama165/sdk-go/logs"
)

const DefaultPath = "/tmp/database/debug"

func LoadBadger(path string) (*badger.DB, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create badger directory: %w", err)
	}
	opts := badger.DefaultOptions(path).WithLoggingLevel(badger.ERROR).WithValueLogFileSize(16 << 20).WithCompactL0OnClose(true)
	return badger.Open(opts)
}

func LoadBluge(path string) (*bluge.Writer, error) {
	return bluge.OpenWriter(bluge.DefaultConfig(path))
}

func CleanupDB(badgerDB *badger.DB, blugeWriter *bluge.Writer) {
	if badgerDB != nil {
		badgerDB.Close()
	}
	if blugeWriter != nil {
		blugeWriter.Close()
	}
}

func SetupBenchmark(path string) (context.Context, *slog.Logger, *badger.DB, *bluge.Writer, error) {
	ctx := context.Background()
	log := logs.GetLoggerFromLevel(slog.LevelError)
	if err := os.RemoveAll(path); err != nil {
		return ctx, nil, nil, nil, err
	}
	badgerDB, err := LoadBadger(path)
	if err != nil {
		return ctx, nil, nil, nil, err
	}
	blugeWriter, err := LoadBluge(path)
	if err != nil {
		badgerDB.Close()
		return ctx, nil, nil, nil, err
	}
	return ctx, log, badgerDB, blugeWriter, nil
}
