package config

import (
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	cfg := Init()
	if cfg.AllowAdminNoMutateToggle != "7b068a99-c02b-410a-bd59-3514bac85e7a" {
		t.Errorf("Init() returned incorrect value for AllowAdminNoMutateToggle, got %v, wanted %v", cfg.AllowAdminNoMutateToggle, "7b068a99-c02b-410a-bd59-3514bac85e7a")
	}
}

func TestGetOSEnv(t *testing.T) {
	expected := "foo"
	os.Setenv("TEST_ENV", expected)
	result := getOSEnv("TEST_ENV", "bar")

	if result != expected {
		t.Errorf("getOSEnv() returned incorrect value, got %s, wanted %s", result, expected)
	}

	os.Unsetenv("TEST_ENV")
	result = getOSEnv("TEST_ENV", expected)
	if result != expected {
		t.Errorf("getOSEnv() returned incorrect value, got %s, wanted %s", result, expected)
	}
}

func TestUpdateValues(t *testing.T) {
	cfg := Config{
		AllowAdminNoMutate:       false,
		AllowAdminNoMutateToggle: "7b068a99-c02b-410a-bd59-3514bac85e7a",
		DockerhubRegistry:        "registry.hub.docker.com",
		NameSpace:                "ingress-nginx",
	}
	cfgFile := configFileStruct{
		AllowAdminNoMutate:       true,
		AllowAdminNoMutateToggle: "test-token",
		DockerhubRegistry:        "registry.example.com",
		Kubernetes: KubernetesStruct{
			Namespace: "example-namespace",
		},
	}

	updateValues(&cfg, cfgFile)
	if cfg.AllowAdminNoMutate != cfgFile.AllowAdminNoMutate {
		t.Errorf("updateValues() returned incorrect value for AllowAdminNoMutate, got %v, wanted %v", cfg.AllowAdminNoMutate, cfgFile.AllowAdminNoMutate)
	}
	if cfg.AllowAdminNoMutateToggle != cfgFile.AllowAdminNoMutateToggle {
		t.Errorf("updateValues() returned incorrect value for AllowAdminNoMutateToken, got %v, wanted %v", cfg.AllowAdminNoMutateToggle, cfgFile.AllowAdminNoMutateToggle)
	}
	if cfg.DockerhubRegistry != cfgFile.DockerhubRegistry {
		t.Errorf("updateValues() returned incorrect value for DockerhubRegistry, got %v, wanted %v", cfg.DockerhubRegistry, cfgFile.DockerhubRegistry)
	}
	if cfg.NameSpace != cfgFile.Kubernetes.Namespace {
		t.Errorf("updateValues() returned incorrect value for NameSpace, got %v, wanted %v", cfg.NameSpace, cfgFile.Kubernetes.Namespace)
	}
	/*
		if cfg.ServiceName == "webhook" && configFileData.Kubernetes.ServiceName != "webhook" {
			cfg.ServiceName = configFileData.Kubernetes.ServiceName
		}
		if len(configFileData.MutateIgnoredImages) != 0 {
			cfg.MutateIgnoredImages = configFileData.MutateIgnoredImages
		}
		if len(configFileData.CertificateAuthority.Certificate) != 0 {
			cfg.CACert = configFileData.CertificateAuthority.Certificate
		}
		if len(configFileData.CertificateAuthority.PrivateKey) != 0 {
			cfg.CAPrivateKey = configFileData.CertificateAuthority.PrivateKey
		}
		if len(configFileData.Certificate.Certificate) != 0 {
			cfg.CertCert = configFileData.Certificate.Certificate
		}
		if len(configFileData.Certificate.PrivateKey) != 0 {
			cfg.CertPrivateKey = configFileData.Certificate.PrivateKey
		}
	*/

}

func TestGetDNSNames(t *testing.T) {
	expected := []string{
		"exampleService",
		"exampleService.exampleNameSpace",
		"exampleService.exampleNameSpace.svc",
		"exampleService.exampleNameSpace.svc.cluster",
		"exampleService.exampleNameSpace.svc.cluster.local",
	}
	result := getDNSNames("exampleService", "exampleNameSpace")
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("getDNSNames() failed to return the expected result, got %s, wanted %s", expected[i], result[i])
		}
	}
	return
}
