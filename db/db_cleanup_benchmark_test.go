package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupBenchmark(t *testing.T) {
	req := require.New(t)
	tmpDir := t.TempDir()

	ctx, log, badgerDB, blugeWriter, err := SetupBenchmark(tmpDir)
	req.NoError(err)

	defer CleanupDB(badgerDB, blugeWriter)

	req.NotNil(ctx)
	req.NotNil(log)
	req.NotNil(badgerDB)
	req.NotNil(blugeWriter)
}

func TestLoadBadger_Error(t *testing.T) {
	req := require.New(t)

	// Given creating a temporary file to simulate an invalid directory path
	tmpFile, err := os.CreateTemp("", "badger_invalid_path")
	req.NoError(err)
	defer os.Remove(tmpFile.Name())

	// Attempting to open Badger on a file path instead of a directory should fail
	_, err = LoadBadger(tmpFile.Name())

	// Verify that an error is correctly returned
	req.Error(err)
}

func TestCleanupDB_GracefulNil(t *testing.T) {
	// Ensure the Cleanup function handles nil pointers without panicking
	// This is critical if Setup fails before all DBs are opened
	require.NotPanics(t, func() {
		CleanupDB(nil, nil)
	})
}
