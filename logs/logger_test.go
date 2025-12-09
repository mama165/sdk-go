package logs

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLevelFromString(t *testing.T) {
	ass := assert.New(t)
	level := GetLevelFromString("DEBUG")
	ass.Equal(slog.LevelDebug, level)
}

func TestGetDefaultLevelFromString(t *testing.T) {
	ass := assert.New(t)
	level := GetLevelFromString("Unknown")
	ass.Equal(slog.LevelInfo, level)
}
