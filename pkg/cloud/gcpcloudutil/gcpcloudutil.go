// Package gcpcloudutil is the package that contains the GCP cloud utility functions.
package gcpcloudutil

import "fmt"

// ServiceAccountAnnotation is a function that returns the annotation for the service account.
func ServiceAccountAnnotation(clusterName string, projectID string) string {
	return fmt.Sprintf("uxp-provider-%s@%s.iam.gserviceaccount.com", clusterName, projectID)
}
