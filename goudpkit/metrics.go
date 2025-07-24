package goudpkit

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	packetsSent = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "goudpkit_packets_sent_total",
		Help: "Total packets sent.",
	})
	packetsReceived = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "goudpkit_packets_received_total",
		Help: "Total packets received.",
	})
	packetsDropped = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "goudpkit_packets_dropped_total",
		Help: "Total packets dropped.",
	})
	retryCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "goudpkit_retry_count_total",
		Help: "Total retry attempts.",
	})
)

func RegisterMetrics() {
	prometheus.MustRegister(packetsSent, packetsReceived, packetsDropped, retryCount)
}

func ExportMetricsHTTP(addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, nil)
}

func IncPacketsSent()     { packetsSent.Inc() }
func IncPacketsReceived() { packetsReceived.Inc() }
func IncPacketsDropped()  { packetsDropped.Inc() }
func IncRetryCount()      { retryCount.Inc() }
