package repository

import "testing"

func TestNormalizeDriverSortBy(t *testing.T) {
	tests := map[string]string{
		"id":           "d.id",
		"surname":      "d.surname",
		"name":         "d.name",
		"iin":          "d.iin",
		"board_number": "COALESCE((SELECT MIN(v_sort.board_number) FROM vehicles v_sort WHERE d.id = ANY(v_sort.drivers_ids)), '')",
		"status":       "ds.code",
		"created_at":   "d.created_at",
		"updated_at":   "d.updated_at",
		"bad":          "d.surname",
		"":             "d.surname",
	}

	for input, want := range tests {
		if got := normalizeDriverSortBy(input); got != want {
			t.Fatalf("normalizeDriverSortBy(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNormalizeDriverListOrder(t *testing.T) {
	tests := map[string]string{
		"asc":  "ASC",
		"ASC":  "ASC",
		"desc": "DESC",
		"bad":  "ASC",
		"":     "ASC",
	}

	for input, want := range tests {
		if got := normalizeDriverListOrder(input); got != want {
			t.Fatalf("normalizeDriverListOrder(%q) = %q, want %q", input, got, want)
		}
	}
}
