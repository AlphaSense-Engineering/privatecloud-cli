// Package cloud is the package that contains the cloud definitions.
package cloud

import "fmt"

// Cloud represents the cloud provider.
type Cloud string

const (
	// AWS is the constant that represents the AWS cloud.
	AWS Cloud = "aws"

	// Azure is the constant that represents the Azure cloud.
	Azure Cloud = "azure"

	// GCP is the constant that represents the GCP cloud.
	GCP Cloud = "gcp"
)

// CrossplaneRoleNameSuffix is the suffix of the Crossplane role name.
const CrossplaneRoleNameSuffix = "crossplane-provider"

// UnsupportedCloudError is the error that is returned when the cloud is unsupported.
type UnsupportedCloudError struct {
	// cloud is the cloud that is unsupported.
	cloud Cloud
}

var _ error = &UnsupportedCloudError{}

// Error is a function that returns the error message.
func (e *UnsupportedCloudError) Error() string {
	return fmt.Sprintf("unsupported cloud type: %s", e.cloud)
}

// NewUnsupportedCloudError is a function that returns a new unsupported cloud error.
func NewUnsupportedCloudError(cloud Cloud) error {
	return &UnsupportedCloudError{cloud: cloud}
}
