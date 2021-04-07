package config

type Config struct {
	Programs []Program `yaml:"programs"`
}

type Program struct {
	Name        string            `yaml:"name"`
	Metrics     Metrics           `yaml:"metrics"`
	Kprobes     map[string]string `yaml:"kprobes"`
	Kretprobes  map[string]string `yaml:"kretprobes"`
	Tracepoints map[string]string `yaml:"tracepoints"`
	Code        string            `yaml:"code"`
	Cflags      []string          `yaml:"cflags"`
}

type Metrics struct {
	Counters   []Counter   `yaml:"counters"`
	Histograms []Histogram `yaml:"histograms"`
}

type Counter struct {
	Name   string  `yaml:"name"`
	Help   string  `yaml:"help"`
	Table  string  `yaml:"table"`
	Labels []Label `yaml:"labels"`
}

type Histogram struct {
	Name             string              `yaml:"name"`
	Help             string              `yaml:"help"`
	Table            string              `yaml:"table"`
	BucketType       HistogramBucketType `yaml:"bucket_type"`
	BucketMultiplier float64             `yaml:"bucket_multiplier"`
	BucketMin        int                 `yaml:"bucket_min"`
	BucketMax        int                 `yaml:"bucket_max"`
	BucketKeys       []float64           `yaml:"bucket_keys"`
	Labels           []Label             `yaml:"labels"`
}

type Label struct {
	Name     string    `yaml:"name"`
	Size     uint      `yaml:"size"`
	Decoders []Decoder `yaml:"decoders"`
}

type Decoder struct {
	Name         string `yaml:"name"`
	AllowUnknown bool   `yaml:"allow_unknown"`
}

// HistogramBucketType is an enum to define how to interpret histogram
type HistogramBucketType string

const (
	// HistogramBucketExp2 means histograms with power-of-two keys
	HistogramBucketExp2 = "exp2"
	// HistogramBucketLinear means histogram with linear keys
	HistogramBucketLinear = "linear"
	// HistogramBucketFixed means histogram with fixed user-defined keys
	HistogramBucketFixed = "fixed"
)
