package main

import (
	"context"
	"flag"
	"time"

	"github.com/ooni/netem"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/netemx"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

func main() {
	// define the DPI flag to use.
	dpiFlag := flag.String("dpi", "none", "select DPI (one of: none, dns, tcp, tls)")
	flag.Parse()

	// create context controlling the overall lifecycle
	ctx, cancel := context.WithCancel(context.Background())

	// create a new star topology
	topology := runtimex.Try1(netem.NewStarTopology(model.DiscardLogger))
	defer topology.Close()

	// run 8.8.8.8:53 simulator in the background
	quad8Ready := make(chan any)
	go simulateQuad8(ctx, topology, quad8Ready)
	<-quad8Ready

	// run twitter.com simulator in the background
	twitterReady := make(chan any)
	go simulateTwitter(ctx, topology, twitterReady)
	<-twitterReady

	// create a client link config
	clientLinkConfig := newClientLinkConfig(*dpiFlag)

	// setup packet captures
	loggerSingleton.Info("writing packet capture at client.pcap")
	clientLinkConfig.LeftNICWrapper = netem.NewPCAPDumper("client.pcap", model.DiscardLogger)

	// create a netstack for the probe
	probeStack := runtimex.Try1(topology.AddHost(
		"192.168.0.174", // probe IP address
		quad8Address,    // default resolver address
		clientLinkConfig,
	))

	// run code using the given netstack
	netemx.WithCustomTProxy(probeStack, probeMain)

	// await one additional second to capture more packets
	loggerSingleton.Info("wating for packet capture to finish")
	time.Sleep(1 * time.Second)

	// shutdown background runners
	cancel()
}
