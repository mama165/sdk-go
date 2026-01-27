package database

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
)

//go:embed inspect.html
var templatesFS embed.FS

var (
	resumeChan  = make(chan struct{}, 1)
	currentPort int
)

// InspectRow represents the generic columns for the UI.
// The SDK doesn't know about the schema of the stored data
// Only these display strings.
type InspectRow struct {
	Key       string
	Type      string
	Timestamp string
	EntityID  string
	Namespace string
	Detail    string
	Scores    string
}

// RowMapper is a function type that translates raw Badger data into an InspectRow.
type RowMapper func(key string, val []byte) InspectRow

// PageData is the container for template rendering.
type PageData struct {
	Prefix string
	Items  []InspectRow
}

// Inspect wraps the debug server lifecycle and blocks execution with Wait.
func Inspect(db *badger.DB, port int, endpoint string, mapper RowMapper, prefix string, fn func()) {
	// Start the server (non-blocking)
	StartDebugServer(db, port, endpoint, mapper)

	// Execute the user's code (e.g., storing data)
	if fn != nil {
		fn()
	}

	// Automatically wait for user interaction
	Wait(prefix)
}

// StartDebugServer starts the HTTP server.
// Pass a custom RowMapper to decode your specific data (e.g., Protobuf).
func StartDebugServer(db *badger.DB, port int, endpoint string, mapper RowMapper) {
	currentPort = port
	mux := http.NewServeMux()

	// Load the embedded template
	tmpl := template.Must(template.ParseFS(templatesFS, "inspect.html"))

	// Use DefaultMapper if none provided
	if mapper == nil {
		mapper = DefaultMapper
	}

	mux.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		prefix := r.URL.Query().Get("prefix")
		if prefix == "" {
			prefix = "analysis:"
		}

		data := PageData{Prefix: prefix}

		_ = db.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()

			for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
				item := it.Item()
				key := string(item.Key())
				_ = item.Value(func(val []byte) error {
					data.Items = append(data.Items, mapper(key, val))
					return nil
				})
			}
			return nil
		})

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.Execute(w, data)
	})

	mux.HandleFunc("/resume", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-resumeChan:
		default:
		}
		resumeChan <- struct{}{}
		fmt.Fprint(w, "RESUMED")
	})

	go func() {
		_ = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), mux)
	}()
}

// Wait blocks the test and prints the inspection link with the given prefix.
func Wait(prefix string) {
	url := fmt.Sprintf("http://localhost:%d/inspect?prefix=%s", currentPort, prefix)
	fmt.Printf("\n--- TEST PAUSED ---\n\n%s\n\n-------------------\n", url)
	<-resumeChan
}

// DefaultMapper provides basic key-based metadata extraction.
// Exported so you can use it as a base in your custom mappers.
func DefaultMapper(key string, val []byte) InspectRow {
	parts := strings.Split(key, ":")

	row := InspectRow{
		Key:       key,
		Type:      "RAW",
		Timestamp: "--:--:--",
		EntityID:  "--------",
		Namespace: "default",
		Detail:    "Size: " + strconv.Itoa(len(val)) + " bytes",
		Scores:    "-",
	}

	if len(parts) >= 4 {
		row.Namespace = parts[1]

		if tsNano, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
			row.Timestamp = time.Unix(0, tsNano).Format("15:04:05")
		}

		row.EntityID = parts[3]
		if len(row.EntityID) > 8 {
			row.EntityID = row.EntityID[:8]
		}
	}

	return row
}
