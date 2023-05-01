package metrics

import "github.com/prometheus/client_golang/prometheus"

var totalRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "http_requests_total",
	Help: "Number of incoming requests",
}, []string{"path"})

var totalHTTPMethods = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "http_methods_total",
	Help: "Number of requests per HTTP method",
}, []string{"method"})

var httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_response_time_seconds",
	Help: "Duration of HTTP requests",
}, []string{"path"})

func init() {
	if err := prometheus.Register(totalRequests); err != nil {
		panic(err)
	}
	if err := prometheus.Register(totalHTTPMethods); err != nil {
		panic(err)
	}
	if err := prometheus.Register(httpDuration); err != nil {
		panic(err)
	}
}
