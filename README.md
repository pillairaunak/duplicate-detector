# Scale-Driven Duplicate Detector Suite

A portfolio repository demonstrating optimal software architectures for duplicate detection under various extreme physical constraints (Memory, Time, Computations, and Distribution). 

Each "realm" represents a distinct physical bottleneck, implementing zero-dependency, idiomatic Go components validated against dynamic relative-constraint testing.

---

## 🌌 Realm 1: Memory-Scarce Environment (External Merge Sort)

### 🛑 The Problem Statement
Given an input file of size $N$ containing millions of 10-digit mobile numbers (E.164 strings), clean and deduplicate the file such that the output contains a strictly unique, sorted list of records. 

**Physical Constraints:**
* Total Dataset Size ($N$): $\approx 4\text{ GB}$ (or infinite stream).
* Max Available System RAM ($M$): $\approx 1\text{ GB}$ (Cannot load dataset into memory).
* I/O Transfer Block Size ($B$): Parameterized to scale performance.

### 🧠 Theoretical Solution (Disk Access Machine Model)
When $N \gg M$, traditional in-memory Hash Sets trigger an Out-Of-Memory (OOM) kernel panic. This implementation utilizes the **Disk Access Machine (DAM) Model**, where the algorithmic efficiency is measured by Disk I/O block transfers rather than CPU cycles.

1. **Phase 1: Run Generation (Chunk & Sort):** The input stream is divided into memory-bounded chunks of size $M$. Each chunk is sorted in-place using Go's introsort (`sort.Strings`) and flushed sequentially to disk as a temporary "sorted run" file using a `bufio.Writer` to maximize throughput.
2. **Phase 2: $K$-Way Streaming Merge:** A custom Min-Heap (Priority Queue) of size $K$ (where $K = \lfloor M/B \rfloor - 1$) is initialized. Each node caches exactly *one* line from its respective run file. The minimum element is sequentially popped, creating a globally sorted continuous stream.
3. **Phase 3: Stream Divergence Filtering:** Because the output of the $K$-way merge is globally sorted, duplicates are guaranteed to be contiguous. Deduplication is achieved with $\mathcal{O}(1)$ auxiliary space by enforcing a divergence invariant (`curr != prev`) across the stream.

### 🛠️ Key Go Optimizations & Idioms Implemented
* **Slice Recycling (`currentChunk[:0]`):** Instead of re-allocating a slice on every chunk iteration (which triggers expensive Garbage Collection sweeps), slicing to zero resets the logical length while retaining underlying memory capacity for array reuse.
* **Bounded Node Scanning:** Pointers to `bufio.Scanner` objects inside the priority queue only cache the active line string (`RunScanner.current`), ensuring the heap memory never scales with the file lengths.
* **Deterministic Resource Defenses:** Active tracking and clean sweeping of opened OS file descriptors inside deferred blocks to prevent kernel quota exhaustion (`EMFILE`).

### 🧪 Validation & Test Rig
Located in `./realm1_memory_scarce/detector_test.go`. Rather than forcing a developer to download a massive multi-gigabyte fixture, the package implements **Relative Constraint Testing**. 

By scaling the file down to ~110 KB and strictly squeezing `maxChunkSizeBytes` down to 10 KB, the test harness replicates the exact multi-pass, 11-way merge tree logic executed at production scale.# Duplicate Detector
