package main

import (
	"time"

	"github.com/google/gopacket/layers"
	"github.com/ooni/netem"
	"github.com/ooni/probe-engine/pkg/model"
)

// newClientLinkConfig configures the link that the client should use.
func newClientLinkConfig(dpi string) *netem.LinkConfig {
	switch dpi {
	case "none":
		return linkWithoutCensorship()
	case "dns":
		return linkThatSpoofsDNS()
	case "tcp":
		return linkThatDropsTCPSYN()
	case "tls":
		return linkThatResetsTLSHandshake()
	default:
		panic("unsupported -dpi value (supported values: none, dns, tcp, tls)")
	}
}

// linkWithoutCensorship models a link without any censorship.
func linkWithoutCensorship() *netem.LinkConfig {
	return &netem.LinkConfig{
		DPIEngine:        nil,
		LeftNICWrapper:   nil,
		LeftToRightDelay: 15 * time.Millisecond,
		LeftToRightPLR:   0,
		RightNICWrapper:  nil,
		RightToLeftDelay: 15 * time.Millisecond,
		RightToLeftPLR:   0,
	}
}

// linkThatSpoofsDNS is a link that spoofs DNS responses.
func linkThatSpoofsDNS() *netem.LinkConfig {
	// create the default config.
	config := linkWithoutCensorship()

	// create the DPI engine.
	dpiEngine := netem.NewDPIEngine(model.DiscardLogger)

	// add DPI rule.
	dpiEngine.AddRule(&netem.DPISpoofDNSResponse{
		Addresses: []string{"10.10.34.35"},
		Logger:    model.DiscardLogger,
		Domain:    twitterDomain,
	})

	// assign the DPI engine.
	config.DPIEngine = dpiEngine

	// return the configured link to the caller.
	return config
}

// linkThatDropsTCPSYN is a link that drops a specific TCP SYN segment.
func linkThatDropsTCPSYN() *netem.LinkConfig {
	// create the default config.
	config := linkWithoutCensorship()

	// create the DPI engine.
	dpiEngine := netem.NewDPIEngine(model.DiscardLogger)

	// add DPI rule.
	dpiEngine.AddRule(&netem.DPIDropTrafficForServerEndpoint{
		Logger:          model.DiscardLogger,
		ServerIPAddress: twitterAddress,
		ServerPort:      443,
		ServerProtocol:  layers.IPProtocolTCP,
	})

	// assign the DPI engine.
	config.DPIEngine = dpiEngine

	// return the configured link to the caller.
	return config
}

// linkThatResetsTLSHandshake is a link that resets during the TLS handshake.
func linkThatResetsTLSHandshake() *netem.LinkConfig {
	// create the default config.
	config := linkWithoutCensorship()

	// create the DPI engine.
	dpiEngine := netem.NewDPIEngine(model.DiscardLogger)

	// add DPI rule.
	dpiEngine.AddRule(&netem.DPIResetTrafficForTLSSNI{
		Logger: model.DiscardLogger,
		SNI:    twitterDomain,
	})

	// assign the DPI engine.
	config.DPIEngine = dpiEngine

	// return the configured link to the caller.
	return config
}
