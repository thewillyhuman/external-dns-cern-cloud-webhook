// Package provider contains the implementation of the ExternalDNS provider.
//
// This package is responsible for the business logic of the webhook, which involves
// interacting with the CERN Cloud DNS service to manage DNS records.
// The Provider struct is the main entry point for this package, and its methods
// correspond to the endpoints defined in the ExternalDNS webhook provider specification.
//
// The current implementation is a scaffold and does not yet contain the actual logic
// for managing DNS records. The TODO comments indicate where this logic needs to be added.
package provider

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/internal/cern"
	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/internal/k8s"
	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/internal/log"
	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/pkg/config"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

// Provider is the main struct for the webhook provider.
type Provider struct {
	config  *config.Config
	manager *cern.Manager
}

// NewProvider creates a new instance of the Provider.
func NewProvider(cfg *config.Config) *Provider {
	client, err := cern.NewClient(cfg)
	if err != nil {
		log.GlobalLogger.Error("Failed to create OpenStack client: %v", err)
		os.Exit(1)
	}

	k8sClient, err := k8s.NewClient()
	if err != nil {
		log.GlobalLogger.Error("Failed to create Kubernetes client: %v", err)
		os.Exit(1)
	}

	return &Provider{
		config:  cfg,
		manager: cern.NewManager(client, k8sClient),
	}
}

// Records implements the GET /records endpoint.
func (p *Provider) Records(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.GlobalLogger.Info("received request for Records from %s", r.RemoteAddr)

	nodes, err := p.manager.GetIngressNodes(ctx, p.config.IngressLabel)
	if err != nil {
		log.GlobalLogger.Error("Failed to get ingress nodes: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	endpoints := cern.ParseEndpointsFromMetadata(nodes)

	w.Header().Set("Content-Type", "application/vnd.external-dns.error+json; version=1")
	if err := json.NewEncoder(w).Encode(endpoints); err != nil {
		log.GlobalLogger.Error("Failed to encode records: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// AdjustEndpoints implements the POST /adjustendpoints endpoint.
func (p *Provider) AdjustEndpoints(w http.ResponseWriter, r *http.Request) {
	log.GlobalLogger.Info("received request for AdjustEndpoints from %s", r.RemoteAddr)

	var endpoints []*endpoint.Endpoint
	if err := json.NewDecoder(r.Body).Decode(&endpoints); err != nil {
		log.GlobalLogger.Error("Failed to decode endpoints: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// For now, we return the endpoints as is.
	// In the future, we could implement logic to filter or modify them.
	w.Header().Set("Content-Type", "application/vnd.external-dns.error+json; version=1")
	if err := json.NewEncoder(w).Encode(endpoints); err != nil {
		log.GlobalLogger.Error("Failed to encode adjusted endpoints: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ApplyChanges implements the POST /records endpoint.
func (p *Provider) ApplyChanges(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.GlobalLogger.Info("received request for ApplyChanges from %s", r.RemoteAddr)

	var changes plan.Changes
	if err := json.NewDecoder(r.Body).Decode(&changes); err != nil {
		log.GlobalLogger.Error("Failed to decode changes: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Get current nodes
	nodes, err := p.manager.GetIngressNodes(ctx, p.config.IngressLabel)
	if err != nil {
		log.GlobalLogger.Error("Failed to get ingress nodes: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Get current endpoints
	currentEndpoints := cern.ParseEndpointsFromMetadata(nodes)

	// 3. Calculate desired endpoints
	// Helper map to deduplicate and manage state
	desiredMap := make(map[string]*endpoint.Endpoint)
	for _, ep := range currentEndpoints {
		desiredMap[ep.DNSName] = ep
	}

	// Apply deletions
	for _, ep := range changes.Delete {
		delete(desiredMap, ep.DNSName)
	}

	// Apply updates (remove old, add new)
	for _, ep := range changes.UpdateOld {
		delete(desiredMap, ep.DNSName)
	}
	for _, ep := range changes.UpdateNew {
		desiredMap[ep.DNSName] = ep
	}

	// Apply creations
	for _, ep := range changes.Create {
		desiredMap[ep.DNSName] = ep
	}

	// Convert map back to slice
	var desiredEndpoints []*endpoint.Endpoint
	for _, ep := range desiredMap {
		desiredEndpoints = append(desiredEndpoints, ep)
	}

	// 4. Sync state
	if p.config.DryRun {
		log.GlobalLogger.Info("Dry run enabled, skipping actual update")
	} else {
		if err := p.manager.SyncState(ctx, nodes, desiredEndpoints); err != nil {
			log.GlobalLogger.Error("Failed to sync state: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// Negotiate implements the GET / endpoint.
func (p *Provider) Negotiate(w http.ResponseWriter, r *http.Request) {
	log.GlobalLogger.Info("received request for Negotiate from %s", r.RemoteAddr)
	// Return basic info. ExternalDNS usually expects specific headers or body
	// for negotiation if it was a sophisticated plugin, but for basic webhook
	// it often just checks connectivity.
	// However, we can return some metadata.
	// For now, just OK.
	w.WriteHeader(http.StatusOK)
}

// Healthz implements the GET /healthz endpoint.
func (p *Provider) Healthz(w http.ResponseWriter, r *http.Request) {
	log.GlobalLogger.Info("received request for Healthz from %s", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
}
