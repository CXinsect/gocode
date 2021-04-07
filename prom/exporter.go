package prom

import (
	"fmt"

	"github.com/iovisor/gobpf/bcc"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/CXinsect/gocode/config"
	"github.com/CXinsect/gocode/decoder"
)

const namespace = "ebpf_exporter"

type Exporter struct {
	config             config.Config
	module             map[string]*bcc.Module
	ksyms              map[uint64]string
	enabledProgramDesc *prometheus.Desc
	programInfoDesc    *prometheus.Desc
	programTags        map[string]map[string]uint64
	descs              map[string]map[string]*prometheus.Desc
	decoder            *decoder.Set
}

func New(config config.Config) *Exporter {
	enabledProgramDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "enabled_programs"),
		"enabledProgramDesc",
		[]string{"name"},
		nil,
	)
	programInfoDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "programInfo"),
		"programInfoDesc",
		[]string{"function"},
		nil,
	)
	return &Exporter{
		config:             config,
		module:             map[string]*bcc.Module{},
		ksyms:              map[uint64]string{},
		enabledProgramDesc: enabledProgramDesc,
		programInfoDesc:    programInfoDesc,
		programTags:        map[string]map[string]uint64{},
		descs:              map[string]map[string]*prometheus.Desc{},
		decoder:            decoder.NewSet(),
	}
}
func (e *Exporter) Attach() error {
	for _, program := range e.config.Programs {
		if _, ok := e.module[program.Name]; ok {
			return fmt.Errorf("the Module Name %s has repeated", program.Name)
		}
		module := bcc.NewModule(program.Code, program.Cflags)
		if module == nil {
			return fmt.Errorf("create module failed with program name %s", program.Name)
		}
		//进行映射的逻辑
		tags, err := attach(module, program.Kprobes, program.Tracepoints)
		if err != nil {
			return fmt.Errorf("program %s attch failed %s", program.Name, err)
		}

		e.programTags[program.Name] = tags

		e.module[program.Name] = module
	}
	return nil
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	desc := func(programName, name, help string, label []config.Label) {
		if _, ok := e.descs[programName][name]; !ok {
			labelName := []string{}
			for _, v := range label {
				labelName = append(labelName, v.Name)
			}
			e.descs[programName][name] = prometheus.NewDesc(prometheus.BuildFQName(namespace, "", name),
				help,
				labelName,
				nil,
			)
		}
		ch <- e.descs[programName][name]
	}
	ch <- e.enabledProgramDesc
	ch <- e.programInfoDesc

	for _, program := range e.config.Programs {
		if _, ok := e.descs[program.Name]; !ok {
			e.descs[program.Name] = map[string]*prometheus.Desc{}
		}
		for _, counter := range program.Metrics.Counters {
			desc(program.Name, counter.Name, counter.Help, counter.Labels)
		}
		for _, histogram := range program.Metrics.Histograms {
			desc(program.Name, histogram.Name, histogram.Help, histogram.Labels)
		}
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	for _, programa := range e.config.Programs {
		ch <- prometheus.MustNewConstMetric(e.enabledProgramDesc, prometheus.GaugeValue, 1, programa.Name)
	}
	for programa, tagMap := range e.programTags {
		for function, _ := range tagMap {
			ch <- prometheus.MustNewConstMetric(e.programInfoDesc, prometheus.GaugeValue, 1, programa, function)
		}
	}
	//开始收集
}

func (e *Exporter) collectCounter(ch chan<- prometheus.Metric) {

}

func (e *Exporter) collectHistogram(ch chan<- prometheus.Metric) {

}
