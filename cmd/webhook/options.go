// Package main contains the logic for parsing command-line flags and environment variables.
package main

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/pkg/config"
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
	pflag.String("openstack-auth-url", "", "OpenStack Auth URL")
	pflag.String("openstack-project-name", "", "OpenStack Project Name")
	pflag.String("openstack-user-domain-name", "", "OpenStack User Domain Name")
	pflag.String("openstack-project-domain-id", "", "OpenStack Project Domain ID")
	pflag.String("openstack-username", "", "OpenStack Username")
	pflag.String("openstack-password", "", "OpenStack Password")
	pflag.String("openstack-region-name", "", "OpenStack Region Name")
	pflag.String("openstack-interface", "", "OpenStack Interface")
	pflag.String("openstack-identity-api-version", "", "OpenStack Identity API Version")
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
		OpenStackAuthURL:            v.GetString("openstack-auth-url"),
		OpenStackProjectName:        v.GetString("openstack-project-name"),
		OpenStackUserDomainName:     v.GetString("openstack-user-domain-name"),
		OpenStackProjectDomainID:    v.GetString("openstack-project-domain-id"),
		OpenStackUsername:           v.GetString("openstack-username"),
		OpenStackPassword:           v.GetString("openstack-password"),
		OpenStackRegionName:         v.GetString("openstack-region-name"),
		OpenStackInterface:          v.GetString("openstack-interface"),
		OpenStackIdentityAPIVersion: v.GetString("openstack-identity-api-version"),
	}

	// Validate that all required OpenStack configuration parameters are present.
	requiredConfigs := []struct {
		value string
		name  string
	}{
		{cfg.OpenStackAuthURL, "openstack-auth-url"},
		{cfg.OpenStackProjectName, "openstack-project-name"},
		{cfg.OpenStackUserDomainName, "openstack-user-domain-name"},
		{cfg.OpenStackProjectDomainID, "openstack-project-domain-id"},
		{cfg.OpenStackUsername, "openstack-username"},
		{cfg.OpenStackPassword, "openstack-password"},
		{cfg.OpenStackRegionName, "openstack-region-name"},
		{cfg.OpenStackInterface, "openstack-interface"},
		{cfg.OpenStackIdentityAPIVersion, "openstack-identity-api-version"},
	}

	for _, required := range requiredConfigs {
		if required.value == "" {
			return nil, fmt.Errorf("missing required configuration: --%s", required.name)
		}
	}

	return cfg, nil
}
