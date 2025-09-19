package controller

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	annotationHourlyCost     = "moneypod.io/hourly-cost"
	annotationNodeHourlyCost = "moneypod.io/node-hourly-cost"
	maxConcurrentReconciles  = 1
)

// Requeue after result object with a default timeout
var requeue = ctrl.Result{RequeueAfter: 10 * time.Second}
