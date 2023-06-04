package main

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/ooni/netem"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

// twitterAddress is the IP address used by Twitter.
const twitterAddress = "104.244.42.193"

// twitterDomain is the domain used by Twitter.
const twitterDomain = "twitter.com"

// twitterWebPage is the web page returned by Twitter.
const twitterWebPage = `<HTML>
<HEAD>
	<TITLE>Home / Twitter</TITLE>
</HEAD>
<BODY>
	<P>Welcome to Twitter!</P>
</BODY>
</HTML>
`

// simulateTwitter simulates the Twitter home page.
func simulateTwitter(ctx context.Context, topology *netem.StarTopology, readych chan<- any) {
	// create the twitter host
	//
	// note: because the stack is created using topology.AddHost, we don't
	// need to call Close when done using it, since the topology will do that
	// for us when we call the topology's Close method.
	serverStack := runtimex.Try1(topology.AddHost(
		twitterAddress, // server IP address
		quad8Address,   // default resolver address
		&netem.LinkConfig{
			LeftToRightDelay: 1 * time.Millisecond,
			RightToLeftDelay: 1 * time.Millisecond,
		},
	))

	// create HTTPS server using the serverStack
	tlsListener := runtimex.Try1(serverStack.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(twitterAddress),
		Port: 443,
		Zone: "",
	}))
	httpsServer := &http.Server{
		TLSConfig: serverStack.ServerTLSConfig(),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(twitterWebPage))
		}),
	}
	go httpsServer.ServeTLS(tlsListener, "", "")

	// let the caller know we're now ready to serve requests.
	close(readych)

	// block until the context is canceled
	<-ctx.Done()
}
