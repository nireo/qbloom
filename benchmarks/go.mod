module github.com/nireo/qbloom/benchmarks

go 1.26

require (
	github.com/bits-and-blooms/bloom/v3 v3.7.1
	github.com/nireo/qbloom v0.0.0
)

require (
	github.com/bits-and-blooms/bitset v1.24.2 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

replace github.com/nireo/qbloom => ..
