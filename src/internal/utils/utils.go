// Package utils provides utility functions and common helpers.
package utils

import (
	"errors"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

// RequeueResult after result object with a default timeout
var RequeueResult = ctrl.Result{RequeueAfter: 10 * time.Second}

var ErrRequestRequeue = errors.New("requeue")

func CheckRequeue(err error) (toRequeue bool) {
	return err.Error() == "requeue"
}
