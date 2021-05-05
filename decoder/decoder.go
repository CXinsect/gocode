package decoder

import (
	"errors"
	"fmt"
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

const (
	FunctionConnect = 0
	FunctionAccept  = 1
	ProgramSoFamily = 3
)

func NewSet() *Set {
	return &Set{
		decoders: map[string]Decoder{
			"uint":        &UInt{},
			"stringsys":   &StringSys{},
			"uintsoftirq": &UIntSoftIrq{},
			"uintmem":     &UIntMem{},
		},
	}
}

func (s *Set) Decode(in []byte, label config.Label) ([]byte, error) {
	result := in
	for _, decoder := range label.Decoders {
		if _, ok := s.decoders[decoder.Name]; !ok {
			return nil, fmt.Errorf("have no decoders named %s", decoder.Name)
		}
		s.mu.Lock()
		decoded, err := s.decoders[decoder.Name].Decode(in, decoder)
		s.mu.Unlock()
		if err != nil {
			return nil, fmt.Errorf("label named %s decoded faield %s", label.Name, err)
		}
		result = decoded
	}
	return result, nil
}
func (s *Set) DecodeLabels(in []byte, labels []config.Label) ([]string, error) {
	values := make([]string, len(labels))
	off := uint(0)
	totalSize := uint(0)

	for _, label := range labels {
		size := label.Size
		if size == 0 {
			return nil, fmt.Errorf("decodelabels labels size is zero err")
		}
		totalSize += size
	}
	if totalSize != uint(len(in)) {
		return nil, fmt.Errorf("key %#v in size %d is not equal with tatalSize %d", in, len(in), totalSize)
	}
	for i, label := range labels {
		if len(label.Decoders) == 0 {
			return nil, fmt.Errorf("label %s have no decoders", label.Name)
		}
		size := label.Size
		decoded, err := s.Decode(in[off:off+size], label)
		if err != nil {
			return nil, fmt.Errorf("label named %s decoded failed %s", label.Name, err)
		}
		off += size
		values[i] = string(decoded)
	}
	return values, nil

}
