package decoder

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/CXinsect/gocode/config"
	"github.com/iovisor/gobpf/bcc"
)

// UInt is a decoder that transforms unsigned integers into their string values
type UInt struct{}

// Decode transforms unsigned integers into their string values
func (u *UInt) Decode(in []byte, conf config.Decoder) ([]byte, error) {
	byteOrder := bcc.GetHostByteOrder()

	result := uint64(0)
	switch len(in) {
	case 32:
		flag := byteOrder.Uint64(in[0:8])
		front := byteOrder.Uint64(in[8:16])
		back := byteOrder.Uint64(in[16:24])
		return generateLabelsOfKey_32(flag, front, back)
	case 24:
		flag := byteOrder.Uint64(in[0:8])
		front := byteOrder.Uint64(in[8:16])
		back := byteOrder.Uint64(in[16:])
		return generateLabelsOfKey_24(flag, front, back)
	case 16:
		front := byteOrder.Uint64(in[0:8])
		back := byteOrder.Uint64(in[8:])
		return []byte(string(strconv.Itoa(int(front))) + ":" + string(strconv.Itoa(int(back)))), nil
	case 8:
		result = byteOrder.Uint64(in)
	case 4:
		result = uint64(byteOrder.Uint32(in))
	case 2:
		result = uint64(byteOrder.Uint16(in))
	case 1:
		result = uint64(in[0])
	default:
		return nil, fmt.Errorf("unknown value length %d for %#v", len(in), in)
	}

	return []byte(strconv.Itoa(int(result))), nil
}

// func getNameOfKey16(pid uint64) []byte {
// 	str := fmt.Sprintf("/proc/%d/cmdline", int(pid))
// 	cmd := exec.Command("cat", str)
// 	var stderr bytes.Buffer
// 	cmd.Stderr = &stderr
// 	pid_name, err := cmd.Output()
// 	if len(pid_name) > 15 || err != nil {
// 		pid_name = []byte(strconv.Itoa(int(pid)))
// 	}
// 	return pid_name
// }

func generateLabelsOfKey_32(flag, pid, proto uint64) ([]byte, error) {
	var ret string
	pid_name := getPidName(pid, proto)
	if flag == 0 {
		ret = "connect:" + string(pid_name) + ":" + tranProtoMap(proto)
	} else {
		ret = "accept:" + string(pid_name) + ":" + tranProtoMap(proto)
	}
	return []byte(ret), nil
}
func generateLabelsOfKey_24(flag, pid, proto uint64) ([]byte, error) {
	var ret string
	pid_name := getPidName(pid, proto)
	if flag == 0 {
		ret = "connect:" + string(pid_name) + ":" + netProtoMap(proto)
	} else {
		ret = "accept:" + string(pid_name) + ":" + netProtoMap(proto)
	}
	return []byte(ret), nil
}
func getPidName(pid, proto uint64) string {
	str := fmt.Sprintf("/proc/%d/cmdline", int(pid))
	cmd := exec.Command("cat", str)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	pid_name, err := cmd.Output()
	if len(pid_name) > 15 || err != nil {
		pid_name = []byte(strconv.Itoa(int(pid)))
	}
	return string(pid_name)
}
func netProtoMap(proto uint64) string {
	switch proto {
	case 0:
		return "AF_UNSPEC"
	case 1:
		return "AF_UNIX"
	case 2:
		return "AF_INET"
	case 10:
		return "AF_INET6"
	default:
		return "unknown protocol"
	}
}

func tranProtoMap(proto uint64) string {
	switch proto {
	case 0:
		return "IPPROTO_IP"
	case 1:
		return "IPPROTO_ICMP"
	case 6:
		return "IPPROTO_TCP"
	case 17:
		return "IPPROTO_UDP"
	default:
		return "unknown proto"
	}
}
