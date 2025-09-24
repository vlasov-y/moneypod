package controller

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

// Requeue after result object with a default timeout
var requeue = ctrl.Result{RequeueAfter: 10 * time.Second}
