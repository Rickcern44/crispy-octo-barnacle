package migrations

import "testing"

func TestNamesAreAvailableAndOrdered(t *testing.T) {
	t.Parallel()

	names, err := Names()
	if err != nil {
		t.Fatalf("Names() error = %v", err)
	}
	if len(names) == 0 {
		t.Fatal("Names() returned no migrations")
	}
	for index := 1; index < len(names); index++ {
		if names[index-1] >= names[index] {
			t.Fatalf("migration names not ordered: %q then %q", names[index-1], names[index])
		}
	}
}
