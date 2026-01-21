package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	ass := assert.New(t)

	input := map[string]any{
		"email":    "user@example.com",
		"password": "supersecret",
		"profile": map[string]any{
			"token":    "123456",
			"nickname": "cooluser",
			"permissions": []any{
				map[string]any{
					"secret": "topsecret",
					"role":   "admin",
				},
			},
		},
	}

	expected := map[string]any{
		"email":    "user@example.com",
		"password": "*****",
		"profile": map[string]any{
			"token":    "*****",
			"nickname": "cooluser",
			"permissions": []any{
				map[string]any{
					"secret": "*****",
					"role":   "admin",
				},
			},
		},
	}

	Sanitize(input)
	ass.Equal(expected, input)
}
