// Package envconfig is the package that implements the environment configuration type.
package envconfig

import (
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/cloud"
	selferrors "github.com/AlphaSense-Engineering/privatecloud-installer/pkg/errors"
	"gopkg.in/yaml.v3"
)

// errNoEnvConfigKindFound is the error that is returned when no environment configuration kind is found in the YAML file.
var errNoEnvConfigKindFound = errors.New("no environment configuration kind found in the YAML file")

// AWSSpec is the type that represents the AWS cloud specification of the environment configuration.
type AWSSpec struct {
	// AccountID is the AWS account ID.
	AccountID string `yaml:"accountID"`

	// OIDCURL is the OIDC URL.
	OIDCURL string `yaml:"oidcUrl"`
}

// AzureSpec is the type that represents the Azure cloud specification of the environment configuration.
type AzureSpec struct {
	// ClientID is the Azure client ID.
	ClientID string `yaml:"clientID"`
	// ResourceGroup is the Azure resource group.
	ResourceGroup string `yaml:"resourceGroup"`
	// SubscriptionID is the Azure subscription ID.
	SubscriptionID string `yaml:"subscriptionID"`
	// TenantID is the Azure tenant ID.
	TenantID string `yaml:"tenantID"`

	// OIDCURL is the OIDC URL.
	OIDCURL string `yaml:"oidcUrl"`
}

// GCPSpec is the type that represents the GCP cloud specification of the environment configuration.
type GCPSpec struct {
	// ProjectID is the GCP project ID.
	ProjectID string `yaml:"projectID"`
	// ProjectNumber is the GCP project number.
	ProjectNumber string `yaml:"projectNumber"`
}

// CloudSpec is the type that represents the cloud specification of the environment configuration.
type CloudSpec struct {
	// CloudZone is the cloud zone.
	CloudZone string `yaml:"cloudZone"`
	// Provider is the cloud provider.
	Provider string `yaml:"provider"`

	// AWS is the AWS cloud specification.
	AWS *AWSSpec `yaml:"aws,omitempty"`
	// Azure is the Azure cloud specification.
	Azure *AzureSpec `yaml:"azure,omitempty"`
	// GCP is the GCP cloud specification.
	GCP *GCPSpec `yaml:"gcp,omitempty"`
}

// Spec is the type that represents the specification of the environment configuration.
type Spec struct {
	// ClientID is the client ID.
	ClientID string `yaml:"clientID"`
	// ClusterName is the cluster name.
	ClusterName string `yaml:"clusterName"`
	// DomainName is the domain name.
	DomainName string `yaml:"domainName"`
	// InstallID is the install ID.
	InstallID string `yaml:"installID"`
	// Version is the version.
	Version string `yaml:"version"`

	// CloudSpec is the cloud specification.
	CloudSpec CloudSpec `yaml:"cloudSpec"`
}

// EnvConfig is the type that represents the environment configuration.
type EnvConfig struct {
	// Kind is a string value representing the REST resource this object represents.
	// Servers may infer this from the endpoint the client submits requests to.
	// Cannot be updated.
	// In CamelCase.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	Kind string `json:"kind,omitempty"`

	// APIVersion defines the versioned schema of this representation of an object.
	// Servers should convert recognized schemas to the latest internal value, and
	// may reject unrecognized values.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources
	APIVersion string `json:"apiVersion,omitempty"`

	// Name must be unique within a namespace. Is required when creating resources, although
	// some resources may allow a client to request the generation of an appropriate name
	// automatically. Name is primarily intended for creation idempotence and configuration
	// definition.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#names
	Name string `json:"name,omitempty"`

	// Namespace defines the space within which each name must be unique. An empty namespace is
	// equivalent to the "default" namespace, but "default" is the canonical representation.
	// Not all objects are required to be scoped to a namespace - the value of this field for
	// those objects will be empty.
	//
	// Must be a DNS_LABEL.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces
	Namespace string `json:"namespace,omitempty"`

	// Spec is the specification.
	Spec Spec `yaml:"spec"`
}

// OIDCURL returns the OIDC URL.
func (e *EnvConfig) OIDCURL() string {
	switch v := cloud.Cloud(e.Spec.CloudSpec.Provider); v {
	case cloud.AWS:
		return e.Spec.CloudSpec.AWS.OIDCURL
	case cloud.Azure:
		return e.Spec.CloudSpec.Azure.OIDCURL
	default:
		panic(selferrors.NewUnsupportedCloud(v))
	}
}

// NewFromBytes returns a new EnvConfig from the given bytes.
func NewFromBytes(data []byte) (*EnvConfig, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		// envConfigKind is the kind of the environment configuration.
		const envConfigKind = "EnvConfig"

		var envConfig EnvConfig

		if err := decoder.Decode(&envConfig); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		if envConfig.Kind == envConfigKind {
			return &envConfig, nil
		}
	}

	return nil, errNoEnvConfigKindFound
}

// NewFromPath returns a new EnvConfig from the given path.
func NewFromPath(path string) (*EnvConfig, error) {
	yamlFile, err := os.ReadFile(path) // nolint:gosec
	if err != nil {
		return nil, err
	}

	return NewFromBytes(yamlFile)
}
