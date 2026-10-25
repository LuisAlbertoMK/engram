package store

import (
	"database/sql/driver"
	"errors"
	"os"
	"testing"
)

// S1: Next performs exactly one generation probe, after the driver row read.
func TestGenerationRowsNextSingleProbe(t *testing.T) {
	generation, _ := newTestDatabaseGeneration(t, false, false)
	rows := generationRows{Rows: &testFenceRows{}, generation: generation}

	original := statFile
	t.Cleanup(func() { statFile = original })
	statCalls := 0
	statFile = func(path string) (os.FileInfo, error) {
		statCalls++
		return original(path)
	}

	if err := rows.Next(make([]driver.Value, 1)); err != nil {
		t.Fatalf("Next: %v", err)
	}
	if statCalls != 3 {
		t.Fatalf("stat calls = %d, want 3 (one generation probe)", statCalls)
	}
}

// S2: a replacement before or during the row read is still rejected by the
// same Next call.
func TestGenerationRowsNextDetectsReplacement(t *testing.T) {
	for _, name := range []string{"before row read", "during row read"} {
		t.Run(name, func(t *testing.T) {
			generation, dbPath := newTestDatabaseGeneration(t, false, false)
			base := &testFenceRows{}
			if name == "during row read" {
				base.next = func() { replaceTestFile(t, dbPath) }
			}
			if name == "before row read" {
				replaceTestFile(t, dbPath)
			}
			rows := generationRows{Rows: base, generation: generation}
			if err := rows.Next(make([]driver.Value, 1)); !errors.Is(err, ErrDatabaseGenerationChanged) {
				t.Fatalf("err = %v, want ErrDatabaseGenerationChanged", err)
			}
		})
	}
}
