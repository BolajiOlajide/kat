package database

// NullStringColumn represents a string that should be inserted/updated as NULL when blank.
func NullStringColumn(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
