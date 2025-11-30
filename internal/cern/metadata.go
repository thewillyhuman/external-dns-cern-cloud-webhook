package cern

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/thewillyhuman/external-dns-cern-cloud-webhook/internal/log"
	"sigs.k8s.io/external-dns/endpoint"
)

const (
	maxMetadataLength = 254
	landbAliasPrefix  = "landb-alias"
)

// GenerateMetadata calculates the required OpenStack metadata for a given node index and list of endpoints.
func GenerateMetadata(nodeIndex int, endpoints []*endpoint.Endpoint) map[string]string {
	var aliases []string
	for _, ep := range endpoints {
		// Only A records are supported for now based on the description
		if ep.RecordType == endpoint.RecordTypeA {
			// Format: <alias>--load-<index>-
			// Note: The prompt implies the alias is the DNS name.
			// Remove trailing dot if present
			dnsName := strings.TrimSuffix(ep.DNSName, ".")
			alias := fmt.Sprintf("%s--load-%d-", dnsName, nodeIndex)
			aliases = append(aliases, alias)
		}
	}

	// Sort aliases to ensure deterministic output
	sort.Strings(aliases)

	// Distribute aliases into keys: landb-alias, landb-alias2, landb-alias3...
	metadata := make(map[string]string)
	currentKeyIndex := 1
	var currentBuilder strings.Builder
	first := true

	for _, alias := range aliases {
		// Calculate potential length: current + comma (if not first) + alias
		potentialLen := currentBuilder.Len() + len(alias)
		if !first {
			potentialLen++ // for comma
		}

		if potentialLen > maxMetadataLength {
			// Flush current builder
			key := getMetadataKey(currentKeyIndex)
			metadata[key] = currentBuilder.String()

			// Reset builder and increment key index
			currentBuilder.Reset()
			currentKeyIndex++
			first = true
		}

		if !first {
			currentBuilder.WriteString(",")
		}
		currentBuilder.WriteString(alias)
		first = false
	}

	// Flush remaining
	if currentBuilder.Len() > 0 {
		key := getMetadataKey(currentKeyIndex)
		metadata[key] = currentBuilder.String()
	}

	return metadata
}

func getMetadataKey(index int) string {
	if index == 1 {
		return landbAliasPrefix
	}
	return fmt.Sprintf("%s%d", landbAliasPrefix, index)
}

// DiffMetadata compares the current metadata with the desired metadata.
// It returns a map of updates (keys to set) and a slice of keys to delete.
// Note: This logic assumes we own all `landb-alias*` keys.
func DiffMetadata(current map[string]string, desired map[string]string) (map[string]string, []string) {
	toUpdate := make(map[string]string)
	toDelete := []string{}

	// Check for updates or new keys
	for k, v := range desired {
		if curVal, ok := current[k]; !ok || curVal != v {
			toUpdate[k] = v
		}
	}

	// Check for keys to delete (present in current but not in desired, and starts with landb-alias)
	for k := range current {
		if strings.HasPrefix(k, landbAliasPrefix) {
			if _, ok := desired[k]; !ok {
				toDelete = append(toDelete, k)
			}
		}
	}

	return toUpdate, toDelete
}

// ParseEndpointsFromMetadata reconstructs endpoints from the `landb-alias` metadata of a set of servers.
// This is primarily for the `Records()` call.
// logic:
// 1. Iterate all servers.
// 2. Collect all alias strings.
// 3. Extract the DNS name from `<dnsname>--load-<index>-`.
// 4. Deduplicate.
func ParseEndpointsFromMetadata(nodes []servers.Server) []*endpoint.Endpoint {
	uniqueDomains := make(map[string]struct{})

	for _, node := range nodes {
		for key, value := range node.Metadata {
			if strings.HasPrefix(key, landbAliasPrefix) {
				// Value is comma-separated aliases
				aliases := strings.Split(value, ",")
				for _, alias := range aliases {
					alias = strings.TrimSpace(alias)
					// Parse: foo.cern.ch--load-0-
					// Find last occurrence of "--load-"
					idx := strings.LastIndex(alias, "--load-")
					if idx != -1 {
						domain := alias[:idx]
						uniqueDomains[domain] = struct{}{}
					}
				}
			}
		}
	}

	result := make([]*endpoint.Endpoint, 0, len(uniqueDomains))
	for domain := range uniqueDomains {
		// ExternalDNS expects endpoints.
		// Since we don't strictly know the targets (IPs) just from metadata (the metadata *implies* the node IPs),
		// we might construct Endpoints with dummy targets or try to infer them.
		// However, for the `Records` call, ExternalDNS mainly cares about "what records do you think you have?".
		// It uses this to calculate deletions.
		// If we return a record "foo.cern.ch", ExternalDNS knows it exists.
		// We set RecordType A.
		ep := endpoint.NewEndpoint(domain, endpoint.RecordTypeA, "") // Target is implicit/not strictly needed for ownership check?
		// Actually, ExternalDNS uses targets to check for updates.
		// But in this provider, the "Target" is effectively the set of Node IPs.
		// If we return empty targets, ExternalDNS might try to update it every time.
		// Let's leave targets empty for now or put a placeholder.
		// Ideally, we should list the IPs of the nodes that *should* be serving this.
		// But that's expensive to compute here (which nodes have the metadata?).
		// Let's assume just existence matters for now.
		result = append(result, ep)
	}
	return result
}

// FilterServers filters the list of servers based on the ingress label.
// If labelKey is empty, it returns all servers sorted by ID.
func FilterServers(allServers []servers.Server, labelKey string) []servers.Server {
	var filtered []servers.Server

	if labelKey == "" {
		filtered = make([]servers.Server, len(allServers))
		copy(filtered, allServers)
	} else {
		for _, server := range allServers {
			if _, ok := server.Metadata[labelKey]; ok {
				filtered = append(filtered, server)
			}
		}
	}

	// Deterministic sort by ID
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ID < filtered[j].ID
	})

	return filtered
}

// Log is a helper to get the logger
var Log = log.GlobalLogger
