/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/vlasov-y/moneypod/test/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespace = "moneypod"
)

// // namespace where the project is deployed in
// const namespace = "moneypod"

// // serviceAccountName created for the project
// const serviceAccountName = "moneypod-app"

// // metricsServiceName is the name of the metrics service of the project
// const metricsServiceName = "moneypod-metrics"

// // metricsRoleBindingName is the name of the RBAC that will be created to allow get the metrics data
// const metricsRoleBindingName = "moneypod-app"

var _ = Describe("Manager", Ordered, func() {
	var podName string
	var c client.Client
	var rc *rest.Config
	var err error
	var cmd *exec.Cmd

	// Before running the tests, set up the environment by creating the namespace,
	// enforce the restricted security policy to the namespace, installing CRDs,
	// and deploying the controller.
	BeforeAll(func() {
		By("Installing operator to the cluster")
		cmd = exec.Command("task", "install-operator")
		_, err = utils.Run(cmd)
		ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to install the operator")

		By("Initializing Kubernetes client")
		var cwd string
		cwd, err = utils.GetProjectDir()
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to get project dir")
		rc, err = clientcmd.BuildConfigFromFlags("", path.Join(cwd, "kubeconfig.yaml"))
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to load kubeconfig")
		c, err = client.New(rc, client.Options{})
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to create the client")
		pods := &corev1.PodList{}
		err = c.List(context.Background(), pods, client.InNamespace(namespace))
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to list pods in the namespace")
		Expect(len(pods.Items)).To(Equal(1), "expected eactly 1 pod in the namespace")
		podName = pods.Items[0].Name
	})

	// After all tests have been executed, clean up by undeploying the controller, uninstalling CRDs,
	// and deleting the namespace.
	AfterAll(func() {
		By("Uninstalling the operator")
		cmd = exec.Command("task", "uninstall-operator")
		_, err = utils.Run(cmd)
		ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to uninstall the operator")
	})

	// After each test, check for failures and collect logs, events,
	// and pod descriptions for debugging.
	AfterEach(func() {
		specReport := CurrentSpecReport()
		if specReport.Failed() {
			clientset := kubernetes.NewForConfigOrDie(rc)

			By("Fetching controller manager pod logs")
			logs, err := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{}).DoRaw(context.Background())
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Controller logs:\n %s", string(logs))
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Controller logs: %s", err)
			}

			By("Fetching Kubernetes events")
			events := &corev1.EventList{}
			err = c.List(context.Background(), events, client.InNamespace(namespace))
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Kubernetes events:\n")
				for _, event := range events.Items {
					_, _ = fmt.Fprintf(GinkgoWriter, "%s %s %s\n", event.LastTimestamp.Time, event.Reason, event.Message)
				}
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Kubernetes events: %s", err)
			}

			// By("Fetching curl-metrics logs")
			// logs, err = clientset.CoreV1().Pods(namespace).GetLogs("curl-metrics", &corev1.PodLogOptions{}).DoRaw(context.Background())
			// if err == nil {
			// 	_, _ = fmt.Fprintf(GinkgoWriter, "Metrics logs:\n %s", string(logs))
			// } else {
			// 	_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get curl-metrics logs: %s", err)
			// }

			By("Fetching controller manager pod description")
			pod := &corev1.Pod{}
			err = c.Get(context.Background(), client.ObjectKey{Name: podName, Namespace: namespace}, pod)
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Pod description:\nName: %s\nNamespace: %s\nStatus: %s\n",
					pod.Name, pod.Namespace, pod.Status.Phase)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to describe controller pod: %s", err)
			}
		}
	})

	Context("Stub", func() {
		It("stub", func() {
			fmt.Println("Stub")
			Expect(errors.New("")).ToNot(HaveOccurred())
		})
	})

	// SetDefaultEventuallyTimeout(2 * time.Minute)
	// SetDefaultEventuallyPollingInterval(time.Second)

	// Context("Manager", func() {
	// 	It("should run successfully", func() {
	// 		By("validating that the manager pod is running as expected")
	// 		verifyControllerUp := func(g Gomega) {
	// 			// Get the name of the controller-manager pod
	// 			cmd := exec.Command("kubectl", "get",
	// 				"pods", "-l", "app.kubernetes.io/name=moneypod",
	// 				"-o", "go-template={{ range .items }}"+
	// 					"{{ if not .metadata.deletionTimestamp }}"+
	// 					"{{ .metadata.name }}"+
	// 					"{{ \"\\n\" }}{{ end }}{{ end }}",
	// 				"-n", namespace,
	// 			)

	// 			podOutput, err := utils.Run(cmd)
	// 			g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve controller-manager pod information")
	// 			podNames := utils.GetNonEmptyLines(podOutput)
	// 			g.Expect(podNames).To(HaveLen(1), "expected 1 controller pod running")
	// 			controllerPodName = podNames[0]
	// 			g.Expect(controllerPodName).To(ContainSubstring("controller-manager"))

	// 			// Validate the pod's status
	// 			cmd = exec.Command("kubectl", "get",
	// 				"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
	// 				"-n", namespace,
	// 			)
	// 			output, err := utils.Run(cmd)
	// 			g.Expect(err).NotTo(HaveOccurred())
	// 			g.Expect(output).To(Equal("Running"), "Incorrect controller-manager pod status")
	// 		}
	// 		Eventually(verifyControllerUp).Should(Succeed())
	// 	})

	// 	It("should ensure the metrics endpoint is serving metrics", func() {
	// 		By("creating a ClusterRoleBinding for the service account to allow access to metrics")
	// 		cmd := exec.Command("kubectl", "create", "clusterrolebinding", metricsRoleBindingName,
	// 			"--clusterrole=moneypod-metrics-reader",
	// 			fmt.Sprintf("--serviceaccount=%s:%s", namespace, serviceAccountName),
	// 		)
	// 		_, err := utils.Run(cmd)
	// 		Expect(err).NotTo(HaveOccurred(), "Failed to create ClusterRoleBinding")

	// 		By("validating that the metrics service is available")
	// 		cmd = exec.Command("kubectl", "get", "service", metricsServiceName, "-n", namespace)
	// 		_, err = utils.Run(cmd)
	// 		Expect(err).NotTo(HaveOccurred(), "Metrics service should exist")

	// 		By("getting the service account token")
	// 		token, err := serviceAccountToken()
	// 		Expect(err).NotTo(HaveOccurred())
	// 		Expect(token).NotTo(BeEmpty())

	// 		By("waiting for the metrics endpoint to be ready")
	// 		verifyMetricsEndpointReady := func(g Gomega) {
	// 			cmd := exec.Command("kubectl", "get", "endpoints", metricsServiceName, "-n", namespace)
	// 			output, err := utils.Run(cmd)
	// 			g.Expect(err).NotTo(HaveOccurred())
	// 			g.Expect(output).To(ContainSubstring("8443"), "Metrics endpoint is not ready")
	// 		}
	// 		Eventually(verifyMetricsEndpointReady).Should(Succeed())

	// 		By("verifying that the controller manager is serving the metrics server")
	// 		verifyMetricsServerStarted := func(g Gomega) {
	// 			cmd := exec.Command("kubectl", "logs", controllerPodName, "-n", namespace)
	// 			output, err := utils.Run(cmd)
	// 			g.Expect(err).NotTo(HaveOccurred())
	// 			g.Expect(output).To(ContainSubstring("controller-runtime.metrics\tServing metrics server"),
	// 				"Metrics server not yet started")
	// 		}
	// 		Eventually(verifyMetricsServerStarted).Should(Succeed())

	// 		By("creating the curl-metrics pod to access the metrics endpoint")
	// 		cmd = exec.Command("kubectl", "run", "curl-metrics", "--restart=Never",
	// 			"--namespace", namespace,
	// 			"--image=curlimages/curl:latest",
	// 			"--overrides",
	// 			fmt.Sprintf(`{
	// 				"spec": {
	// 					"containers": [{
	// 						"name": "curl",
	// 						"image": "curlimages/curl:latest",
	// 						"command": ["/bin/sh", "-c"],
	// 						"args": ["curl -v -k -H 'Authorization: Bearer %s' https://%s.%s.svc.cluster.local:8443/metrics"],
	// 						"securityContext": {
	// 							"allowPrivilegeEscalation": false,
	// 							"capabilities": {
	// 								"drop": ["ALL"]
	// 							},
	// 							"runAsNonRoot": true,
	// 							"runAsUser": 1000,
	// 							"seccompProfile": {
	// 								"type": "RuntimeDefault"
	// 							}
	// 						}
	// 					}],
	// 					"serviceAccount": "%s"
	// 				}
	// 			}`, token, metricsServiceName, namespace, serviceAccountName))
	// 		_, err = utils.Run(cmd)
	// 		Expect(err).NotTo(HaveOccurred(), "Failed to create curl-metrics pod")

	// 		By("waiting for the curl-metrics pod to complete.")
	// 		verifyCurlUp := func(g Gomega) {
	// 			cmd := exec.Command("kubectl", "get", "pods", "curl-metrics",
	// 				"-o", "jsonpath={.status.phase}",
	// 				"-n", namespace)
	// 			output, err := utils.Run(cmd)
	// 			g.Expect(err).NotTo(HaveOccurred())
	// 			g.Expect(output).To(Equal("Succeeded"), "curl pod in wrong status")
	// 		}
	// 		Eventually(verifyCurlUp, 5*time.Minute).Should(Succeed())

	// 		By("getting the metrics by checking curl-metrics logs")
	// 		metricsOutput := getMetricsOutput()
	// 		Expect(metricsOutput).To(ContainSubstring(
	// 			"controller_runtime_reconcile_total",
	// 		))
	// 	})

	// 	// +kubebuilder:scaffold:e2e-webhooks-checks

	// 	// TODO: Customize the e2e test suite with scenarios specific to your project.
	// 	// Consider applying sample/CR(s) and check their status and/or verifying
	// 	// the reconciliation by using the metrics, i.e.:
	// 	// metricsOutput := getMetricsOutput()
	// 	// Expect(metricsOutput).To(ContainSubstring(
	// 	//    fmt.Sprintf(`controller_runtime_reconcile_total{controller="%s",result="success"} 1`,
	// 	//    strings.ToLower(<Kind>),
	// 	// ))
	// })
})

// // serviceAccountToken returns a token for the specified service account in the given namespace.
// // It uses the Kubernetes TokenRequest API to generate a token by directly sending a request
// // and parsing the resulting token from the API response.
// func serviceAccountToken() (string, error) {
// 	const tokenRequestRawString = `{
// 		"apiVersion": "authentication.k8s.io/v1",
// 		"kind": "TokenRequest"
// 	}`

// 	// Temporary file to store the token request
// 	secretName := fmt.Sprintf("%s-token-request", serviceAccountName)
// 	tokenRequestFile := filepath.Join("/tmp", secretName)
// 	err := os.WriteFile(tokenRequestFile, []byte(tokenRequestRawString), os.FileMode(0o644))
// 	if err != nil {
// 		return "", err
// 	}

// 	var out string
// 	verifyTokenCreation := func(g Gomega) {
// 		// Execute kubectl command to create the token
// 		cmd := exec.Command("kubectl", "create", "--raw", fmt.Sprintf(
// 			"/api/v1/namespaces/%s/serviceaccounts/%s/token",
// 			namespace,
// 			serviceAccountName,
// 		), "-f", tokenRequestFile)

// 		output, err := cmd.CombinedOutput()
// 		g.Expect(err).NotTo(HaveOccurred())

// 		// Parse the JSON output to extract the token
// 		var token tokenRequest
// 		err = json.Unmarshal(output, &token)
// 		g.Expect(err).NotTo(HaveOccurred())

// 		out = token.Status.Token
// 	}
// 	Eventually(verifyTokenCreation).Should(Succeed())

// 	return out, err
// }

// // getMetricsOutput retrieves and returns the logs from the curl pod used to access the metrics endpoint.
// func getMetricsOutput() string {
// 	By("getting the curl-metrics logs")
// 	cmd := exec.Command("kubectl", "logs", "curl-metrics", "-n", namespace)
// 	metricsOutput, err := utils.Run(cmd)
// 	Expect(err).NotTo(HaveOccurred(), "Failed to retrieve logs from curl pod")
// 	Expect(metricsOutput).To(ContainSubstring("< HTTP/1.1 200 OK"))
// 	return metricsOutput
// }

// // tokenRequest is a simplified representation of the Kubernetes TokenRequest API response,
// // containing only the token field that we need to extract.
// type tokenRequest struct {
// 	Status struct {
// 		Token string `json:"token"`
// 	} `json:"status"`
// }
