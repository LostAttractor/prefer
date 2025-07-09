// Package prefer is a CoreDNS plugin that provides prefer ipversion functionality.
package prefer

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("prefer")

// Prefer is a plugin that provides prefer ipversion functionality.
type Prefer struct {
	Next    plugin.Handler
	version string
}

// ServeDNS implements the plugin.Handler interface.
func (f *Prefer) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	qtype := state.QType()

	// Prefered version will not be filtered
	switch {
	case qtype != dns.TypeA && qtype != dns.TypeAAAA,
		f.version == "ipv4" && qtype == dns.TypeA,
		f.version == "ipv6" && qtype == dns.TypeAAAA:
		return plugin.NextOrFailure(f.Name(), f.Next, ctx, w, r)
	}

	var checkType uint16
	if qtype == dns.TypeAAAA {
		checkType = dns.TypeA
	} else {
		checkType = dns.TypeAAAA
	}

	// Create a new request to check for the other IP type.
	checkReq := new(dns.Msg)
	checkReq.SetQuestion(state.QName(), checkType)
	checkReq.RecursionDesired = r.RecursionDesired

	// Use a custom ResponseWriter to capture the response from the next plugin
	recorder := NewResponseRecorder(w)

	// Call next plugin to resolve the check request
	_, err := f.Next.ServeDNS(ctx, recorder, checkReq)
	if err != nil {
		log.Debugf("Error while checking for alternate record type: %v", err)
		return plugin.NextOrFailure(f.Name(), f.Next, ctx, w, r)
	}

	// If there are answers for the other type, we filter the original request.
	if recorder.msg != nil && len(recorder.msg.Answer) > 0 {
		log.Debugf("Filtering request for %s type %s because type %s record exists", state.QName(), dns.TypeToString[qtype], dns.TypeToString[checkType])
		requestFilteredCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

		// Return an empty NOERROR response
		m := new(dns.Msg)
		m.SetReply(r)
		w.WriteMsg(m)
		return dns.RcodeSuccess, nil
	}

	// Otherwise, process the original request normally
	return plugin.NextOrFailure(f.Name(), f.Next, ctx, w, r)
}

// Name implements the Handler interface.
func (f *Prefer) Name() string { return "prefer" }

// responseRecorder is a dns.ResponseWriter that records the response.
type responseRecorder struct {
	dns.ResponseWriter
	msg *dns.Msg
}

func NewResponseRecorder(w dns.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w}
}

func (r *responseRecorder) WriteMsg(res *dns.Msg) error {
	r.msg = res
	return nil
}
