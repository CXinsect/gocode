package decoder

import (
	"fmt"
	"strconv"

	"github.com/CXinsect/gocode/config"
	"github.com/iovisor/gobpf/bcc"
)

// UInt is a decoder that transforms unsigned integers into their string values
type UIntSoftIrq struct{}

// Decode transforms unsigned integers into their string values
func (u *UIntSoftIrq) Decode(in []byte, conf config.Decoder) ([]byte, error) {
	byteOrder := bcc.GetHostByteOrder()
	switch len(in) {

	case 16:
		vec := byteOrder.Uint64(in[0:8])
		slot := byteOrder.Uint64(in[8:16])
		return []byte("softirq_" + vecToName(vec) + ":" + strconv.Itoa(int(slot))), nil
	default:
		return nil, fmt.Errorf("unknown size of decoder uint24")
	}
}

var nameArr = [...]string{"hi", "timer", "net_tx", "net_rx", "block", "irq_poll",
	"tasklet", "sched", "hrtimer", "rcu"}

func vecToName(nr uint64) string {
	return nameArr[nr]
}
