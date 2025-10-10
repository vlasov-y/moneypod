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
