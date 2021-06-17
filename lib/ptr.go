package lib

// NewUint32Ptr allows creating a uint32 pointer in 1 line.
//
// Before:
//
//   var versionNumber uint32 = 1
//   doSomething(&versionNumber)
//
// After:
//
//   doSomething(lib.NewUint32Ptr(1))
func NewUint32Ptr(val uint32) *uint32 {
	return &val
}
