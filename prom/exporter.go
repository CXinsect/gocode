package prom

import (
	"fmt"
	"log"
	"strconv"
	"strings"

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

		for _, perfEvent := range program.PerfEvents {
			target, err := module.LoadPerfEvent(perfEvent.Target)
			if err != nil {
				return fmt.Errorf("perEvent targeted %s in program named %s load PerfEvent error %s", perfEvent.Target, program.Name, err)
			}

			err = module.AttachPerfEvent(perfEvent.Type, perfEvent.Name, perfEvent.Period, perfEvent.Frequency, -1, -1, -1, target)

			if err != nil {
				return fmt.Errorf("perEvent targeted %s in program named %s attach perevent error %s", perfEvent.Target, program.Name, err)
			}

		}
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
	for _, tagMap := range e.programTags {
		for function, _ := range tagMap {
			ch <- prometheus.MustNewConstMetric(e.programInfoDesc, prometheus.GaugeValue, 1, function)
		}
	}
	//TODO收集
	e.collectCounter(ch)
	e.collectHistogram(ch)
}

func (e *Exporter) collectCounter(ch chan<- prometheus.Metric) {
	for _, program := range e.config.Programs {
		for _, counter := range program.Metrics.Counters {
			mvs, err := e.tableKeyAndValues(e.module[program.Name], counter.Table, counter.Labels)
			if err != nil {
				log.Printf("program named %s Table named %s Counter named %s get table values error\n", program.Name, counter.Table, counter.Name)
				continue
			}
			desc := e.descs[program.Name][counter.Name]
			for _, mv := range mvs {
				ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, mv.value, mv.labels...)
			}
		}
	}
}

func (e *Exporter) collectHistogram(ch chan<- prometheus.Metric) {
	for _, program := range e.config.Programs {
		for _, histogram := range program.Metrics.Histograms {
			skip := false
			mvs, err := e.tableKeyAndValues(e.module[program.Name], histogram.Table, histogram.Labels)
			if err != nil {
				log.Printf("program named %s Table named %s Histogram named %s get table values error\n", program.Name, histogram.Table, histogram.Name)
				continue
			}
			histograms := map[string]histogramWithLabels{}
			for _, mv := range mvs {
				labels := mv.labels[0 : len(mv.labels)-1]
				key := fmt.Sprintf("%#v", labels)
				tmp_v := mv.labels[len(mv.labels)-1]
				var str []string
				var e_label string = "default"
				if strings.Contains(tmp_v, ":") {
					str = strings.Split(tmp_v, ":")
					tmp_v = str[1]
					e_label = str[0]
				}
				labels = append(labels, e_label)
				if _, ok := histograms[key]; !ok {
					histograms[key] = histogramWithLabels{
						labels:  labels,
						buckets: map[float64]uint64{},
					}
				}
				leUint, err := strconv.ParseUint(tmp_v, 0, 64)
				if err != nil {
					log.Printf("histogram get value transform failed %s", err)
					skip = true
					break
				}
				histograms[key].buckets[float64(leUint)] = uint64(mv.value)
			}
			if skip {
				continue
			}
			desc := e.descs[program.Name][histogram.Name]
			for _, histogramSet := range histograms {
				bucket, count, sum, err := transformHistogram(histogramSet.buckets, histogram)
				if err != nil {
					log.Printf("histogram named %s transformHistogram failed", histogram.Name)
					continue
				}
				ch <- prometheus.MustNewConstHistogram(desc, count, sum, bucket, histogramSet.labels...)
			}

		}
	}
}

type metricValue struct {
	labels []string
	value  float64
}

func (e *Exporter) tableKeyAndValues(module *bcc.Module, tableName string, labels []config.Label) ([]metricValue, error) {
	values := []metricValue{}
	table := bcc.NewTable((module.TableId(tableName)), module)
	iter := table.Iter()

	if iter != nil {
		for iter.Next() {
			key := iter.Key()
			raw, _ := table.KeyBytesToStr(key)
			fmt.Println("The content of key: ", raw)
			mv := metricValue{
				labels: make([]string, len(labels)),
			}
			var err error
			mv.labels, err = e.decoder.DecodeLabels(key, labels)
			if err != nil {
				return nil, err
			}
			mv.value = float64(bcc.GetHostByteOrder().Uint64(iter.Leaf()))
			values = append(values, mv)
		}
	}
	return values, nil
}
