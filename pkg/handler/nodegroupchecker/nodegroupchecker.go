// Package nodegroupchecker is the package that contains the check functions for node groups.
package nodegroupchecker

import (
	"context"
	"errors"

	"github.com/AlphaSense-Engineering/privatecloud-installer/pkg/handler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	// errNoNodesWithGPULabel is the error that is returned when no nodes with the GPU label are found.
	errNoNodesWithGPULabel = errors.New("no nodes with GPU label found")
)

const (
	// LogMsgNodeGroupsChecked is the message that is logged when the node groups are checked.
	LogMsgNodeGroupsChecked = "checked node groups"

	// LogMsgNodeGroupsCheckedWithError is the message that is logged when the node groups are checked with an error.
	LogMsgNodeGroupsCheckedWithError = "checked node groups; %s"
)

// NodeGroupChecker is the type that contains the node group check functions.
type NodeGroupChecker struct {
	// clientset is the Kubernetes client.
	clientset kubernetes.Interface
}

var _ handler.Handler = &NodeGroupChecker{}

// Handle is the function that handles the node group checking.
//
// The arguments are not used.
// It returns nothing on success, or an error on failure.
func (c *NodeGroupChecker) Handle(ctx context.Context, _ ...any) ([]any, error) {
	const (
		// labelType is the label that is used to identify the node type.
		labelType = "type"

		// typeGPU is the type of node that has GPUs.
		typeGPU = "gpu"
	)

	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	for _, node := range nodes.Items {
		if label, ok := node.Labels[labelType]; ok && label == typeGPU {
			return nil, nil
		}
	}

	return nil, errNoNodesWithGPULabel
}

// New is the function that creates a new NodeGroupChecker.
func New(clientset kubernetes.Interface) *NodeGroupChecker {
	return &NodeGroupChecker{
		clientset: clientset,
	}
}
