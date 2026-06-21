package database

import (
	"strconv"
	"testing"
)

func TestMigrationVersionsAreUniqueAndSortable(t *testing.T) {
	files, err := findUpMigrations()
	if err != nil {
		t.Fatalf("find migrations: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no migrations found")
	}

	seen := make(map[string]string, len(files))
	for i, file := range files {
		if len(file.Version) != 12 {
			t.Errorf("migration %q has version %q; want YYYYMMDDNNNN", file.Name, file.Version)
		}
		if _, err := strconv.ParseUint(file.Version, 10, 64); err != nil {
			t.Errorf("migration %q has non-numeric version %q", file.Name, file.Version)
		}
		if previous, ok := seen[file.Version]; ok {
			t.Errorf("migrations %q and %q share version %q", previous, file.Name, file.Version)
		}
		seen[file.Version] = file.Name

		if i > 0 && files[i-1].Version >= file.Version {
			t.Errorf("migrations are not strictly ordered: %q before %q", files[i-1].Name, file.Name)
		}
	}
}
