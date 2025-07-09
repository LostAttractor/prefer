package prefer

import (
	"sync"

	"github.com/coredns/coredns/plugin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// requestFilteredCount is a prometheus metric that is incremented every time a query is filtered.
var requestFilteredCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: plugin.Namespace,
	Subsystem: "prefer",
	Name:      "request_filtered_count_total",
	Help:      "Counter of filtered requests.",
}, []string{"server"})

var once sync.Once
