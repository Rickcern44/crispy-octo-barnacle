// Package migrations exposes embedded database migrations used by Cassor.
package migrations

import (
	"embed"
	"io/fs"
	"sort"
)

// Files contains the ordered SQL migration files. The database layer applies
// them transactionally and records which versions have been installed.
//
//go:embed sql/*.sql
var Files embed.FS

// Names returns embedded migration paths in deterministic order.
func Names() ([]string, error) {
	entries, err := fs.ReadDir(Files, "sql")
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, "sql/"+entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
