// Package config provides a centralized configuration management system for the application.
//
// This package is responsible for defining the application's configuration structure.
// The use of a dedicated configuration package promotes a clean separation of concerns
// and makes it easier to manage the application's settings.
package config

// Config holds all the configuration for the application.
//
// This struct is a single source of truth for all application settings.
// Each field corresponds to a specific configuration option.
// The use of a single configuration struct makes it easy to pass configuration
// throughout the application and to see all the available options in one place.
type Config struct {
	// ListenAddress is the IP address that the webhook server will listen on.
	ListenAddress string
	// ListenPort is the port that the webhook server will listen on.
	ListenPort int
	// LogLevel is the logging level for the application.
	LogLevel string
	// OpenStackAuthURL is the URL of the OpenStack Keystone authentication service.
	OpenStackAuthURL string
	// OpenStackProjectName is the name of the OpenStack project to use.
	OpenStackProjectName string
	// OpenStackUserDomainName is the name of the OpenStack user's domain.
	OpenStackUserDomainName string
	// OpenStackProjectDomainID is the ID of the OpenStack project's domain.
	OpenStackProjectDomainID string
	// OpenStackUsername is the username for authenticating with OpenStack.
	OpenStackUsername string
	// OpenStackPassword is the password for authenticating with OpenStack.
	OpenStackPassword string
	// OpenStackRegionName is the name of the OpenStack region to use.
	OpenStackRegionName string
	// OpenStackInterface is the network interface to use for OpenStack services.
	OpenStackInterface string
	// OpenStackIdentityAPIVersion is the version of the OpenStack Identity API to use.
	OpenStackIdentityAPIVersion string
}
