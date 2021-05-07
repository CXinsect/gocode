package prom

import (
	"fmt"

	"github.com/iovisor/gobpf/bcc"
)

const binaryPath string = "/home/go/exporter/tmp/oom"

//"/root/project/gocode/tmp/oom"

type attacheType func(*bcc.Module, map[string]string) (map[string]uint64, error)

func generateTags(dst map[string]uint64, attach attacheType, module *bcc.Module, attachcontents map[string]string) error {
	src, err := attach(module, attachcontents)
	if err != nil {
		return err
	}
	for name, tag := range src {
		dst[name] = tag
	}
	return nil
}

func attach(module *bcc.Module, kprobes, tracepoints, uprobes map[string]string) (map[string]uint64, error) {
	tags := map[string]uint64{}

	if err := generateTags(tags, attachKprobes, module, kprobes); err != nil {
		return nil, fmt.Errorf("failed to attach kprobes %s", err)
	}
	if err := generateTags(tags, attachTracepoints, module, tracepoints); err != nil {
		return nil, fmt.Errorf("failed to attach tracepoints %s", err)
	}
	if err := generateTags(tags, attachUprobes, module, uprobes); err != nil {
		return nil, fmt.Errorf("failed to attach uprobes %s", err)
	}
	return tags, nil
}

type loader func(string) (int, error)
type attacher func(string, int) error
type attacherWithMax func(string, int, int) error
type attacherForUprobes func(string, string, int, int) error

func attachMerges(module *bcc.Module, probeLoad loader, probeAttach attacher, probes map[string]string) (map[string]uint64, error) {
	tags := map[string]uint64{}

	for probe, tName := range probes {
		target, err := probeLoad(tName)
		if err != nil {
			return nil, fmt.Errorf("func attachLoad failed %s", err)
		}
		tag, err := module.GetProgramTag(target)
		if err != nil {
			return nil, fmt.Errorf("get program Tag failed %s", err)
		}
		tags[tName] = tag
		err = probeAttach(probe, target)
		if err != nil {
			return nil, fmt.Errorf("probeAttach %s:%s failed %s ", probe, tName, err)
		}

	}
	return tags, nil
}
func withMaxValue(attach attacherWithMax, max int) attacher {
	return func(probe string, target int) error {
		return attach(probe, target, max)
	}
}
func attachKprobes(module *bcc.Module, kprobes map[string]string) (map[string]uint64, error) {
	return attachMerges(module, module.LoadKprobe, withMaxValue(module.AttachKprobe, 0), kprobes)
}
func attachTracepoints(module *bcc.Module, tracepoint map[string]string) (map[string]uint64, error) {
	return attachMerges(module, module.LoadTracepoint, module.AttachTracepoint, tracepoint)
}

func packageUprobes(attachPackage attacherForUprobes, path string, pid int) attacher {
	return func(name string, fd int) error {
		return attachPackage(path, name, fd, pid)
	}
}
func attachUprobes(module *bcc.Module, uprobes map[string]string) (map[string]uint64, error) {
	return attachMerges(module, module.LoadUprobe, packageUprobes(module.AttachUprobe, binaryPath, -1), uprobes)
}
