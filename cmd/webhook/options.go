// Package main contains the logic for parsing command-line flags and environment variables.
package main

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/pkg/config"
)

const (
	OpenStackAuthURL            = "os-auth-url"
	OpenStackProjectName        = "os-project-name"
	OpenStackUserDomainName     = "os-user-domain-name"
	OpenStackProjectDomainID    = "os-project-domain-id"
	OpenStackUsername           = "os-username"
	OpenStackPassword           = "os-password"
	OpenStackRegionName         = "os-region-name"
	OpenStackInterface          = "os-interface"
	OpenStackIdentityAPIVersion = "os-identity-api-version"
)

// loadConfig initializes and returns the application's configuration.
//
// This function is responsible for defining all command-line flags, setting up viper
// to read from environment variables, and then populating the config.Config struct.
// This approach centralizes all command-line and environment variable handling in the
// cmd package, cleanly separating it from the application's core configuration definition.
func loadConfig() (*config.Config, error) {
	// Define command-line flags using the pflag library.
	// Each flag is defined with a name, a default value, and a description.
	// The descriptions are used to generate the help text for the application.
	pflag.String("listen-address", "0.0.0.0", "The IP address to listen on")
	pflag.Int("listen-port", 8888, "The port to listen on")
	pflag.String("log-level", "info", "Log level (debug, info, warn, error)")
	pflag.String(OpenStackAuthURL, "", "OpenStack Auth URL")
	pflag.String(OpenStackProjectName, "", "OpenStack Project Name")
	pflag.String(OpenStackUserDomainName, "", "OpenStack User Domain Name")
	pflag.String(OpenStackProjectDomainID, "", "OpenStack Project Domain ID")
	pflag.String(OpenStackUsername, "", "OpenStack Username")
	pflag.String(OpenStackPassword, "", "OpenStack Password")
	pflag.String(OpenStackRegionName, "", "OpenStack Region Name")
	pflag.String(OpenStackInterface, "", "OpenStack Interface")
	pflag.String(OpenStackIdentityAPIVersion, "", "OpenStack Identity API Version")
	pflag.Bool("dry-run", false, "Run in dry-run mode")
	pflag.String("ingress-label", "node-role.kubernetes.io/ingress", "Label to filter ingress nodes")
	pflag.StringSlice("domain-filter", []string{}, "Filter domains")
	pflag.StringSlice("exclude-domains", []string{}, "Exclude domains")
	pflag.String("txt-prefix", "", "TXT record prefix")
	pflag.String("txt-suffix", "", "TXT record suffix")
	pflag.Parse()

	// Initialize viper to manage configuration.
	// Viper is a powerful library that can read configuration from various sources,
	// including environment variables, config files, and remote key-value stores.
	v := viper.New()

	// Configure viper to automatically read environment variables.
	// The SetEnvKeyReplacer is used to map environment variables with underscores
	// to command-line flags with hyphens (e.g., LISTEN_ADDRESS to --listen-address).
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	// Bind the pflag command-line flags to viper.
	// This allows viper to read the values of the flags and makes them available
	// through the viper interface.
	if err := v.BindPFlags(pflag.CommandLine); err != nil {
		return nil, err
	}

	// Create a new Config object and populate it with the values from viper.
	// The GetString and GetInt methods are used to retrieve the values of the
	// configuration options.
	cfg := &config.Config{
		ListenAddress:               v.GetString("listen-address"),
		ListenPort:                  v.GetInt("listen-port"),
		LogLevel:                    v.GetString("log-level"),
		OpenStackAuthURL:            v.GetString(OpenStackAuthURL),
		OpenStackProjectName:        v.GetString(OpenStackProjectName),
		OpenStackUserDomainName:     v.GetString(OpenStackUserDomainName),
		OpenStackProjectDomainID:    v.GetString(OpenStackProjectDomainID),
		OpenStackUsername:           v.GetString(OpenStackUsername),
		OpenStackPassword:           v.GetString(OpenStackPassword),
		OpenStackRegionName:         v.GetString(OpenStackRegionName),
		OpenStackInterface:          v.GetString(OpenStackInterface),
		OpenStackIdentityAPIVersion: v.GetString(OpenStackIdentityAPIVersion),
		DryRun:                      v.GetBool("dry-run"),
		IngressLabel:                v.GetString("ingress-label"),
		DomainFilter:                v.GetStringSlice("domain-filter"),
		ExcludeDomains:              v.GetStringSlice("exclude-domains"),
		TXTPrefix:                   v.GetString("txt-prefix"),
		TXTSuffix:                   v.GetString("txt-suffix"),
	}

	// Validate that all required OpenStack configuration parameters are present.
	requiredConfigs := []struct {
		value string
		name  string
	}{
		{cfg.OpenStackAuthURL, OpenStackAuthURL},
		{cfg.OpenStackProjectName, OpenStackProjectName},
		{cfg.OpenStackUserDomainName, OpenStackUserDomainName},
		{cfg.OpenStackProjectDomainID, OpenStackProjectDomainID},
		{cfg.OpenStackUsername, OpenStackUsername},
		{cfg.OpenStackPassword, OpenStackPassword},
		{cfg.OpenStackRegionName, OpenStackRegionName},
		{cfg.OpenStackInterface, OpenStackInterface},
		{cfg.OpenStackIdentityAPIVersion, OpenStackIdentityAPIVersion},
	}

	for _, required := range requiredConfigs {
		if required.value == "" {
			return nil, fmt.Errorf("missing required configuration: --%s", required.name)
		}
	}

	return cfg, nil
}
