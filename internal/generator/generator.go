package generator

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"
)

// Generate writes totalRecords of 10-digit numbers to filePath.
// duplicateRatio (0.0 to 1.0) dictates the probability of generating a duplicate.
func Generate(filePath string, totalRecords int, duplicateRatio float64) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Use a 64KB buffer instead of default 4KB for even faster large-file I/O
	writer := bufio.NewWriterSize(file, 64*1024)
	// Ensure the last partial block in memory is flushed to disk!
	defer writer.Flush()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	min := int64(1_000_000_000)
	max := int64(9_999_999_999)

	// State machine: keep a pool of already generated numbers to create duplicates
	maxPoolSize := 10_000
	seenPool := make([]int64, 0, maxPoolSize)

	for i := 0; i < totalRecords; i++ {
		var mobileNum int64

		// Decide if we should emit a duplicate (and ensure we have something to duplicate)
		if len(seenPool) > 0 && rng.Float64() < duplicateRatio {
			// Pick a random number we've already generated
			randomIndex := rng.Intn(len(seenPool))
			mobileNum = seenPool[randomIndex]
		} else {
			// Generate a brand new number
			mobileNum = min + rng.Int63n(max-min+1)

			// Manage the seen pool size to avoid memory leaks on massive datasets
			if len(seenPool) < maxPoolSize {
				seenPool = append(seenPool, mobileNum)
			} else if rng.Float32() < 0.1 {
				// 10% chance to overwrite an old item to keep the pool fresh
				seenPool[rng.Intn(maxPoolSize)] = mobileNum
			}
		}

		// Write to the memory buffer
		_, err := fmt.Fprintln(writer, mobileNum)
		if err != nil {
			return fmt.Errorf("error writing record to buffer: %w", err)
		}
	}

	return nil
}
