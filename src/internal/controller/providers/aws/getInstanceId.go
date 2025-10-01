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

func getInstanceID(ctx context.Context, r record.EventRecorder, node *corev1.Node) (id string, err error) {
	log := logf.FromContext(ctx)
	// Getting instance ID
	var rx *regexp.Regexp
	if rx, err = regexp.Compile(`^aws:///[a-z0-9-]+/i-\w+$`); err != nil {
		log.V(1).Error(err, "regexp compile error")
		return
	}
	if !rx.MatchString(node.Spec.ProviderID) {
		log.V(1).Error(err, "failed to match node provider id")
		r.Eventf(node, corev1.EventTypeWarning, "UnknownProviderID", ".spec.providerId does not match %s", r)
		return
	}
	id = "i-" + strings.Split(node.Spec.ProviderID, "/i-")[1]
	return
}
