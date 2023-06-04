package main

import (
	"context"
	"time"

	"github.com/ooni/netem"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

// quad8Address is the 8.8.8.8 IPv4 address.
const quad8Address = "8.8.8.8"

// simulateQuad8 simulates the 8.8.8.8:53 DNS server.
func simulateQuad8(ctx context.Context, topology *netem.StarTopology, readych chan<- any) {
	// create the quad8 host
	//
	// note: because the stack is created using topology.AddHost, we don't
	// need to call Close when done using it, since the topology will do that
	// for us when we call the topology's Close method.
	serverStack := runtimex.Try1(topology.AddHost(
		quad8Address, // server IP address
		quad8Address, // default resolver address
		&netem.LinkConfig{
			LeftToRightDelay: 1 * time.Millisecond,
			RightToLeftDelay: 1 * time.Millisecond,
		},
	))

	// create configuration for the 8.8.8.8:53 DNS server.
	dnsConfig := netem.NewDNSConfig()
	dnsConfig.AddRecord(
		twitterDomain,
		"", // CNAME
		twitterAddress,
	)

	// create DNS server using the serverStack
	dnsServer := runtimex.Try1(netem.NewDNSServer(
		model.DiscardLogger,
		serverStack,
		quad8Address,
		dnsConfig,
	))
	defer dnsServer.Close()

	// let the caller know we're now ready to serve requests.
	close(readych)

	// block until the context is canceled
	<-ctx.Done()
}
