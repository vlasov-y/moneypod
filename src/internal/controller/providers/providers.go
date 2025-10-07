// Package providers provides an interface to implement by all cloud proviers for getting costs and information about the nodes.
package providers

import (
	"context"
	"strings"

	"github.com/vlasov-y/moneypod/internal/controller/providers/aws"
	"github.com/vlasov-y/moneypod/internal/controller/providers/manual"
	"github.com/vlasov-y/moneypod/internal/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

type Provider interface {
	GetNodeHourlyCost(ctx context.Context, r record.EventRecorder, node *corev1.Node) (hourlyCost float64, err error)
	GetNodeInfo(ctx context.Context, r record.EventRecorder, node *corev1.Node) (info types.NodeInfo, err error)
}

func NewProvider(node *corev1.Node) (provider Provider) {
	if strings.HasPrefix(node.Spec.ProviderID, "aws://") {
		return &aws.Provider{}
	}
	return &manual.Provider{}
}
