package realm1

import (
	"bufio"
	"os"
)

// RunScanner wraps a file and its scanner, holding the current line in memory.
// This acts as a node in our Min-Heap.
type RunScanner struct {
	file    *os.File
	scanner *bufio.Scanner
	current string
}

// Next advances the scanner to the next line.
// Returns false if we hit EOF or an error occurs.
func (rs *RunScanner) Next() bool {
	if rs.scanner.Scan() {
		rs.current = rs.scanner.Text()
		return true
	}
	return false
}

// ScannerHeap implements container/heap for our K-Way merge.
// It is a slice of pointers to RunScanners.
type ScannerHeap []*RunScanner

// --- container/heap Interface Implementation ---

func (h ScannerHeap) Len() int { return len(h) }

// Less ensures this is a MIN-Heap. We compare lexicographically.
func (h ScannerHeap) Less(i, j int) bool {
	return h[i].current < h[j].current
}

func (h ScannerHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *ScannerHeap) Push(x any) {
	*h = append(*h, x.(*RunScanner))
}

func (h *ScannerHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	// Avoid memory leaks by nil-ing out the pointer
	old[n-1] = nil
	*h = old[0 : n-1]
	return item
}
