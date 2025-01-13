// Package gcpcloudutil is the package that contains the GCP cloud utility functions.
package gcpcloudutil

import "fmt"

// ServiceAccountAnnotationKey is the key for the GCP service account annotation.
const ServiceAccountAnnotationKey = "iam.gke.io/gcp-service-account"

// ServiceAccountAnnotation is a function that returns the annotation for the service account.
func ServiceAccountAnnotation(clusterName string, projectID string) string {
	return fmt.Sprintf("uxp-provider-%s@%s.iam.gserviceaccount.com", clusterName, projectID)
}
