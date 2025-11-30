package cern

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/pkg/config"
)

// Client wraps the Gophercloud compute client.
type Client struct {
	Compute *gophercloud.ServiceClient
}

// NewClient creates a new OpenStack compute client.
func NewClient(cfg *config.Config) (*Client, error) {
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: cfg.OpenStackAuthURL,
		Username:         cfg.OpenStackUsername,
		Password:         cfg.OpenStackPassword,
		DomainName:       cfg.OpenStackUserDomainName,
		TenantName:       cfg.OpenStackProjectName,
	}

	// Create a custom HTTP client to handle potential TLS issues or proxies if needed.
	// For now, we use a standard client but allow for expansion.
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // TODO: Add config for insecure
		},
	}

	provider, err := openstack.NewClient(opts.IdentityEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenStack client: %w", err)
	}
	provider.HTTPClient = *httpClient

	if err := openstack.Authenticate(provider, opts); err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	endpointOpts := gophercloud.EndpointOpts{
		Region: cfg.OpenStackRegionName,
	}

	compute, err := openstack.NewComputeV2(provider, endpointOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	return &Client{Compute: compute}, nil
}
