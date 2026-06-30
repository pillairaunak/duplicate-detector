package realm1

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// FindDuplicates reads a massive file, sorts it in bounded memory chunks,
// and writes deduplicated (or only duplicate) entries to outputPath.
func FindDuplicates(inputPath, outputPath string, maxChunkSizeBytes int) error {
	// 1. Create a secure temporary directory for our runs
	// We use the OS's temp directory so we don't clutter the project folder.
	tempDir, err := os.MkdirTemp("", "duplicate-detector-runs-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	// STAFF INSIGHT: Always defer cleanup immediately after creation.
	// If the system panics or errors out, we don't leak gigabytes of temp files.
	defer os.RemoveAll(tempDir)

	// 2. Phase 1: Chunk and Sort
	runFiles, err := chunkAndSort(inputPath, tempDir, maxChunkSizeBytes)
	if err != nil {
		return fmt.Errorf("chunk and sort phase failed: %w", err)
	}

	// 3. Phase 2: K-Way Merge (Stubbed for now)

	return kWayMerge(runFiles, outputPath)
}

// chunkAndSort reads the input file, creating sorted run files bounded by maxBytes.
// It returns a slice of file paths to those runs.
func chunkAndSort(inputPath, tempDir string, maxBytes int) ([]string, error) {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()

	var runs []string
	scanner := bufio.NewScanner(inFile)

	// We pre-allocate a reasonable capacity to avoid early array resizing overhead
	currentChunk := make([]string, 0, 10000)
	currentBytes := 0
	runIndex := 0

	for scanner.Scan() {
		line := scanner.Text()
		currentChunk = append(currentChunk, line)

		// Calculate memory footprint (length of string + 1 byte for newline)
		currentBytes += len(line) + 1

		// If we hit our memory ceiling, flush to disk
		if currentBytes >= maxBytes {
			runPath, err := writeRun(currentChunk, tempDir, runIndex)
			if err != nil {
				return nil, err
			}
			runs = append(runs, runPath)

			// STAFF INSIGHT: Memory Reuse Trick
			// By slicing to zero, we keep the underlying array capacity.
			// The garbage collector doesn't have to clean up the old array,
			// and we don't have to allocate a new one.
			currentChunk = currentChunk[:0]
			currentBytes = 0
			runIndex++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input file: %w", err)
	}

	// Flush any remaining data in the final chunk
	if len(currentChunk) > 0 {
		runPath, err := writeRun(currentChunk, tempDir, runIndex)
		if err != nil {
			return nil, err
		}
		runs = append(runs, runPath)
	}

	return runs, nil
}

// writeRun sorts a chunk in memory and writes it to a temporary file.
func writeRun(lines []string, tempDir string, index int) (string, error) {
	// Sort the slice in-place (O(N log N) using Introsort under the hood)
	sort.Strings(lines)

	path := filepath.Join(tempDir, fmt.Sprintf("run_%d.txt", index))
	outFile, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	// Use a buffer for fast disk I/O, just like in our generator
	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return "", err
		}
	}

	return path, nil
}

/* kWayMerge is a placeholder for our next step.
func kWayMerge(runFiles []string, outputPath string) error {
	fmt.Printf("Ready to merge %d files!\n", len(runFiles))
	return nil
}*/

func kWayMerge(runFiles []string, outputPath string) error {
	// 1. Prepare the output file with a large buffer
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	writer := bufio.NewWriterSize(outFile, 64*1024)
	defer writer.Flush()

	// 2. Initialize the Heap
	h := &ScannerHeap{}
	heap.Init(h)

	// Keep track of files so we can close them safely
	var openFiles []*os.File
	defer func() {
		for _, f := range openFiles {
			f.Close()
		}
	}()

	// 3. Seed the heap with the first element from each run
	for _, runPath := range runFiles {
		file, err := os.Open(runPath)
		if err != nil {
			return fmt.Errorf("failed to open run file %s: %w", runPath, err)
		}
		openFiles = append(openFiles, file)

		rs := &RunScanner{
			file: file,
			// Using a larger scanner buffer to minimize system calls
			scanner: bufio.NewScanner(file),
		}

		// If the file isn't empty, push its first line onto the heap
		if rs.Next() {
			heap.Push(h, rs)
		}
	}

	// 4. The Streaming Merge & Duplicate Detection
	var prev string
	//isFirstLoop := true

	for h.Len() > 0 {
		// Pop the run that currently has the smallest string
		minScanner := heap.Pop(h).(*RunScanner)
		curr := minScanner.current

		// DUPLICATE DETECTION LOGIC
		// If current equals previous, we found a duplicate!
		/*if !isFirstLoop && curr == prev {
			if _, err := writer.WriteString(curr + "\n"); err != nil {
				return fmt.Errorf("failed writing to output: %w", err)
			}
		}*/

		if curr != prev {
			if _, err := writer.WriteString(curr + "\n"); err != nil {
				return fmt.Errorf("failed writing to output: %w", err)
			}
		}

		prev = curr
		//isFirstLoop = false

		// Advance the scanner. If it has more data, push it back onto the heap.
		if minScanner.Next() {
			heap.Push(h, minScanner)
		}
	}

	return nil
}
