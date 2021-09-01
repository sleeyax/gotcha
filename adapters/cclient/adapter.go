package cclient

import (
	tls "github.com/refraction-networking/utls"
	"github.com/sleeyax/gotcha"
	"github.com/x04/cclient"
)

type Adapter struct {
	// Optional proxy to connect to.
	proxyUrl string

	// TLS client hello ID to use.
	ClientHello tls.ClientHelloID
}

func parseProxy(proxies []string) string {
	if len(proxies) == 0 {
		return ""
	}
	return proxies[0]
}

func NewAdapter(clientHello tls.ClientHelloID, proxyUrl ...string) *Adapter {
	return &Adapter{
		proxyUrl:    parseProxy(proxyUrl),
		ClientHello: clientHello,
	}
}

func (a *Adapter) DoRequest(options *gotcha.Options) (*gotcha.Response, error) {
	if options.Proxy != nil {
		a.proxyUrl = options.Proxy.String()
	}

	client, err := cclient.NewClient(a.ClientHello, a.proxyUrl)
	if err != nil {
		return nil, err
	}

	requestAdapter := gotcha.RequestAdapter{
		RoundTripper: client.Transport,
	}

	return requestAdapter.DoRequest(options)
}
