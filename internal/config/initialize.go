package config

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"mutating-webhook/internal/certificate"
	"os"
	"reflect"
	"strings"
	"time"
)

func getOSEnv(env, def string) string {
	if val, ok := os.LookupEnv(strings.ToUpper(env)); ok {
		return val
	}
	return def
}

// Init initializes the application configuration by reading default values from the struct's tags
// and environment variables. Tags processed by this process are as follows:
// `ignored:"true" env:"ENVIRONMENT_VARIABLE" default:"default value"`
func Init() Config {
	cfg := DefaultConfig()

	cfgInfo, err := getStructInfo(&cfg)
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}

	for _, info := range cfgInfo {
		switch info.Type.String() {
		case "string":
			var dv string
			if info.DefaultValue != nil {
				dv = info.DefaultValue.(string)
			}
			p := reflect.ValueOf(&cfg).Elem().FieldByName(info.Name).Addr().Interface().(*string)
			flag.StringVar(p, info.Name, dv, "("+info.Key+")")

		case "bool":
			var dv bool
			if info.DefaultValue != nil {
				dv = info.DefaultValue.(bool)
			}
			p := reflect.ValueOf(&cfg).Elem().FieldByName(info.Name).Addr().Interface().(*bool)
			flag.BoolVar(p, info.Name, dv, "("+info.Key+")")

		case "int":
			var dv int
			if info.DefaultValue != nil {
				dv = int(info.DefaultValue.(int64))
			}
			p := reflect.ValueOf(&cfg).Elem().FieldByName(info.Name).Addr().Interface().(*int)
			flag.IntVar(p, info.Name, dv, "("+info.Key+")")
		}
	}
	flag.Parse()

	// set logging level
	setLogLevel(cfg)

	// timezone & format configuration
	cfg.TZoneUTC, _ = time.LoadLocation("UTC")
	if err != nil {
		log.Fatalf("[ERROR] Unable to parse timezone string. Please use one of the timezone database values listed here: %s", "https://en.wikipedia.org/wiki/List_of_tz_database_time_zones")
	}
	cfg.TZoneLocal, err = time.LoadLocation(cfg.TimeZoneLocal)
	if err != nil {
		log.Fatalf("[ERROR] Unable to parse timezone string. Please use one of the timezone database values listed here: %s", "https://en.wikipedia.org/wiki/List_of_tz_database_time_zones")
	}
	time.Now().Format(cfg.TimeFormat)

	// read config file
	configFileData, err := getConfigFileData(cfg.ConfigFile)
	if err != nil {
		log.Fatalf("[FATAL] Unable to read configuration file: %v", err)
	}
	updateValues(&cfg, configFileData)

	// Generate certificates if needed
	if err := certificateInit(&cfg); err != nil {
		log.Fatalf("[FATAL] Unable to initialize certificate data: %v", err)
	}

	// print running config
	printRunningConfig(&cfg, cfgInfo)

	log.Println("[INFO] initialization sequence complete")
	return cfg
}

func updateValues(cfg *Config, configFileData configFileStruct) {
	if cfg.AllowAdminNoMutate == false && configFileData.AllowAdminNoMutate != false {
		cfg.AllowAdminNoMutate = configFileData.AllowAdminNoMutate
	}
	if cfg.NameSpace == "openshift-webhook" && configFileData.Kubernetes.Namespace != "openshift-webhook" {
		cfg.NameSpace = configFileData.Kubernetes.Namespace
	}
	if cfg.ServiceName == "custom-labels-webhook" && configFileData.Kubernetes.ServiceName != "custom-labels-webhook" {
		cfg.ServiceName = configFileData.Kubernetes.ServiceName
	}
	if len(configFileData.ExcludedNamespaces) != 0 {
		cfg.ExcludedNamespaces = configFileData.ExcludedNamespaces
	}
	if len(configFileData.CustomLabels) != 0 {
		cfg.CustomLabels = configFileData.CustomLabels
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
}

func certificateInit(cfg *Config) error {
	// certificate authority private key does not exist, generate key pair
	if len(cfg.CAPrivateKey) == 0 {
		log.Printf("[TRACE] No certificate authority private key detected")
		keyPair, err := certificate.CreateRSAKeyPair(4096)
		if err != nil {
			return fmt.Errorf("Create RSA Key (%v)", err)
		}
		// pem encode private key
		k := new(bytes.Buffer)
		pem.Encode(k, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(keyPair),
		})
		cfg.CAPrivateKey = k.String()
	}

	// certificate authority certificate is missing, create it
	if len(cfg.CACert) == 0 {
		log.Printf("[TRACE] No certificate authority certificate detected")
		caCert, err := certificate.CreateCA(cfg.CAPrivateKey)
		if err != nil {
			return fmt.Errorf("Create CA (%v)", err)
		}
		cfg.CACert = caCert
	}

	// certificate private key does not exist, generate key pair
	if len(cfg.CertPrivateKey) == 0 {
		log.Printf("[TRACE] No server private key detected")
		keyPair, err := certificate.CreateRSAKeyPair(4096)
		if err != nil {
			return fmt.Errorf("Create RSA Key (%v)", err)
		}
		// pem encode private key
		k := new(bytes.Buffer)
		pem.Encode(k, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(keyPair),
		})
		cfg.CertPrivateKey = k.String()
	}

	// certificate certificate is missing, create it
	if len(cfg.CertCert) == 0 {
		log.Printf("[TRACE] No server certificate detected")
		csr, err := certificate.CreateCSR(cfg.CertPrivateKey, getDNSNames(cfg.ServiceName, cfg.NameSpace,))
		if err != nil {
			return fmt.Errorf("Create CSR (%v)", err)
		}
		cert, err := certificate.SignCert(cfg.CACert, cfg.CAPrivateKey, csr)
		if err != nil {
			return fmt.Errorf("Sign Cert (%v)", err)
		}
		cfg.CertCert = cert
	}

	return nil
}

func getDNSNames(service, ns string) []string {
	return []string{
		fmt.Sprintf("%s", service),
		fmt.Sprintf("%s.%s", service, ns),
		fmt.Sprintf("%s.%s.svc", service, ns),
		fmt.Sprintf("%s.%s.svc.cluster", service, ns),
		fmt.Sprintf("%s.%s.svc.cluster.local", service, ns),
	}
}
