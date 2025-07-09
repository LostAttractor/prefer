package prefer

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

// init registers this plugin.
func init() { plugin.Register("prefer", setup) }

// setup is the function that gets called when the config parser see the token "prefer".
func setup(c *caddy.Controller) error {
	c.Next() // Ignore "prefer" and give us the next token.

	if !c.NextArg() {
		return plugin.Error("prefer", c.ArgErr())
	}

	version := c.Val()
	if version != "ipv4" && version != "ipv6" {
		return plugin.Error("prefer", c.Errf("invalid ip version preference: %s, must be 'ipv4' or 'ipv6'", version))
	}

	if c.NextArg() {
		// If there was another token, return an error, because we don't have more configuration.
		// Any errors returned from this setup function should be wrapped with plugin.Error, so we
		// can present a slightly nicer error message to the user.
		return plugin.Error("prefer", c.ArgErr())
	}

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return &Prefer{Next: next, version: version}
	})

	// All OK, return a nil error.
	return nil
}
