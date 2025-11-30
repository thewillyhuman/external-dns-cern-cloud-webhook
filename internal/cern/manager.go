package cern

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/internal/k8s"
	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/internal/log"
	"sigs.k8s.io/external-dns/endpoint"
)

// Manager handles the interaction with OpenStack servers and metadata.
type Manager struct {
	client    *Client
	k8sClient *k8s.Client
}

// NewManager creates a new Manager.
func NewManager(client *Client, k8sClient *k8s.Client) *Manager {
	return &Manager{
		client:    client,
		k8sClient: k8sClient,
	}
}

// GetIngressNodes retrieves all OpenStack servers that correspond to Kubernetes nodes matching the label.
func (m *Manager) GetIngressNodes(ctx context.Context, labelKey string) ([]servers.Server, error) {
	// 1. Get K8s Node Names
	nodeNames, err := m.k8sClient.GetIngressNodeNames(ctx, labelKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get ingress node names from k8s: %w", err)
	}

	// Create a map for O(1) lookups
	targetNames := make(map[string]struct{})
	for _, name := range nodeNames {
		targetNames[name] = struct{}{}
	}

	// 2. List all OpenStack servers
	// We list all active servers and filter client-side by name.
	opts := servers.ListOpts{
		Status: "ACTIVE",
	}

	pager := servers.List(m.client.Compute, opts)
	var matchingServers []servers.Server

	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			return false, err
		}

		for _, server := range serverList {
			if _, ok := targetNames[server.Name]; ok {
				matchingServers = append(matchingServers, server)
			}
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list openstack servers: %w", err)
	}

	// 3. Sort for deterministic behavior
	FilterServers(matchingServers, "")

	return matchingServers, nil
}

// UpdateNodeMetadata updates the metadata of a specific node.
func (m *Manager) UpdateNodeMetadata(ctx context.Context, serverID string, toUpdate map[string]string, toDelete []string) error {
	// Update items
	if len(toUpdate) > 0 {
		log.GlobalLogger.Info("Updating metadata for server %s: %v", serverID, toUpdate)
		_, err := servers.UpdateMetadata(m.client.Compute, serverID, servers.MetadataOpts(toUpdate)).Extract()
		if err != nil {
			return fmt.Errorf("failed to update metadata for server %s: %w", serverID, err)
		}
	}

	// Delete items
	for _, key := range toDelete {
		log.GlobalLogger.Info("Deleting metadata key %s for server %s", key, serverID)
		err := servers.DeleteMetadatum(m.client.Compute, serverID, key).ExtractErr()
		if err != nil {
			// If it's already gone, maybe ignore? But for now report error.
			return fmt.Errorf("failed to delete metadata key %s for server %s: %w", key, serverID, err)
		}
	}

	return nil
}

// SyncState synchronizes the state of all ingress nodes to match the desired endpoints.
func (m *Manager) SyncState(ctx context.Context, nodes []servers.Server, endpoints []*endpoint.Endpoint) error {
	// 1. Calculate desired state for each node.
	// 2. Diff with current state.
	// 3. Apply changes.

	// We process nodes in order (0, 1, 2...).
	for i, node := range nodes {
		desiredMetadata := GenerateMetadata(i, endpoints)
		currentMetadata := node.Metadata

		toUpdate, toDelete := DiffMetadata(currentMetadata, desiredMetadata)

		if len(toUpdate) > 0 || len(toDelete) > 0 {
			if err := m.UpdateNodeMetadata(ctx, node.ID, toUpdate, toDelete); err != nil {
				return err // Return on first error? Or try best effort? "Atomic as possible".
				// If we fail halfway, we might leave inconsistent state.
				// But we can't rollback easily.
				// Returning error will likely cause ExternalDNS to retry.
			}
		}
	}

	return nil
}
