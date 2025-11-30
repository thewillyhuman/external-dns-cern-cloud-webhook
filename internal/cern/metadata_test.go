package cern

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"sigs.k8s.io/external-dns/endpoint"
)

func TestGenerateMetadata(t *testing.T) {
	tests := []struct {
		name      string
		nodeIndex int
		endpoints []*endpoint.Endpoint
		expected  map[string]string
	}{
		{
			name:      "Empty endpoints",
			nodeIndex: 0,
			endpoints: []*endpoint.Endpoint{},
			expected:  map[string]string{},
		},
		{
			name:      "Single endpoint",
			nodeIndex: 0,
			endpoints: []*endpoint.Endpoint{
				{DNSName: "foo.cern.ch", RecordType: endpoint.RecordTypeA},
			},
			expected: map[string]string{
				"landb-alias": "foo.cern.ch--load-0-",
			},
		},
		{
			name:      "Multiple endpoints",
			nodeIndex: 1,
			endpoints: []*endpoint.Endpoint{
				{DNSName: "foo.cern.ch", RecordType: endpoint.RecordTypeA},
				{DNSName: "bar.cern.ch", RecordType: endpoint.RecordTypeA},
			},
			expected: map[string]string{
				"landb-alias": "bar.cern.ch--load-1-,foo.cern.ch--load-1-", // sorted
			},
		},
		{
			name:      "Overflow 254 chars",
			nodeIndex: 0,
			endpoints: func() []*endpoint.Endpoint {
				// Generate enough endpoints to overflow 254 chars
				// "alias-X.cern.ch--load-0-" is approx 24 chars.
				// 10 of them is 240. 11 will overflow.
				eps := []*endpoint.Endpoint{}
				for i := 0; i < 15; i++ {
					eps = append(eps, &endpoint.Endpoint{
						DNSName:    fmt.Sprintf("alias-%02d.cern.ch", i),
						RecordType: endpoint.RecordTypeA,
					})
				}
				return eps
			}(),
			// Expected outcome depends on sorting and comma overhead.
			// Just checking if we have multiple keys.
			expected: nil, // We'll check keys manually in test logic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateMetadata(tt.nodeIndex, tt.endpoints)
			if tt.expected != nil {
				if !reflect.DeepEqual(got, tt.expected) {
					t.Errorf("GenerateMetadata() = %v, want %v", got, tt.expected)
				}
			} else {
				// For the overflow case, check if we have landb-alias and landb-alias2
				if _, ok := got["landb-alias"]; !ok {
					t.Errorf("Expected landb-alias key")
				}
				if _, ok := got["landb-alias2"]; !ok {
					t.Errorf("Expected landb-alias2 key")
				}
				// Verify lengths
				for k, v := range got {
					if len(v) > 254 {
						t.Errorf("Key %s has value length %d > 254", k, len(v))
					}
				}
			}
		})
	}
}

func TestParseEndpointsFromMetadata(t *testing.T) {
	tests := []struct {
		name     string
		nodes    []servers.Server
		expected []string // Just check DNS names for simplicity
	}{
		{
			name:     "No metadata",
			nodes:    []servers.Server{{ID: "1"}},
			expected: []string{},
		},
		{
			name: "Single node single alias",
			nodes: []servers.Server{
				{
					Metadata: map[string]string{
						"landb-alias": "foo.cern.ch--load-0-",
					},
				},
			},
			expected: []string{"foo.cern.ch"},
		},
		{
			name: "Multiple nodes same alias",
			nodes: []servers.Server{
				{
					Metadata: map[string]string{
						"landb-alias": "foo.cern.ch--load-0-",
					},
				},
				{
					Metadata: map[string]string{
						"landb-alias": "foo.cern.ch--load-1-",
					},
				},
			},
			expected: []string{"foo.cern.ch"}, // Deduped
		},
		{
			name: "Multiple keys",
			nodes: []servers.Server{
				{
					Metadata: map[string]string{
						"landb-alias":  "foo.cern.ch--load-0-",
						"landb-alias2": "bar.cern.ch--load-0-",
					},
				},
			},
			expected: []string{"foo.cern.ch", "bar.cern.ch"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseEndpointsFromMetadata(tt.nodes)
			gotNames := make(map[string]bool)
			for _, ep := range got {
				gotNames[ep.DNSName] = true
			}

			if len(got) != len(tt.expected) {
				t.Errorf("ParseEndpointsFromMetadata() returned %d endpoints, want %d", len(got), len(tt.expected))
			}

			for _, name := range tt.expected {
				if !gotNames[name] {
					t.Errorf("ParseEndpointsFromMetadata() missing %s", name)
				}
			}
		})
	}
}

func TestDiffMetadata(t *testing.T) {
	tests := []struct {
		name    string
		current map[string]string
		desired map[string]string
		wantUpd map[string]string
		wantDel []string
	}{
		{
			name:    "No changes",
			current: map[string]string{"landb-alias": "foo"},
			desired: map[string]string{"landb-alias": "foo"},
			wantUpd: map[string]string{},
			wantDel: []string{},
		},
		{
			name:    "Update existing",
			current: map[string]string{"landb-alias": "foo"},
			desired: map[string]string{"landb-alias": "bar"},
			wantUpd: map[string]string{"landb-alias": "bar"},
			wantDel: []string{},
		},
		{
			name:    "Add new",
			current: map[string]string{"landb-alias": "foo"},
			desired: map[string]string{"landb-alias": "foo", "landb-alias2": "bar"},
			wantUpd: map[string]string{"landb-alias2": "bar"},
			wantDel: []string{},
		},
		{
			name:    "Delete existing",
			current: map[string]string{"landb-alias": "foo", "landb-alias2": "bar"},
			desired: map[string]string{"landb-alias": "foo"},
			wantUpd: map[string]string{},
			wantDel: []string{"landb-alias2"},
		},
		{
			name:    "Ignore non-landb keys",
			current: map[string]string{"other": "keepme", "landb-alias": "foo"},
			desired: map[string]string{"landb-alias": "bar"},
			wantUpd: map[string]string{"landb-alias": "bar"},
			wantDel: []string{}, // "other" should not be deleted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUpd, gotDel := DiffMetadata(tt.current, tt.desired)
			if !reflect.DeepEqual(gotUpd, tt.wantUpd) {
				t.Errorf("DiffMetadata() updates = %v, want %v", gotUpd, tt.wantUpd)
			}

			// Sort deletes for comparison
			sort.Strings(gotDel)
			sort.Strings(tt.wantDel)
			if !reflect.DeepEqual(gotDel, tt.wantDel) {
				t.Errorf("DiffMetadata() deletes = %v, want %v", gotDel, tt.wantDel)
			}
		})
	}
}
