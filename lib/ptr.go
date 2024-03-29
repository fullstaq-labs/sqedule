package lib

// NewBoolPtr allows creating a bool pointer in 1 line.
//
// Before:
//
//   var val = true
//   doSomething(&val)
//
// After:
//
//   doSomething(lib.NewBoolPtr(1))
func NewBoolPtr(val bool) *bool {
	return &val
}

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

// NewStringPtr allows creating a string pointer in 1 line.
//
// Before:
//
//   var data = "hello"
//   doSomething(&data)
//
// After:
//
//   doSomething(lib.NewStringPtr("hello"))
func NewStringPtr(val string) *string {
	return &val
}

// DerefBoolPtrWithDefault dereferences the given pointer if it's non-nil, or returns the default
// value if it is nil.
//
// It shortens code. Before:
//
//   if val == nil {
//     doSomething(*val)
//   } else {
//     doSomething(true)
//   }
//
// After:
//
//   doSomething(DerefBoolPtrWithDefault(val, true))
func DerefBoolPtrWithDefault(ptr *bool, defaultValue bool) bool {
	if ptr == nil {
		return defaultValue
	} else {
		return *ptr
	}
}

// CopyBoolPtr copies the value if the given boolean pointer into a new boolean pointer.
// If the pointer is nil, then it returns nil.
func CopyBoolPtr(ptr *bool) *bool {
	if ptr == nil {
		return nil
	} else {
		return NewBoolPtr(*ptr)
	}
}
