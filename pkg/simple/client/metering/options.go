package metering

type Options struct {
	Enable bool `json:"enable" yaml:"enable"`
}

func NewMeteringOptions() *Options {
	return &Options{}
}
