// Copyright 2025 The MoneyPod Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package providers provides an interface to implement by all cloud proviers for getting costs and information about the nodes.
package providers

import (
	"context"
	"strings"

	"github.com/vlasov-y/moneypod/internal/providers/aws"
	"github.com/vlasov-y/moneypod/internal/providers/manual"
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
