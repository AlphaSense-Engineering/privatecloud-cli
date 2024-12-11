// Package cloud is the package that contains the cloud definitions.
package cloud

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
