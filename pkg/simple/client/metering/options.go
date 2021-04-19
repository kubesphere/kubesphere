package metering

type PriceInfo struct {
	CpuPerCorePerHour                        float64 `json:"cpuPerCorePerHour" yaml:"cpuPerCorePerHour"`
	MemPerGigabytesPerHour                   float64 `json:"memPerGigabytesPerHour" yaml:"memPerGigabytesPerHour"`
	IngressNetworkTrafficPerMegabytesPerHour float64 `json:"ingressNetworkTrafficPerMegabytesPerHour" yaml:"ingressNetworkTrafficPerGiagabytesPerHour"`
	EgressNetworkTrafficPerMegabytesPerHour  float64 `json:"egressNetworkTrafficPerMegabytesPerHour" yaml:"egressNetworkTrafficPerGigabytesPerHour"`
	PvcPerGigabytesPerHour                   float64 `json:"pvcPerGigabytesPerHour" yaml:"pvcPerGigabytesPerHour"`
	CurrencyUnit                             string  `json:"currencyUnit" yaml:"currencyUnit"`
}

type Billing struct {
	PriceInfo PriceInfo `json:"priceInfo" yaml:"priceInfo"`
}

type Options struct {
	RetentionDay string  `json:"retentionDay" yaml:"retentionDay"`
	Billing      Billing `json:"billing" yaml:"billing"`
}

var DefaultMeteringOption = Options{
	RetentionDay: "7d",
	Billing: Billing{
		PriceInfo: PriceInfo{
			CpuPerCorePerHour:                        0,
			MemPerGigabytesPerHour:                   0,
			IngressNetworkTrafficPerMegabytesPerHour: 0,
			EgressNetworkTrafficPerMegabytesPerHour:  0,
			PvcPerGigabytesPerHour:                   0,
			CurrencyUnit:                             "",
		},
	},
}

func NewMeteringOptions() *Options {
	return &Options{}
}
