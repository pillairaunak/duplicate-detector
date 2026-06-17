package generator

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerate(t *testing.T) {
	// t.TempDir() creates a temporary directory that is automatically cleaned up
	tempDir := t.TempDir()

	t.Run("Generates correct number of lines and format", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "format_test.txt")
		total := 5000

		err := Generate(filePath, total, 0.5)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		file, err := os.Open(filePath)
		if err != nil {
			t.Fatalf("Failed to open generated file: %v", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineCount := 0
		for scanner.Scan() {
			line := scanner.Text()
			lineCount++

			// Invariant: Must be exactly 10 characters
			if len(line) != 10 {
				t.Errorf("Expected 10 digit number, got %d digits: %s", len(line), line)
			}
		}

		if err := scanner.Err(); err != nil {
			t.Fatalf("Error reading file: %v", err)
		}

		if lineCount != total {
			t.Errorf("Expected %d lines, got %d", total, lineCount)
		}
	})

	t.Run("Zero duplicate ratio generates strictly unique numbers", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "unique_test.txt")
		total := 10000

		err := Generate(filePath, total, 0.0)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		uniqueNumbers := countUniqueLines(t, filePath)
		if uniqueNumbers != total {
			t.Errorf("Expected %d unique numbers (0.0 ratio), got %d", total, uniqueNumbers)
		}
	})

	t.Run("100% duplicate ratio generates exactly one unique number", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "duplicate_test.txt")
		total := 5000

		err := Generate(filePath, total, 1.0)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		uniqueNumbers := countUniqueLines(t, filePath)
		if uniqueNumbers != 1 {
			t.Errorf("Expected exactly 1 unique number (1.0 ratio), got %d", uniqueNumbers)
		}
	})

	t.Run("Fails gracefully on bad file path", func(t *testing.T) {
		// Attempting to write to a directory path instead of a file should error
		err := Generate(tempDir, 100, 0.0)
		if err == nil {
			t.Error("Expected error when providing a directory instead of a file path, got nil")
		}
	})
}

// Helper function to read a file and count unique lines
func countUniqueLines(t *testing.T, filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file in helper: %v", err)
	}
	defer file.Close()

	seen := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		seen[scanner.Text()] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error scanning file in helper: %v", err)
	}

	return len(seen)
}
