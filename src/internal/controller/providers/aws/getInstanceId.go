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

// Package aws provides AWS-specific functionality for the controller.
package aws

import (
	"context"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (*Provider) getInstanceID(ctx context.Context, r record.EventRecorder, node *corev1.Node) (id string, err error) {
	log := logf.FromContext(ctx)
	// Getting instance ID
	var rx *regexp.Regexp
	if rx, err = regexp.Compile(`^aws:///[a-z0-9-]+/i-\w+$`); err != nil {
		log.Error(err, "regexp compile error")
		return
	}
	if !rx.MatchString(node.Spec.ProviderID) {
		log.Error(err, "failed to match node provider id")
		r.Eventf(node, corev1.EventTypeWarning, "UnknownProviderID", ".spec.providerId does not match %s", r)
		return
	}
	id = "i-" + strings.Split(node.Spec.ProviderID, "/i-")[1]
	return
}
