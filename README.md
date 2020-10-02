---
## Construct fix sized block FM-Index with hybrid compression encode

```go

import "github.com/rleiwang/hfmi/ctor"

var bwt []byte
// construct bwt

index := ctor.New(bwt)
```

The index is a self compressed succinct data structure can be used to locate a string pattern or extract/restore partial or full original text content

```go
// Succinct defines Rank/Select succinct data structure
type Succinct interface {
	// Access returns byte and its rank at p-th position, p is zero based offset
	Access(p uint) (a byte, r uint, ok bool)

	// Select returns the position p of r-th ranked a, p is zero based offset
	Select(a byte, r uint) (p uint, ok bool)

	// Rank returns the rank, r-th of byte at the position p, p is zero based offset
	Rank(a byte, p uint) (r uint, ok bool)

	// Bytes returns bit vector of this succinct data structure
	Bytes() []byte

	//
	Dictionary() []byte
}

type FMI interface {
	Succinct

	// Locate returns the bucket byte and its rank at p-th position, p is zero based offset
	Locate(uint) (byte, uint, bool)

	// Count the number of pattern occurrence
	Count(string) uint

	// Search search pattern, return range in BWT (s, e]
	Search(string) ([]uint, bool)

	// Size return the size of header and body bit vector
	Size() (int, int)

	// Len return original text len
	Len() uint

	// Histogram return [256]uint
	Histogram() []uint

	// CharsInBound return all chars between bound [start, end]
	CharsInBound(uint, uint) []byte

	// GetBound return (start, end] range bound
	GetBound(byte) (uint, uint, bool)

	//
	Restore(io.Writer) bool

	// ForwardExtractToChar return []byte, position, ok, walking BWT forward from p, until found b
	ForwardExtractToChar(uint, byte) ([]byte, uint, bool)

	// BackwardExtractToChar return []byte, position, ok, walking BWT backward from p, until found b
	BackwardExtractToChar(uint, byte) ([]byte, uint, bool)

	// BackwardJumpToChar return position, ok, walking BWT backward from p, until found b
	BackwardJumpToChar(uint, byte) (uint, bool)

	// ExtractFields, returns all fields where p falls in.
	// sep -> separator, p -> position of BWT, fc -> field count
	ExtractFields(sep byte, p uint, fc uint) ([][]byte, bool)

	ExtractAllFields(sep byte, fc uint) ([][][]byte, bool)

	// ExtractRange, return []byte between from and to or 0/1 byte, which ever comes first
	ExtractRange(from, to uint) ([]byte, bool)
}
```

---
## References

* [Hybrid Compression of Bitvectors for the FM-Index](https://ieeexplore.ieee.org/abstract/document/6824438) Juha Kärkkäinen, Dominik Kempa, Simon J. Puglisi
* [Fixed Block Compression Boosting in FM-Indexes: Theory and Practice](https://link.springer.com/article/10.1007/s00453-018-0475-9) Simon Gog, Juha Kärkkäinen, Dominik Kempa, Matthias Petri & Simon J. Puglisi
* [Compact data structures: A practical approach](https://www.google.com/search?tbm=bks&hl=en&q=compact+data+structures) Gonzalo Navarro
