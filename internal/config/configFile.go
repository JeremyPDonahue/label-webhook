package config

import (
	"os"

	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type configFileStruct struct {
	AllowAdminNoMutate       bool             `yaml:"allow-admin-nomutate"`
	AllowAdminNoMutateToggle string           `yaml:"allow-admin-nomutate-toggle"`
	DockerhubRegistry        string           `yaml:"dockerhub-registry"`
	MutateIgnoredImages      []string         `yaml:"mutate-ignored-images"`
	CertificateAuthority     CertStruct       `yaml:"certificate-authority"`
	Certificate              CertStruct       `yaml:"certificate"`
	Kubernetes               KubernetesStruct `yaml:"kubernetes"`
}

type CertStruct struct {
	Certificate string `yaml:"certificate"`
	PrivateKey  string `yaml:"private-key"`
	PublicKey   string `yaml:"public-key"`
}

type KubernetesStruct struct {
	Namespace   string `yaml:"namespace"`
	ServiceName string `yaml:"service-name"`
}

func getConfigFileData(fileLocation string) (configFileStruct, error) {
	// does file exist
	if _, err := os.Stat(fileLocation); os.IsNotExist(err) {
		return configFileStruct{}, err
	}
	// read file
	rd, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return configFileStruct{}, err
	}
	// convert config file data to struct
	var output configFileStruct
	if err := yaml.Unmarshal(rd, &output); err != nil {
		return output, err
	}

	return output, nil
}
