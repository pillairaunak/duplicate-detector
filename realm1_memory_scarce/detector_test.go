package realm1

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"

	// Replace with your actual module name from go.mod!
	"github.com/pillairaunak/duplicate-detector/internal/generator"
)

func TestExternalSortDuplicateDetection(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Successfully deduplicates while enforcing tight memory constraints", func(t *testing.T) {
		inputPath := filepath.Join(tempDir, "input_standard.txt")
		outputPath := filepath.Join(tempDir, "output_standard.txt")

		totalRecords := 10_000 // ~110 KB of data
		duplicateRatio := 0.30 // ~30% duplicates

		// 1. Generate the test data
		err := generator.Generate(inputPath, totalRecords, duplicateRatio)
		if err != nil {
			t.Fatalf("Failed to generate test data: %v", err)
		}

		// 2. Determine the exact baseline of unique records
		// Because the generator is probabilistic, we must calculate the exact target
		expectedUniqueCount := countUniqueBaseline(t, inputPath)

		// 3. Run the algorithm with a 10 KB memory limit.
		// 110 KB data / 10 KB limit = ~11 sorted runs to merge.
		maxMemoryBytes := 10 * 1024
		err = FindDuplicates(inputPath, outputPath, maxMemoryBytes)
		if err != nil {
			t.Fatalf("FindDuplicates failed: %v", err)
		}

		// 4. Verify the output is perfectly sorted and completely unique
		verifySortedAndUnique(t, outputPath, expectedUniqueCount)
	})

	t.Run("Handles 100% duplicate file (zero entropy)", func(t *testing.T) {
		inputPath := filepath.Join(tempDir, "input_all_dupes.txt")
		outputPath := filepath.Join(tempDir, "output_all_dupes.txt")

		totalRecords := 5_000
		err := generator.Generate(inputPath, totalRecords, 1.0)
		if err != nil {
			t.Fatalf("Failed to generate test data: %v", err)
		}

		// Max memory is half the file size, forcing exactly 2 runs
		maxMemoryBytes := (totalRecords * 11) / 2
		err = FindDuplicates(inputPath, outputPath, maxMemoryBytes)
		if err != nil {
			t.Fatalf("FindDuplicates failed: %v", err)
		}

		// A 100% duplicate file should yield exactly 1 unique record
		verifySortedAndUnique(t, outputPath, 1)
	})

	t.Run("Handles 0% duplicate file (all unique)", func(t *testing.T) {
		inputPath := filepath.Join(tempDir, "input_no_dupes.txt")
		outputPath := filepath.Join(tempDir, "output_no_dupes.txt")

		totalRecords := 5_000
		err := generator.Generate(inputPath, totalRecords, 0.0)
		if err != nil {
			t.Fatalf("Failed to generate test data: %v", err)
		}

		maxMemoryBytes := 10 * 1024
		err = FindDuplicates(inputPath, outputPath, maxMemoryBytes)
		if err != nil {
			t.Fatalf("FindDuplicates failed: %v", err)
		}

		verifySortedAndUnique(t, outputPath, totalRecords)
	})
}

// --- Test Helpers ---

// countUniqueBaseline uses an unbounded memory map to establish the ground truth.
func countUniqueBaseline(t *testing.T, filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file for baseline: %v", err)
	}
	defer file.Close()

	seen := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		seen[scanner.Text()] = struct{}{}
	}
	return len(seen)
}

// verifySortedAndUnique ensures the output file has no duplicates,
// is strictly sorted lexicographically, and matches the expected line count.
func verifySortedAndUnique(t *testing.T, filePath string, expectedCount int) {
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var prev string
	count := 0

	for scanner.Scan() {
		curr := scanner.Text()

		if count > 0 {
			// Invariant 1: No duplicates
			if curr == prev {
				t.Fatalf("Test Failed: Found duplicate in output: %s", curr)
			}
			// Invariant 2: Must be strictly sorted
			if curr < prev {
				t.Fatalf("Test Failed: Output is not sorted. %s came after %s", curr, prev)
			}
		}

		prev = curr
		count++
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Scanner error during verification: %v", err)
	}

	// Invariant 3: Total count must match
	if count != expectedCount {
		t.Fatalf("Test Failed: Expected %d unique records, but output had %d", expectedCount, count)
	}
}
