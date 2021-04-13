package decoder

import (
	"fmt"
	"strconv"

	"github.com/CXinsect/gocode/config"
	"github.com/iovisor/gobpf/bcc"
)

// UInt is a decoder that transforms unsigned integers into their string values
type UIntMem struct{}

// Decode transforms unsigned integers into their string values
func (u *UIntMem) Decode(in []byte, conf config.Decoder) ([]byte, error) {
	byteOrder := bcc.GetHostByteOrder()

	switch len(in) {
	case 16:
		tflag := byteOrder.Uint64(in[0:8])
		pid := byteOrder.Uint64(in[8:16])
		return generateMemRelativedLabels(tflag, pid)
	default:
		return nil, fmt.Errorf("unknown value length %d for %#v", len(in), in)
	}
}

func generateMemRelativedLabels(flag, pid uint64) ([]byte, error) {
	if flag == 0 {
		return []byte("kprobes_handle_mm_fault" + ":" + strconv.Itoa(int(pid))), nil
	} else {
		return []byte("brk_transfered" + ":" + strconv.Itoa(int(pid))), nil
	}
}
