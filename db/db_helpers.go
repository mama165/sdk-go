package db

import (
	"context"
	"log/slog"

	"github.com/blugelabs/bluge"
	"github.com/dgraph-io/badger/v4"
	"github.com/mama165/sdk-go/logs"
)

// LoadBadger opens a small Badger DB for testing or benchmarks
func LoadBadger(path string) (*badger.DB, error) {
	opts := badger.DefaultOptions(path).
		WithLoggingLevel(badger.ERROR).
		WithValueLogFileSize(16 << 20)
	return badger.Open(opts)
}

// LoadBluge opens a Bluge writer for testing or benchmarks
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

// SetupBenchmark sets up logger, Badger DB and Bluge writer
func SetupBenchmark(path string) (context.Context, *slog.Logger, *badger.DB, *bluge.Writer, error) {
	ctx := context.Background()
	log := logs.GetLoggerFromLevel(slog.LevelError)

	badgerDB, err := LoadBadger(path)
	if err != nil {
		return ctx, nil, nil, nil, err
	}

	blugeWriter, err := LoadBluge(path)
	if err != nil {
		return ctx, nil, nil, nil, err
	}

	return ctx, log, badgerDB, blugeWriter, nil
}
