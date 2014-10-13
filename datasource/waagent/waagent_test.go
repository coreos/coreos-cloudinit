package waagent

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type mockFilesystem map[string][]byte

func (m mockFilesystem) readFile(filename string) ([]byte, error) {
	if contents := m[filename]; contents != nil {
		return contents, nil
	}
	return nil, os.ErrNotExist
}

func TestFetchMetadata(t *testing.T) {
	for _, tt := range []struct {
		root     string
		files    mockFilesystem
		metadata map[string]string
	}{
		{
			"/",
			mockFilesystem{},
			nil,
		},
		{
			"/",
			mockFilesystem{"/SharedConfig.xml": []byte("")},
			nil,
		},
		{
			"/var/lib/waagent",
			mockFilesystem{"/var/lib/waagent/SharedConfig.xml": []byte("")},
			nil,
		},
		{
			"/var/lib/waagent",
			mockFilesystem{"/var/lib/waagent/SharedConfig.xml": []byte(`<?xml version="1.0" encoding="utf-8"?>
<SharedConfig version="1.0.0.0" goalStateIncarnation="1">
  <Deployment name="c8f9e4c9c18948e1bebf57c5685da756" guid="{1d10394f-c741-4a1a-a6bb-278f213c5a5e}" incarnation="0" isNonCancellableTopologyChangeEnabled="false">
    <Service name="core-test-1" guid="{00000000-0000-0000-0000-000000000000}" />
    <ServiceInstance name="c8f9e4c9c18948e1bebf57c5685da756.0" guid="{1e202e9a-8ffe-4915-b6ef-4118c9628fda}" />
  </Deployment>
  <Incarnation number="1" instance="core-test-1" guid="{8767eb4b-b445-4783-b1f5-6c0beaf41ea0}" />
  <Role guid="{53ecc81e-257f-fbc9-a53a-8cf1a0a122b4}" name="core-test-1" settleTimeSeconds="0" />
  <LoadBalancerSettings timeoutSeconds="0" waitLoadBalancerProbeCount="8">
    <Probes>
      <Probe name="D41D8CD98F00B204E9800998ECF8427E" />
      <Probe name="C9DEC1518E1158748FA4B6081A8266DD" />
    </Probes>
  </LoadBalancerSettings>
  <OutputEndpoints>
    <Endpoint name="core-test-1:openInternalEndpoint" type="SFS">
      <Target instance="core-test-1" endpoint="openInternalEndpoint" />
    </Endpoint>
  </OutputEndpoints>
  <Instances>
    <Instance id="core-test-1" address="100.73.202.64">
      <FaultDomains randomId="0" updateId="0" updateCount="0" />
      <InputEndpoints>
        <Endpoint name="openInternalEndpoint" address="100.73.202.64" protocol="any" isPublic="false" enableDirectServerReturn="false" isDirectAddress="false" disableStealthMode="false">
          <LocalPorts>
            <LocalPortSelfManaged />
          </LocalPorts>
        </Endpoint>
        <Endpoint name="ssh" address="100.73.202.64:22" protocol="tcp" hostName="core-test-1ContractContract" isPublic="true" loadBalancedPublicAddress="191.239.39.77:22" enableDirectServerReturn="false" isDirectAddress="false" disableStealthMode="false">
          <LocalPorts>
            <LocalPortRange from="22" to="22" />
          </LocalPorts>
        </Endpoint>
      </InputEndpoints>
    </Instance>
  </Instances>
</SharedConfig>`)},
			map[string]string{
				"local-ipv4":  "100.73.202.64",
				"public-ipv4": "191.239.39.77",
			},
		},
	} {
		a := waagent{tt.root, tt.files.readFile}
		metadataBytes, err := a.FetchMetadata()
		if err != nil {
			t.Fatalf("bad error for %q: want %q, got %q", tt, nil, err)
		}
		var metadata map[string]string
		if len(metadataBytes) > 0 {
			if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
				panic(err)
			}
		}
		if !reflect.DeepEqual(tt.metadata, metadata) {
			t.Fatalf("bad metadata for %q: want %q, got %q", tt, tt.metadata, metadata)
		}
	}
}

func TestFetchUserdata(t *testing.T) {
	for _, tt := range []struct {
		root  string
		files mockFilesystem
	}{
		{
			"/",
			mockFilesystem{},
		},
		{
			"/",
			mockFilesystem{"/CustomData": []byte{}},
		},
		{
			"/var/lib/waagent/",
			mockFilesystem{"/var/lib/waagent/CustomData": []byte{}},
		},
	} {
		a := waagent{tt.root, tt.files.readFile}
		_, err := a.FetchUserdata()
		if err != nil {
			t.Fatalf("bad error for %q: want %q, got %q", tt, nil, err)
		}
	}
}

func TestConfigRoot(t *testing.T) {
	for _, tt := range []struct {
		root       string
		configRoot string
	}{
		{
			"/",
			"/",
		},
		{
			"/var/lib/waagent",
			"/var/lib/waagent",
		},
	} {
		a := waagent{tt.root, nil}
		if configRoot := a.ConfigRoot(); configRoot != tt.configRoot {
			t.Fatalf("bad config root for %q: want %q, got %q", tt, tt.configRoot, configRoot)
		}
	}
}

func TestNewDatasource(t *testing.T) {
	for _, tt := range []struct {
		root       string
		expectRoot string
	}{
		{
			root:       "",
			expectRoot: "",
		},
		{
			root:       "/var/lib/waagent",
			expectRoot: "/var/lib/waagent",
		},
	} {
		service := NewDatasource(tt.root)
		if service.root != tt.expectRoot {
			t.Fatalf("bad root (%q): want %q, got %q", tt.root, tt.expectRoot, service.root)
		}
	}
}
