package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(c.Request.URL.Path))

		totalRequests.WithLabelValues(c.Request.URL.Path).Inc()
		totalHTTPMethods.WithLabelValues(c.Request.Method).Inc()

		c.Next()

		timer.ObserveDuration()
	}
}
