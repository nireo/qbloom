# qbloom: very fast bloom filter for go

A fast bloom filter that is based on the [fastbloom](https://github.com/tomtomwombat/fastbloom) implementation in Rust. The general speed-up comes from only requiring a single hash per item as compared to multiple in regular implementations. The optimizations are based on [this](https://www.eecs.harvard.edu/~michaelm/postscripts/rsa2008.pdf) paper. In addition, this implementation provides an atomic bloom filter that supports concurrent writes without relying on a mutex as compared to many other bloom filters.
