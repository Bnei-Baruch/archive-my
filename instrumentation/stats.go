package instrumentation

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var Stats = new(Collectors)

type Collectors struct {
	RequestDurationHistogram *prometheus.HistogramVec
}

func (c *Collectors) Init() {
	c.RequestDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "feed_api",
		Subsystem: "api",
		Name:      "request_duration",
		Help:      "Time (in seconds) spent serving HTTP requests.",
	}, []string{"method", "route", "status_code"})

	prometheus.MustRegister(c.RequestDurationHistogram)

	prometheus.MustRegister(collectors.NewBuildInfoCollector())
}

func (c *Collectors) Reset() {
	c.RequestDurationHistogram.Reset()
}
