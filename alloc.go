package gno

import "reflect"

// Keeps track of in-memory allocations.
// In the future, allocations within realm boundaries will be
// (optionally?) condensed (objects to be GC'd will be discarded),
// but for now, allocations strictly increment across the whole tx.
type Allocator struct {
	bytes int64
}

// for gonative, which doesn't consider the allocator.
var nilAllocator = (*Allocator)(nil)

const maxAllocations = 1000000000 // TODO parameterize. 1GB for now.

const (
	allocString      = 1
	allocStringByte  = 1
	allocBigint      = 1
	allocBigintByte  = 1
	allocPointer     = 1
	allocArray       = 1
	allocArrayItem   = 1
	allocSlice       = 1
	allocStruct      = 1
	allocStructField = 1
	allocFunc        = 1
	allocMap         = 1
	allocMapItem     = 1
	allocBoundMethod = 1
	allocBlock       = 1
	allocBlockItem   = 1
	allocNative      = 1
	allocType        = 1
	// allocDataByte    = 1
	// allocPackge = 1
)

func NewAllocator() *Allocator {
	return &Allocator{}
}

func (alloc *Allocator) Allocate(size int64) {
	if alloc == nil {
		// this can happen for map items just prior to assignment.
		return
	}
	alloc.bytes += size
	if alloc.bytes > maxAllocations {
		panic("allocation limit exceeded")
	}
}

func (alloc *Allocator) AllocateString(size int64) {
	alloc.Allocate(allocString + allocStringByte*size)
}

func (alloc *Allocator) AllocatePointer() {
	alloc.Allocate(allocPointer)
}

func (alloc *Allocator) AllocateDataArray(size int64) {
	alloc.Allocate(allocArray + size)
}

func (alloc *Allocator) AllocateListArray(items int64) {
	alloc.Allocate(allocArray + allocArrayItem*items)
}

func (alloc *Allocator) AllocateSlice() {
	alloc.Allocate(allocSlice)
}

// NOTE: fields must be allocated separately.
func (alloc *Allocator) AllocateStruct() {
	alloc.Allocate(allocStruct)
}

func (alloc *Allocator) AllocateStructFields(fields int64) {
	alloc.Allocate(allocStructField * fields)
}

func (alloc *Allocator) AllocateFunc() {
	alloc.Allocate(allocFunc)
}

func (alloc *Allocator) AllocateMap(items int64) {
	alloc.Allocate(allocMap + allocMapItem*items)
}

func (alloc *Allocator) AllocateMapItem() {
	alloc.Allocate(allocMapItem)
}

func (alloc *Allocator) AllocateBoundMethod() {
	alloc.Allocate(allocBoundMethod)
}

func (alloc *Allocator) AllocateBlock(items int64) {
	alloc.Allocate(allocBlock + allocBlockItem*items)
}

func (alloc *Allocator) AllocateBlockItems(items int64) {
	alloc.Allocate(allocBlockItem * items)
}

// NOTE: does not allocate for the underlying value.
func (alloc *Allocator) AllocateNative() {
	alloc.Allocate(allocNative)
}

/* NOTE: Not used, account for with AllocatePointer.
func (alloc *Allocator) AllocateDataByte() {
	alloc.Allocate(allocDataByte)
}
*/

func (alloc *Allocator) AllocateType() {
	alloc.Allocate(allocType)
}

//----------------------------------------
// constructor utilities.

func (alloc *Allocator) NewString(s string) StringValue {
	alloc.AllocateString(int64(len(s)))
	return StringValue(s)
}

func (alloc *Allocator) NewSliceFromList(list []TypedValue) *SliceValue {
	alloc.AllocateSlice()
	alloc.AllocateListArray(int64(cap(list)))
	fullList := list[:cap(list)]
	return &SliceValue{
		Base: &ArrayValue{
			List: fullList,
		},
		Offset: 0,
		Length: len(list),
		Maxcap: cap(list),
	}
}

func (alloc *Allocator) NewSliceFromData(data []byte) *SliceValue {
	alloc.AllocateSlice()
	alloc.AllocateDataArray(int64(cap(data)))
	fullData := data[:cap(data)]
	return &SliceValue{
		Base: &ArrayValue{
			Data: fullData,
		},
		Offset: 0,
		Length: len(data),
		Maxcap: cap(data),
	}
}

// NOTE: fields must be allocated (e.g. from NewStructFields)
func (alloc *Allocator) NewStruct(fields []TypedValue) *StructValue {
	alloc.AllocateStruct()
	return &StructValue{
		Fields: fields,
	}
}

func (alloc *Allocator) NewStructFields(fields int) []TypedValue {
	alloc.AllocateStructFields(int64(fields))
	return make([]TypedValue, fields)
}

// NOTE: fields will be allocated.
func (alloc *Allocator) NewStructWithFields(fields ...TypedValue) *StructValue {
	tvs := alloc.NewStructFields(len(fields))
	copy(tvs, fields)
	return alloc.NewStruct(tvs)
}

func (alloc *Allocator) NewBlock(source BlockNode, parent *Block) *Block {
	alloc.AllocateBlock(int64(source.GetNumNames()))
	return NewBlock(source, parent)
}

func (alloc *Allocator) NewNative(rv reflect.Value) *NativeValue {
	alloc.AllocateNative()
	return &NativeValue{
		Value: rv,
	}
}

func (alloc *Allocator) NewType(t Type) Type {
	alloc.AllocateType()
	return t
}
