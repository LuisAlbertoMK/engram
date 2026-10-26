package store

import "testing"

// A missing or unreadable observation must surface as an error so the semantic
// scan skips the pair instead of building a prompt from an empty snippet.
func TestScanObservationSnippetReturnsLookupError(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.scanObservationSnippet("missing-sync-id", 0); err == nil {
		t.Fatal("missing observation returned an empty snippet instead of an error")
	}

	if _, err := s.db.Exec(`ALTER TABLE observations RENAME TO observations_snippet_probe`); err != nil {
		t.Skipf("cannot simulate lookup failure via table rename: %v", err)
	}
	t.Cleanup(func() {
		_, _ = s.db.Exec(`ALTER TABLE observations_snippet_probe RENAME TO observations`)
	})
	if _, err := s.scanObservationSnippet("any-sync-id", 0); err == nil {
		t.Fatal("unreadable observation returned an empty snippet instead of an error")
	}
}
