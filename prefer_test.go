package prefer

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

// testHandler returns a handler that will respond with an A record for `example.org.`
// and a AAAA record for `example.net.`.
func testHandler() test.HandlerFunc {
	return func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
		m := new(dns.Msg)
		m.SetReply(r)
		switch r.Question[0].Qtype {
		case dns.TypeA:
			if r.Question[0].Name == "example.org." {
				m.Answer = append(m.Answer, test.A("example.org. IN A 127.0.0.1"))
			}
		case dns.TypeAAAA:
			if r.Question[0].Name == "example.net." {
				m.Answer = append(m.Answer, test.AAAA("example.net. IN AAAA ::1"))
			}
		}
		w.WriteMsg(m)
		return dns.RcodeSuccess, nil
	}
}

func TestPrefer(t *testing.T) {
	ctx := context.Background()

	// Test case 1: prefer ipv4, AAAA query for example.org., A record exists. Should filtered.
	f1 := &Prefer{Next: testHandler(), version: "ipv4"}
	rec1 := dnstest.NewRecorder(&test.ResponseWriter{})
	r1 := new(dns.Msg)
	r1.SetQuestion("example.org.", dns.TypeAAAA)

	rcode1, err1 := f1.ServeDNS(ctx, rec1, r1)
	if err1 != nil {
		t.Fatalf("Test 1: Expected no error, got %v", err1)
	}
	if rcode1 != dns.RcodeSuccess {
		t.Fatalf("Test 1: Expected RcodeSuccess, got %d", rcode1)
	}
	if len(rec1.Msg.Answer) != 0 {
		t.Errorf("Test 1: Expected empty answer, got %d answers", len(rec1.Msg.Answer))
	}

	// Test case 2: prefer ipv4, AAAA query for example.net, A record does not exist. Should NOT filtered.
	// The next handler will return an AAAA record for example.net.
	f2 := &Prefer{Next: testHandler(), version: "ipv4"}
	rec2 := dnstest.NewRecorder(&test.ResponseWriter{})
	r2 := new(dns.Msg)
	r2.SetQuestion("example.net.", dns.TypeAAAA)

	f2.ServeDNS(ctx, rec2, r2)
	if len(rec2.Msg.Answer) != 1 {
		t.Errorf("Test 2: Expected 1 answer, got %d answers", len(rec2.Msg.Answer))
	}

	// Test case 3: prefer ipv6, A query for example.net, AAAA record exists. Should filtered.
	f3 := &Prefer{Next: testHandler(), version: "ipv6"}
	rec3 := dnstest.NewRecorder(&test.ResponseWriter{})
	r3 := new(dns.Msg)
	r3.SetQuestion("example.net.", dns.TypeA)

	f3.ServeDNS(ctx, rec3, r3)
	if len(rec3.Msg.Answer) != 0 {
		t.Errorf("Test 3: Expected empty answer, got %d answers", len(rec3.Msg.Answer))
	}

	// Test case 4: prefer ipv6, A query for example.org, AAAA record does not exist. Should NOT filtered.
	// The next handler will return an A record for example.org.
	f4 := &Prefer{Next: testHandler(), version: "ipv6"}
	rec4 := dnstest.NewRecorder(&test.ResponseWriter{})
	r4 := new(dns.Msg)
	r4.SetQuestion("example.org.", dns.TypeA)

	f4.ServeDNS(ctx, rec4, r4)
	if len(rec4.Msg.Answer) != 1 {
		t.Errorf("Test 4: Expected 1 answer, got %d answers", len(rec4.Msg.Answer))
	}
}
