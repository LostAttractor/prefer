package prefer

import (
	"strings"
	"testing"

	"github.com/coredns/caddy"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
		errString string
	}{
		{`prefer ipv4`, false, ""},
		{`prefer ipv6`, false, ""},
		{`prefer`, true, "Wrong argument count"},
		{`prefer ipv4 extra`, true, "Wrong argument count"},
		{`prefer blah`, true, "invalid ip version preference"},
	}

	for i, test := range tests {
		c := caddy.NewTestController("dns", test.input)
		err := setup(c)
		if test.shouldErr {
			if err == nil {
				t.Errorf("Test %d: Expected error but got nil for input %s", i+1, test.input)
			} else if !strings.Contains(err.Error(), test.errString) {
				t.Errorf("Test %d: Expected error to contain '%s', but got '%s'", i+1, test.errString, err.Error())
			}
		} else {
			if err != nil {
				t.Errorf("Test %d: Expected no error but got %v for input %s", i+1, err, test.input)
			}
		}
	}
}
