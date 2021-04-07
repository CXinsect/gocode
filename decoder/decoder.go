package decoder

import (
	"errors"
	"sync"

	"github.com/CXinsect/gocode/config"
)

var ErrSkipLabelSet = errors.New("this label of Set has been skipped")

type Decoder interface {
	Decode([]byte, config.Decoder) ([]byte, error)
}

type Set struct {
	mu       sync.Mutex
	decoders map[string]Decoder
}

func NewSet() *Set {
	return &Set{
		decoders: map[string]Decoder{
			"uint": &UInt{},
		},
	}
}
