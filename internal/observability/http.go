package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "bar108_http_requests_total",
		Help: "HTTP requests by method, route, and status code.",
	}, []string{"method", "route", "status"})
	duration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "bar108_http_request_duration_seconds",
		Help:    "HTTP request duration by method and route.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
)

func init() {
	prometheus.MustRegister(requests, duration)
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		requests.WithLabelValues(c.Request.Method, route, strconv.Itoa(c.Writer.Status())).Inc()
		duration.WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
	}
}

func Handler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}
