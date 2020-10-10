package mecab

// #include <mecab.h>
// #include <stdlib.h>
import "C"
import (
	"reflect"
	"runtime"
	"unsafe"
)

type lattice struct {
	lattice *C.mecab_lattice_t
}

func newLattice(l *C.mecab_lattice_t) *lattice {
	ret := &lattice{
		lattice: l,
	}
	runtime.SetFinalizer(ret, finalizeLattice)
	return ret
}

// It is a marker that a lattice must not be copied after the first use.
// See https://github.com/golang/go/issues/8005#issuecomment-190753527
// for details.
func (*lattice) Lock() {}

func finalizeLattice(l *lattice) {
	if l.lattice != nil {
		C.mecab_lattice_destroy(l.lattice)
	}
	l.lattice = nil
}

// Lattice is a lattice.
type Lattice struct {
	l *lattice
}

// NewLattice creates new lattice.
func NewLattice() (Lattice, error) {
	// C.mecab_lattice_new sets an error in the thread local storage.
	// so C.mecab_lattice_new and C.mecab_strerror must be call in same thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	l := C.mecab_lattice_new()
	if l == nil {
		return Lattice{}, newError(nil)
	}
	return Lattice{l: newLattice(l)}, nil
}

// Destroy frees the lattice.
func (l Lattice) Destroy() {
	C.mecab_lattice_destroy(l.l.lattice)
	l.l.lattice = nil
	runtime.KeepAlive(l.l)
}

// Clear set empty string to the lattice.
func (l Lattice) Clear() {
	C.mecab_lattice_clear(l.l.lattice)
	runtime.KeepAlive(l.l)
}

// IsAvailable returns the lattice is available.
func (l Lattice) IsAvailable() bool {
	return C.mecab_lattice_is_available(l.l.lattice) != 0
}

// BOSNode returns the Begin Of Sentence node.
func (l Lattice) BOSNode() Node {
	return Node{
		node:    C.mecab_lattice_get_bos_node(l.l.lattice),
		lattice: l.l,
	}
}

// EOSNode returns the End Of Sentence node.
func (l Lattice) EOSNode() Node {
	return Node{
		node:    C.mecab_lattice_get_eos_node(l.l.lattice),
		lattice: l.l,
	}
}

// Sentence returns the sentence in the lattice.
func (l Lattice) Sentence() string {
	s := C.GoString(C.mecab_lattice_get_sentence(l.l.lattice))
	runtime.KeepAlive(l.l)
	return s
}

// SetSentence set the sentence in the lattice.
func (l Lattice) SetSentence(s string) {
	length := C.size_t(len(s))
	if s == "" {
		s = "dummy"
	}
	header := (*reflect.StringHeader)(unsafe.Pointer(&s))
	input := (*C.char)(unsafe.Pointer(header.Data))

	C.mecab_lattice_add_request_type(l.l.lattice, 64) // MECAB_ALLOCATE_SENTENCE = 64
	C.mecab_lattice_set_sentence2(l.l.lattice, input, length)
	runtime.KeepAlive(l.l)
}

func (l Lattice) String() string {
	return C.GoString(C.mecab_lattice_tostr(l.l.lattice))
}
