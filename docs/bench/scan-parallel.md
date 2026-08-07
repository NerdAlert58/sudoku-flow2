# Scan-parallel vs sequential solve — measured result (UC-5)

Benchmark of `solver.SolveScanParallel` (intra-puzzle scan-parallel variant,
ADR-0015) against sequential `solver.Solve`. Each benchmark op solves all 10
VERY-HARD corpus seeds.

## Command

```
go test -bench 'BenchmarkSolve(Sequential|ScanParallel)' -run '^$' -benchmem -count=5 ./solver/
```

## Environment

- go1.26.5 darwin/arm64
- Apple M4 Max, 14 cores, GOMAXPROCS=14
- macOS 26.5.1

## Raw output

```
goos: darwin
goarch: arm64
pkg: github.com/NerdAlert58/sudoku-flow2/solver
cpu: Apple M4 Max
BenchmarkSolveSequential-14      	     504	   2156967 ns/op	  458948 B/op	    7415 allocs/op
BenchmarkSolveSequential-14      	     562	   2093765 ns/op	  458936 B/op	    7415 allocs/op
BenchmarkSolveSequential-14      	     556	   2110687 ns/op	  458937 B/op	    7415 allocs/op
BenchmarkSolveSequential-14      	     578	   2129362 ns/op	  458946 B/op	    7415 allocs/op
BenchmarkSolveSequential-14      	     577	   2122220 ns/op	  458937 B/op	    7415 allocs/op
BenchmarkSolveScanParallel-14    	      52	  22887445 ns/op	 7468862 B/op	  132875 allocs/op
BenchmarkSolveScanParallel-14    	      52	  22934217 ns/op	 7463232 B/op	  132863 allocs/op
BenchmarkSolveScanParallel-14    	      51	  22740388 ns/op	 7462948 B/op	  132863 allocs/op
BenchmarkSolveScanParallel-14    	      51	  22919574 ns/op	 7463086 B/op	  132864 allocs/op
BenchmarkSolveScanParallel-14    	      52	  22762389 ns/op	 7462920 B/op	  132863 allocs/op
PASS
ok  	github.com/NerdAlert58/sudoku-flow2/solver	12.043s
```

## Conclusion

Negative result, as the PRD predicted (UC-5). Sequential wins by ~10.8x:
~2.12 ms vs ~22.85 ms per 10-puzzle batch (~212 µs vs ~2.29 ms per puzzle).
The scan-parallel variant also allocates ~16x more memory (7.46 MB vs 459 KB
per op) across ~18x more allocations. A 9x9 pass fires on the ladder's cheap
low rungs almost every time, so each pass does microseconds of useful work —
and the variant pays for 13 goroutine launches, 13 state snapshots, and a
WaitGroup join per pass to get it. Goroutine overhead dominates sub-millisecond
solves; the sequential solver remains the only serving-path implementation
(containment guard: `TestSolveScanParallel_NoReferenceOutsideSolverPackage`).
