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

package manual

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// If annotation is a selector for node or label - unpack
func (*Provider) parseAnnotationLabelSelector(node *corev1.Node, value string) (result string, err error) {
	var m map[string]string
	var selector string
	var selectorType string

	if strings.HasPrefix(value, "label=") {
		m = node.GetLabels()
		selector = strings.Split(value, "label=")[1]
		selectorType = "label"
	} else if strings.HasPrefix(value, "annotation=") {
		m = node.GetAnnotations()
		selector = strings.Split(value, "annotation=")[1]
		selectorType = "annotation"
	} else {
		// If there is no selector prefix - return value as it is, since it is a literal
		return value, err
	}

	if m == nil {
		m = map[string]string{}
	}

	// Try find referenced label or annotation
	var exists bool
	if value, exists = m[selector]; !exists {
		return value, fmt.Errorf("could not find %s %s", selectorType, selector)
	}
	return
}
