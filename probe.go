package main

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/ooni/probe-engine/pkg/dslx"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

// probeMain simulates what ooniprobe would do.
func probeMain() {
	// create an ID generator.
	idgen := &atomic.Int64{}

	// obtain the IP addresses to use for twitter.com.
	addrs := probeDNSLookup(idgen)

	// make sure we're getting the expected result.
	runtimex.Assert(len(addrs.M) == 1, "expected exactly one IP address")

	// fetch the webpage using the discovered IP address(es).
	probeFetchUsingHTTPS(idgen, twitterDomain, addrs)
}

// probeDNSLookup resolves the twitter.com domain to a set of IP addresses.
func probeDNSLookup(idgen *atomic.Int64) *dslx.AddressSet {
	// describe what we want to resolve using the DNS.
	domain := dslx.NewDomainToResolve(
		dslx.DomainName(twitterDomain),
		dslx.DNSLookupOptionLogger(loggerSingleton),
		dslx.DNSLookupOptionIDGenerator(idgen),
	)

	// define the function to use to resolve.
	fx := dslx.DNSLookupUDP("8.8.8.8:53")

	// apply the function to its arguments.
	result := fx.Apply(context.Background(), domain)

	// make sure the result is the expected one and return.
	runtimex.Assert(result.Error == nil, "unexpected DNS lookup failure")
	return dslx.NewAddressSet(result)
}

// probeFetchUsingHTTPS fetches a webpage using HTTPS.
func probeFetchUsingHTTPS(idgen *atomic.Int64, domain string, addressSet *dslx.AddressSet) {
	// convert the set of addresses to TCP endpoints.
	//
	// For example, [1.2.3.4] => [1.2.3.4:443/tcp].
	endpoints := addressSet.ToEndpoints(
		dslx.EndpointNetwork("tcp"),
		dslx.EndpointPort(443),
		dslx.EndpointOptionLogger(loggerSingleton),
		dslx.EndpointOptionIDGenerator(idgen),
	)

	// make sure we have a single endpoint -- (this is a simplifying assumption)
	runtimex.Assert(len(endpoints) == 1, "expected a single endpoint")
	endpoint := endpoints[0]

	// create connections pool
	pool := &dslx.ConnPool{}
	defer pool.Close()

	// define the function to use to fetch the webpage.
	fx := dslx.Compose3(
		dslx.TCPConnect(pool),
		dslx.TLSHandshake(
			pool,
			dslx.TLSHandshakeOptionServerName(twitterDomain),
		),
		dslx.HTTPRequestOverTLS(
			dslx.HTTPRequestOptionHost(twitterDomain),
		),
	)

	// obtain the results
	results := fx.Apply(context.Background(), endpoint)

	// handle an error
	if results.Error != nil {
		loggerSingleton.Warnf("cannot fetch webpage: %s", results.Error.Error())
		return
	}

	// print the webpage we obtained
	fmt.Printf("\n%s\n", string(results.State.HTTPResponseBodySnapshot))
}
