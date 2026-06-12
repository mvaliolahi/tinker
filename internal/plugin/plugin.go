package plugin

import (
	"github.com/spf13/cobra"
)

// Service is the interface that each built-in service type must implement.
// This allows tinker to be extended with new service types (Redis, Kafka, etc.)
// without modifying core code.
type Service interface {
	// Name returns the service identifier (e.g., "database", "api", "grpc", "log").
	Name() string

	// Label returns the UI badge label for the dashboard.
	Label() string

	// Detect runs auto-detection on the project directory.
	// Returns nil if the service is not detected.
	Detect(dir string) interface{}

	// Commands returns the cobra commands to register for this service.
	Commands() []*cobra.Command

	// IsConfigured checks if the config has this service set up.
	IsConfigured(cfg interface{}) bool
}

// registry holds all registered services.
var registry []Service

// Register adds a service to the global registry.
func Register(s Service) {
	registry = append(registry, s)
}

// Services returns all registered services.
func Services() []Service {
	return registry
}

// Find returns a service by name.
func Find(name string) Service {
	for _, s := range registry {
		if s.Name() == name {
			return s
		}
	}
	return nil
}
