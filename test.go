package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "nvidia_gpu"
)

var (
	addr = flag.String("web.listen-address", ":9445", "Address to listen on for web interface and telemetry.")

	labels = []string{"uuid", "name"}
)

type Collector struct {
	sync.Mutex
	testGaugeVec *prometheus.GaugeVec
	testDesc     *prometheus.Desc
	count        int
}

func NewCollector() *Collector {
	return &Collector{
		testGaugeVec: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "test_gaugevec",
				Help:      "Fanspeed of the GPU device as a percent of its maximum",
			},
			labels,
		),
		testDesc: prometheus.NewDesc("test_desc", "mytest", labels, nil),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.testDesc
	c.testGaugeVec.Describe(ch)
	// c.testDesc.Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	// Only one Collect call in progress at a time.
	c.Lock()
	defer c.Unlock()

	// 如果reset的话，会删除所有指标，否则指标会保存在内存里面
	c.testGaugeVec.Reset()
	if c.count%2 == 0 {
		c.testGaugeVec.WithLabelValues("first", fmt.Sprintf("%v", c.count)).Set(float64(c.count))
	}
	// 不会保存在内存中！
	ch <- prometheus.MustNewConstMetric(c.testDesc, prometheus.GaugeValue, float64(c.count), "first", fmt.Sprintf("%v", c.count))
	c.count++
	c.testGaugeVec.Collect(ch)
}

func main() {
	flag.Parse()
	r := prometheus.NewRegistry()

	r.MustRegister(NewCollector())
	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})

	// Serve on all paths under addr
	log.Fatalf("ListenAndServe error: %v", http.ListenAndServe(*addr, handler))
}
