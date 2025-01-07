// Package storageclasschecker is the package that contains the check functions for the storage class.
package storageclasschecker

import (
	"context"
	"errors"
	"strconv"

	"github.com/AlphaSense-Engineering/privatecloud-cli/pkg/handler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// errNoDefaultStorageClass is the error that is returned when no default storage class is found.
var errNoDefaultStorageClass = errors.New("no default storage class found")

// StorageClassChecker is the type that contains the check functions for the storage class.
type StorageClassChecker struct {
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface
}

var _ handler.Handler = &StorageClassChecker{}

// Handle is the function that handles the storage class checking.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
func (c *StorageClassChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	// defaultStorageClassAnnotation is the annotation that is used to determine if a storage class is the default.
	const defaultStorageClassAnnotation = "storageclass.kubernetes.io/is-default-class"

	storageClasses, err := c.clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, sc := range storageClasses.Items {
		if sc.Annotations[defaultStorageClassAnnotation] == strconv.FormatBool(true) {
			return nil, nil
		}
	}

	return nil, errNoDefaultStorageClass
}

// New is a function that returns a new StorageClassChecker.
func New(clientset kubernetes.Interface) *StorageClassChecker {
	return &StorageClassChecker{clientset: clientset}
}
