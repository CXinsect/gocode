package decoder

import (
	"fmt"

	"github.com/CXinsect/gocode/config"
	"github.com/iovisor/gobpf/bcc"
)

// String is a decoder that decodes strings coming from the kernel
type StringSys struct{}

// Decode transforms byte slice from the kernel into string
func (s *StringSys) Decode(in []byte, conf config.Decoder) ([]byte, error) {
	op := bcc.GetHostByteOrder().Uint64(in)
	switch op {
	case 0:
		return []byte("accept"), nil
	case 1:
		return []byte("connect"), nil
	case 2:
		return []byte("bind"), nil
	case 3:
		return []byte("socket"), nil
	default:
		return nil, fmt.Errorf("unknown value length %d for %#v", len(in), in)
	}
}
