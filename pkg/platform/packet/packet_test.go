// Copyright 2020 The Lokomotive Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package packet

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCheckNotEmptyWorkersEmpty(t *testing.T) {
	c := config{}

	if d := c.checkNotEmptyWorkers(); !d.HasErrors() {
		t.Errorf("Expected to fail with empty workers")
	}
}

func TestCheckNotEmptyWorkers(t *testing.T) {
	c := config{WorkerPools: []workerPool{{Name: "test"}}}

	if d := c.checkNotEmptyWorkers(); d.HasErrors() {
		t.Errorf("Should not fail with no duplicated worker pool names")
	}
}

func TestCheckWorkerPoolNamesUniqueDup(t *testing.T) {
	c := config{
		WorkerPools: []workerPool{
			{
				Name: "dup",
			},
			{
				Name: "dup",
			},
		},
	}

	if d := c.checkWorkerPoolNamesUnique(); !d.HasErrors() {
		t.Error("Should fail with duplicated worker pool names")
	}
}

func TestCheckWorkerPoolNamesUniqueNotDup(t *testing.T) {
	c := config{
		WorkerPools: []workerPool{
			{
				Name: "not",
			},
			{
				Name: "dup",
			},
		},
	}

	if d := c.checkWorkerPoolNamesUnique(); d.HasErrors() {
		t.Error("Should work with no duplicated worker pool names")
	}
}

//nolint: funlen
func TestValidateOSVersion(t *testing.T) {
	type testCase struct {
		// Config to test
		cfg config
		// Expected output after running test
		hasError bool
	}

	cases := []testCase{
		{
			cfg: config{
				ClusterName: "c",
				WorkerPools: []workerPool{
					{
						Name:      "1",
						OSVersion: "current",
					},
				},
			},
			hasError: true,
		},
		{
			cfg: config{
				ClusterName: "c",
				OSVersion:   "current",
				WorkerPools: []workerPool{
					{
						Name: "2",
					},
				},
			},
			hasError: true,
		},
		{
			cfg: config{
				ClusterName: "c",
				WorkerPools: []workerPool{
					{
						Name: "3",
					},
				},
			},
			hasError: false,
		},
		{
			cfg: config{
				ClusterName: "c",
				WorkerPools: []workerPool{
					{
						Name:          "4",
						OSVersion:     "current",
						IPXEScriptURL: "https://demo.version",
					},
				},
			},
			hasError: false,
		},
		{
			cfg: config{
				ClusterName:   "c",
				OSVersion:     "current",
				IPXEScriptURL: "https://demo.version",
				WorkerPools: []workerPool{
					{
						Name: "5",
					},
				},
			},
			hasError: false,
		},
	}

	for tcIdx, tc := range cases {
		output := tc.cfg.validateOSVersion()
		if output.HasErrors() != tc.hasError {
			t.Errorf("In test %v, expected %v, got %v", tcIdx+1, tc.hasError, output.HasErrors())
		}
	}
}

//nolint:funlen
func TestCheckValidConfig(t *testing.T) {
	cases := map[string]struct {
		mutateF     func(*config)
		expectError bool
	}{
		"base_config_is_valid": {
			mutateF: func(*config) {},
		},
		"reservation_IDs_for_controller_nodes_can't_have_random_prefix": {
			mutateF: func(c *config) {
				c.ReservationIDs = map[string]string{"foo-1": "bar"}
			},
			expectError: true,
		},
		"reservation_IDs_for_controller_nodes_must_be_prefixed_with_'controller'": {
			mutateF: func(c *config) {
				c.ControllerCount = 1
				c.ReservationIDs = map[string]string{"controller-0": "bar"}
			},
		},
		"reservation_IDs_for_worker_nodes_can't_have_random_prefix": {
			mutateF: func(c *config) {
				c.WorkerPools[0].ReservationIDs = map[string]string{"foo-0": "bar"}
			},
			expectError: true,
		},
		"reservation_IDs_for_worker_nodes_must_be_prefixed_with_'worker'": {
			mutateF: func(c *config) {
				c.WorkerPools[0].Count = 1
				c.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "bar"}
			},
		},
		"reservation_IDs_for_worker_nodes_can't_be_mixed_default_reservation_ID": {
			mutateF: func(c *config) {
				c.WorkerPools[0].ReservationIDsDefault = "next-available"
				c.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "bar"}
			},
			expectError: true,
		},
		"reservation_IDs_for_worker_nodes_can't_be_set_to_'next-available'": {
			mutateF: func(c *config) {
				c.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "next-available"}
			},
			expectError: true,
		},
		"reservation_IDs_can't_be_empty": {
			mutateF: func(c *config) {
				c.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": ""}
			},
			expectError: true,
		},
		"reservation_IDs_must_be_sequential": {
			mutateF: func(c *config) {
				c.WorkerPools[0].Count = 2
				c.WorkerPools[0].ReservationIDs = map[string]string{
					"worker-0": "foo",
					"worker-1": "bar",
				}
			},
		},
		"not_sequential_reservation_IDs_are_invalid": {
			mutateF: func(c *config) {
				c.WorkerPools[0].ReservationIDs = map[string]string{
					"worker-0": "foo",
					"worker-2": "bar",
				}
			},
			expectError: true,
		},
		"negative_conntrack_max_per_core_is_invalid": {
			mutateF: func(c *config) {
				c.ConntrackMaxPerCore = -1
			},
			expectError: true,
		},
		"positive_conntrack_max_per_core_is_valid": {
			mutateF: func(c *config) {
				c.ConntrackMaxPerCore = 10
			},
			expectError: false,
		},
		"NodePrivateCIDR_both_fields_provided": {
			mutateF: func(c *config) {
				c.NodePrivateCIDR = "10.10.10.10"
				c.NodePrivateCIDRs = []string{"10.11.11.11"}
			},
			expectError: true,
		},
		"NodePrivateCIDR_repeated_CIDRs": {
			mutateF: func(c *config) {
				// Override the default value.
				c.NodePrivateCIDRs = nil
				c.NodePrivateCIDRs = []string{"10.10.10.10", "10.10.10.10"}
			},
			expectError: true,
		},

		"NodePrivateCIDR_no_node_private_cidr_given": {
			mutateF: func(c *config) {
				// Override the default value.
				c.NodePrivateCIDRs = nil
			},
			expectError: true,
		},
	}

	for name, c := range cases {
		c := c

		t.Run(name, func(t *testing.T) {
			config := baseConfig()

			c.mutateF(config)

			diagnostics := config.checkValidConfig()

			if diagnostics.HasErrors() && !c.expectError {
				t.Fatalf("unexpected validation error: %v", diagnostics)
			}

			if !diagnostics.HasErrors() && c.expectError {
				t.Fatalf("expected error, but validation passed")
			}
		})
	}
}

func baseConfig() *config {
	return &config{
		ClusterName: "c",
		WorkerPools: []workerPool{
			{
				Name: "1",
			},
			{
				Name: "2",
			},
		},
		NodePrivateCIDRs: []string{"11.11.11.11"},
	}
}

//nolint: funlen
func TestTerraformAddDeps(t *testing.T) {
	cases := map[string]struct {
		configF         func(*config)
		expectedConfigF func(*config)
	}{
		"worker pool without reservation IDs depends on worker pool with reservation ID": {
			func(c *config) {
				c.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "dummy"}
			},
			func(c *config) {
				c.WorkerPools[1].NodesDependOn = []string{poolTarget("1", "worker_nodes_ids")}
			},
		},
		"all worker pools without reservation IDs depends on worker pool with reservation ID": {
			func(c *config) {
				c.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "dummy"}
				c.WorkerPools = append(c.WorkerPools, workerPool{Name: "3"})
			},
			func(c *config) {
				c.WorkerPools[1].NodesDependOn = []string{poolTarget("1", "worker_nodes_ids")}
			},
		},
		"worker pool without reservation IDs depends on all worker pools with reservation IDs": {
			func(c *config) {
				c.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "dummy"}
				c.WorkerPools[1].ReservationIDs = map[string]string{"worker-0": "dummy"}
				c.WorkerPools = append(c.WorkerPools, workerPool{Name: "3"})
			},
			func(c *config) {
				c.WorkerPools[2].NodesDependOn = []string{
					poolTarget("1", "worker_nodes_ids"),
					poolTarget("2", "worker_nodes_ids"),
				}
			},
		},
		"worker pools without reservation IDs depends on controller nodes with reservation IDs": {
			func(c *config) {
				c.ReservationIDs = map[string]string{"controller-0": "dummy"}
			},
			func(c *config) {
				c.WorkerPools[0].NodesDependOn = []string{clusterTarget("1", "device_ids")}
				c.WorkerPools[1].NodesDependOn = []string{clusterTarget("1", "device_ids")}
			},
		},
		"worker pool without reservation IDs depends on controller nodes and worker pools with reservation IDs": {
			func(c *config) {
				c.ReservationIDs = map[string]string{"controller-0": "dummy"}
				c.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "dummy"}
			},
			func(c *config) {
				c.WorkerPools[1].NodesDependOn = []string{clusterTarget("1", "device_ids"), poolTarget("1", "device_ids")}
			},
		},
	}

	for name, c := range cases {
		c := c

		t.Run(name, func(t *testing.T) {
			// Create copy of base config.
			config := baseConfig()

			// Mutate it.
			c.configF(config)

			// Copy mutated config.
			expectedConfig := config

			// Mutate to expected config.
			c.expectedConfigF(expectedConfig)

			// Add dependencies.
			config.terraformAddDeps()

			for i, workerPool := range config.WorkerPools {
				dependencies := workerPool.NodesDependOn
				expectedDependencies := expectedConfig.WorkerPools[i].NodesDependOn

				if !reflect.DeepEqual(dependencies, expectedDependencies) {
					t.Fatalf("Expected %v, got %v", expectedDependencies, dependencies)
				}
			}
		})
	}
}

func Test_findDuplicateString(t *testing.T) {
	tests := []struct {
		name string
		strs []string
		want string
	}{
		{
			name: "repeated_strings",
			strs: []string{"1", "2", "3", "1"},
			want: "1",
		},
		{
			name: "unique_strings",
			strs: []string{"1", "2", "3"},
			want: "",
		},
		{
			name: "empty_input",
			strs: []string{},
			want: "",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := findDuplicateString(tt.strs); got != tt.want {
				t.Errorf("findDuplicateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolveNodePrivateCIDRs(t *testing.T) { //nolint:funlen
	tests := []struct {
		name    string
		cfg     *config
		want    []string
		wantErr bool
	}{
		{
			name: "both_fields_provided",
			cfg: &config{
				NodePrivateCIDR:  "10.10.10.10",
				NodePrivateCIDRs: []string{"10.11.11.11"},
			},
			wantErr: true,
		},
		{
			name: "repeated_CIDRs",
			cfg: &config{
				NodePrivateCIDRs: []string{"10.10.10.10", "10.10.10.10"},
			},
			wantErr: true,
		},
		{
			name:    "no_node_private_cidr_given",
			cfg:     &config{},
			wantErr: true,
		},
		{
			name: "only_node_private_cidr_given",
			cfg: &config{
				NodePrivateCIDR: "10.10.10.10",
			},
			want: []string{"10.10.10.10"},
		},
		{
			name: "only_node_private_cidrs_given",
			cfg: &config{
				NodePrivateCIDRs: []string{"10.10.10.10", "10.10.10.11"},
			},
			want: []string{"10.10.10.10", "10.10.10.11"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, diags := tt.cfg.resolveNodePrivateCIDRs()
			if diags.HasErrors() && !tt.wantErr || !diags.HasErrors() && tt.wantErr {
				t.Fatalf("got error: %v\nwantErr: %v", diags.Error(), tt.wantErr)
			}

			if diags.HasErrors() && tt.wantErr {
				t.Logf("Successfully failed with err: %v", diags.Error())

				return
			}

			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatalf("unexpected list -want +got)\n%s", diff)
			}
		})
	}
}
