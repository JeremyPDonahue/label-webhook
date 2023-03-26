package config

import (
	"testing"
)

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
