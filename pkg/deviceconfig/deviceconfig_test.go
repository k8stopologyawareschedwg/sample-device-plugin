package deviceconfig

import (
	"github.com/google/go-cmp/cmp"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name         string
		conf         []byte
		resName      string
		confFileName string
		expected     *NodesDevices
	}{
		{
			name: "valid config without wildcard",
			conf: []byte(
				`
devices:
   kind-control-plane:
     - id: DevA1
       healthy: true
       numanode: 0
     - id: DevA2
       healthy: true
       numanode: 1
     - id: DevA3
       healthy: false
       numanode: 1
`),
			resName:      "test/dev",
			confFileName: "test_dev.yaml",
			expected: &NodesDevices{
				Devices: map[string][]SampleDevice{
					"kind-control-plane": {
						{
							ID:       "DevA1",
							Healthy:  true,
							NUMANode: 0,
						},
						{
							ID:       "DevA2",
							Healthy:  true,
							NUMANode: 1,
						},
						{
							ID:       "DevA3",
							Healthy:  false,
							NUMANode: 1,
						},
					},
				},
			},
		},
		{
			name: "valid config with wildcard",
			conf: []byte(
				`
devices:
  '*':
    - id: DevA1
      healthy: true
      numanode: 0
    - id: DevA2
      healthy: true
      numanode: 1
    - id: DevA3
      healthy: true
      numanode: 1
`),
			resName:      "prefix/test.dev",
			confFileName: "prefix_test_dev.yaml",
			expected: &NodesDevices{
				Devices: map[string][]SampleDevice{
					"*": {
						{
							ID:       "DevA1",
							Healthy:  true,
							NUMANode: 0,
						},
						{
							ID:       "DevA2",
							Healthy:  true,
							NUMANode: 1,
						},
						{
							ID:       "DevA3",
							Healthy:  true,
							NUMANode: 1,
						},
					},
				},
			},
		},
		{
			name: "invalid config",
			conf: []byte(
				`
devices:
  '*':
    - fail: DevA1
      wrong: true
`),
			resName:      "test/fail/dev",
			confFileName: "test_fail_dev.yaml",
			expected: &NodesDevices{
				Devices: map[string][]SampleDevice{"*": {{}}}},
		},
	}

	confDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to open temporary dir")
	}

	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatalf("failed to delete dir %q", path)
		}
	}(confDir)

	for _, tc := range testCases {
		t.Logf("test %s", tc.name)
		fullFilePath := filepath.Join(confDir, tc.confFileName)
		if err := ioutil.WriteFile(fullFilePath, tc.conf, 0644); err != nil {
			t.Fatalf("failed to write to %q", fullFilePath)
		}
		devices, err := Parse(confDir, tc.resName)
		if err != nil {
			t.Errorf("failed to parse conf file %q with resource name %q; error: %v", fullFilePath, tc.resName, err)
		}
		if diff := cmp.Diff(tc.expected, devices); diff != "" {
			t.Errorf("configuration mismatch; diff %v", diff)
		}
	}
}

func TestGenerate(t *testing.T) {
	testCases := []struct {
		name     string
		count    int
		expected *NodesDevices
	}{
		{
			name:  "no devices",
			count: 0,
			expected: &NodesDevices{
				Devices: map[string][]SampleDevice{
					"*": {},
				},
			},
		},
		{
			name:  "typical devices amount",
			count: 3,
			expected: &NodesDevices{
				Devices: map[string][]SampleDevice{
					"*": {
						{
							ID:       "Dev-0",
							Healthy:  true,
							NUMANode: -1,
						},
						{
							ID:       "Dev-1",
							Healthy:  true,
							NUMANode: -1,
						},
						{
							ID:       "Dev-2",
							Healthy:  true,
							NUMANode: -1,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Generate(tc.count)
			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("generation mismatch; diff %v", diff)
			}
		})
	}
}
