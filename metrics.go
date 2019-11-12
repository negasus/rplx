package rplx

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	variablesGot               *prometheus.CounterVec
	variablesSent              *prometheus.CounterVec
	variablesSentResponseCodes *prometheus.CounterVec
	variablesSentDuration      *prometheus.HistogramVec
}

func newMetrics() *metrics {
	m := &metrics{}

	m.variablesGot = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "rplx_variables_got",
		Help: "Rplx Variables Got",
	}, []string{"remote_node_id"})

	m.variablesSent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "rplx_variables_sent",
		Help: "Rplx Variables Sent",
	}, []string{"remote_node_id"})

	m.variablesSentResponseCodes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "rplx_variables_sent_response_codes",
		Help: "Rplx Variables Sent Response Codes",
	}, []string{"remote_node_id", "code"})

	m.variablesSentDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "rplx_variables_sent_duration",
		Help:    "Rplx Variables Sent Duration",
		Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1, 2, 5},
	}, []string{"remote_node_id"})

	return m
}

func (m *metrics) register() {
	prometheus.MustRegister(m.variablesGot)
	prometheus.MustRegister(m.variablesSent)
	prometheus.MustRegister(m.variablesSentResponseCodes)
	prometheus.MustRegister(m.variablesSentDuration)
}
