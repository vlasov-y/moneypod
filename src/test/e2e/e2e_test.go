// Copyright 2025 The MoneyPod Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	promclient "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	. "github.com/vlasov-y/moneypod/internal/types"
	. "github.com/vlasov-y/moneypod/test/utils"
	appsv1 "k8s.io/api/apps/v1"
	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/describe"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespace            = "moneypod"
	serviceAccountName   = "moneypod-app"
	metricsNodePort      = 30000
	metricsTLSServerName = "moneypod-metrics"
)

var _ = Describe("Manager", Ordered, func() {
	var podName string
	var c client.Client
	var cs *kubernetes.Clientset
	var rc *rest.Config
	var err error
	var cmd *exec.Cmd
	var metricsCAPool *x509.CertPool
	var metricsBearerToken string

	var getPrometheusMetrics = func() (metrics string, err error) {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:    metricsCAPool,
					ServerName: metricsTLSServerName},
			},
		}
		req, err := http.NewRequest("GET", fmt.Sprintf("https://localhost:%d/metrics", metricsNodePort), nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", metricsBearerToken))
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		return string(body), err
	}

	var parsePrometheusMetrics = func(metricsText string) (map[string]*promclient.MetricFamily, error) {
		parser := expfmt.NewTextParser(model.UTF8Validation)
		reader := strings.NewReader(metricsText)
		return parser.TextToMetricFamilies(reader)
	}

	// Before running the tests, set up the environment by creating the namespace,
	// enforce the restricted security policy to the namespace, installing CRDs,
	// and deploying the controller.
	BeforeAll(func() {
		By("Installing operator to the cluster")
		cmd = exec.Command("task", "install-operator")
		err = Run(cmd)
		ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to install the operator")

		By("Initializing Kubernetes clients")
		var cwd string
		cwd, err = GetProjectDir()
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to get project dir")
		rc, err = clientcmd.BuildConfigFromFlags("", path.Join(cwd, "kubeconfig.yaml"))
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to load kubeconfig")
		c, err = client.New(rc, client.Options{})
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to create the client")
		cs, err = kubernetes.NewForConfig(rc)
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to create the clientset")

		By("Finding operator pod name")
		var pod corev1.Pod
		pods := &corev1.PodList{}
		err = c.List(context.Background(), pods, client.InNamespace(namespace))
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to list pods in the namespace")
		Expect(pods.Items).To(HaveLen(1), "expected eactly 1 pod in the namespace")
		pod = pods.Items[0]
		podName = pod.Name

		By("Fetching metrics CA certificate")
		var secretName string
		for _, vol := range pod.Spec.Volumes {
			if vol.Name == "metrics-tls" {
				secretName = vol.Secret.SecretName
				break
			}
		}
		ExpectWithOffset(7, secretName).ToNot(BeEmpty())
		secret := &corev1.Secret{}
		err = c.Get(context.Background(), client.ObjectKey{Name: secretName, Namespace: namespace}, secret)
		ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to get metrics CA secret")
		block, rest := pem.Decode(secret.Data["ca.crt"])
		ExpectWithOffset(1, rest).To(BeEmpty(), "failed to decode metrics CA certificate PEM block")
		var crt *x509.Certificate
		crt, err = x509.ParseCertificate(block.Bytes)
		ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to parse metrics CA certificate")
		metricsCAPool = x509.NewCertPool()
		metricsCAPool.AddCert(crt)

		By("Issuing a bearer token for metrics authorization")
		var result *authv1.TokenRequest
		result, err = cs.CoreV1().ServiceAccounts(namespace).CreateToken(
			context.Background(),
			serviceAccountName,
			&authv1.TokenRequest{
				Spec: authv1.TokenRequestSpec{
					ExpirationSeconds: ptr.To(int64(3600)),
				},
			},
			metav1.CreateOptions{},
		)
		ExpectWithOffset(10, err).ToNot(HaveOccurred())
		metricsBearerToken = result.Status.Token
		ExpectWithOffset(1, metricsBearerToken).ToNot(BeEmpty())

		By("Fetching metrics values")
		metrics, err := getPrometheusMetrics()
		ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to get metrics")
		ExpectWithOffset(2, metrics).ToNot(BeEmpty(), "no metrics in the list")
	})

	// After all tests have been executed, clean up by undeploying the controller, uninstalling CRDs,
	// and deleting the namespace.
	AfterAll(func() {
		By("Uninstalling the operator")
		cmd = exec.Command("task", "uninstall-operator")
		err = Run(cmd)
		ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to uninstall the operator")
	})

	// After each test, check for failures and collect logs, events,
	// and pod descriptions for debugging.
	AfterEach(func() {
		specReport := CurrentSpecReport()
		if specReport.Failed() {
			By("Fetching controller manager pod logs")
			logs, err := cs.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{}).DoRaw(context.Background())
			ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to get Controller logs")
			fmt.Fprintf(GinkgoWriter, ">>>\n>>> Controller logs\n>>>\n %s\n", string(logs))

			By("Fetching Kubernetes events")
			events := &corev1.EventList{}
			err = c.List(context.Background(), events, client.InNamespace(namespace))
			ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to get Kubernetes events")
			fmt.Fprintln(GinkgoWriter, ">>>\n>>> Kubernetes Events\n>>>")
			for _, event := range events.Items {
				fmt.Fprintf(GinkgoWriter, "%s | %s | %s\n", event.LastTimestamp.Time, event.Reason, event.Message)
			}

			By("Describing operator pod")
			fmt.Fprintln(GinkgoWriter, ">>>\n>>> Pod Describe\n>>>")
			describer := describe.PodDescriber{Interface: cs}
			output, err := describer.Describe(namespace, podName, describe.DescriberSettings{
				ShowEvents: true,
			})
			ExpectWithOffset(2, err).NotTo(HaveOccurred(), "failed to describe the pod")
			fmt.Fprintln(GinkgoWriter, output)

			By("Fetching metrics values")
			fmt.Fprintln(GinkgoWriter, ">>>\n>>> Metrics\n>>>")
			metrics, err := getPrometheusMetrics()
			ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to get metrics")
			ExpectWithOffset(2, metrics).ToNot(BeEmpty(), "no metrics in the list")
			for _, line := range strings.Split(metrics, "\n") {
				if strings.HasPrefix(line, "moneypod_") {
					fmt.Fprintln(GinkgoWriter, line)
				}
			}
		}
	})

	Context("Metrics", func() {
		var metricFamilies map[string]*promclient.MetricFamily

		It("should parse metrics text", func() {
			By("Get metrics")
			metrics, err := getPrometheusMetrics()
			ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to get metrics")
			By("Parse metrics")
			metricFamilies, err = parsePrometheusMetrics(metrics)
			ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to parse metrics")
		})

		It("should have node metrics families", func() {
			By("Check that moneypod_node_hourly_cost exists")
			Expect(metricFamilies).To(HaveKey("moneypod_node_hourly_cost"))
			reconcileMetric := metricFamilies["moneypod_node_hourly_cost"]
			By("Check that moneypod_node_hourly_cost is Gauge")
			ExpectWithOffset(1, reconcileMetric.GetType()).To(Equal(promclient.MetricType_GAUGE))
		})

		It("moneypod_node_hourly_cost should match nodes annotations", func() {
			reconcileMetric := metricFamilies["moneypod_node_hourly_cost"]
			for _, metric := range reconcileMetric.GetMetric() {
				labels := make(map[string]string)
				for _, label := range metric.GetLabel() {
					labels[label.GetName()] = label.GetValue()
				}

				By("Check node and name labels values are equal")
				Expect(labels).To(HaveKey("node"), "failed to find a node label")
				Expect(labels).To(HaveKeyWithValue("name", labels["node"]), "name label must be equal to pod label")

				By("Get Node to compare with")
				node := &corev1.Node{}
				err = c.Get(context.Background(), client.ObjectKey{Name: labels["node"]}, node)
				ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to get Node mentioned in label")
				annotations := node.GetAnnotations()
				ExpectWithOffset(1, annotations).ToNot(BeEmpty(), "node does not have annotations")

				By("Ensure Node is properly annotated")
				Expect(annotations).To(HaveKey(AnnotationNodeAvailabilityZone))
				Expect(annotations).To(HaveKey(AnnotationNodeCapacity))
				Expect(annotations).To(HaveKey(AnnotationNodeHourlyCost))
				Expect(annotations).To(HaveKey(AnnotationNodeType))

				By("Compare metric labels and Node annotations")
				Expect(labels).To(HaveKeyWithValue("availability_zone", annotations[AnnotationNodeAvailabilityZone]))
				Expect(labels).To(HaveKeyWithValue("capacity", annotations[AnnotationNodeCapacity]))
				Expect(labels).To(HaveKeyWithValue("type", annotations[AnnotationNodeType]))

				By("Compare metric value with hourly cost from Node annotation")
				var hourlyCost float64
				hourlyCost, err = strconv.ParseFloat(annotations[AnnotationNodeHourlyCost], 64)
				ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to parse Node hourly cost from the annotation")
				Expect(metric.GetGauge().GetValue()).To(BeNumerically("==", hourlyCost))
			}
		})

		It("should have pod metrics families", func() {
			for _, family := range []string{
				"moneypod_pod_cpu_hourly_cost",
				"moneypod_pod_memory_hourly_cost",
				"moneypod_pod_requests_hourly_cost",
			} {
				By(fmt.Sprintf("Check that %s exists", family))
				Expect(metricFamilies).To(HaveKey(family))
				reconcileMetric := metricFamilies[family]
				By(fmt.Sprintf("Check that %s is Gauge", family))
				ExpectWithOffset(1, reconcileMetric.GetType()).To(Equal(promclient.MetricType_GAUGE))
			}
		})

		It("should have valid pod labels", func() {
			for _, family := range []string{
				"moneypod_pod_cpu_hourly_cost",
				"moneypod_pod_memory_hourly_cost",
				"moneypod_pod_requests_hourly_cost",
			} {
				reconcileMetric := metricFamilies[family]
				for _, metric := range reconcileMetric.GetMetric() {
					labels := make(map[string]string)
					for _, label := range metric.GetLabel() {
						labels[label.GetName()] = label.GetValue()
					}

					Expect(labels).To(HaveKey("pod"), "failed to find a pod label")
					Expect(labels).To(HaveKey("namespace"), "failed to find a namespace label")

					By(fmt.Sprintf("%s/%s", labels["namespace"], labels["pod"]))
					Expect(labels).To(HaveKeyWithValue("name", labels["pod"]), "name label must be equal to pod label")

					pod := &corev1.Pod{}
					err = c.Get(context.Background(), client.ObjectKey{Name: labels["pod"], Namespace: labels["namespace"]}, pod)
					ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to get Pod mentioned in label")

					Expect(labels).To(HaveKey("node"), "failed to find a node label")
					Expect(labels["node"]).To(Equal(pod.Spec.NodeName), "node label does not match pod.spec.nodeName")

					Expect(labels).To(HaveKey("owner_kind"), "failed to find a owner_kind label")
					Expect(labels).To(HaveKey("owner_name"), "failed to find a owner_name label")
					if labels["owner_kind"] != "" {
						Expect(pod.GetOwnerReferences()).ToNot(BeEmpty(), "label has owner set, but pod does not have an owner")
						if labels["owner_kind"] == "Deployment" {
							deployment := &appsv1.Deployment{}
							err = c.Get(context.Background(),
								client.ObjectKey{Name: labels["owner_name"], Namespace: labels["namespace"]}, deployment)
							ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to get Deployment mentioned in owner labels")
						} else {
							owner := pod.GetOwnerReferences()[0]
							Expect(labels).To(HaveKeyWithValue("owner_kind", owner.Kind), "failed to find a owner_kind label")
							Expect(labels).To(HaveKeyWithValue("owner_name", owner.Name), "failed to find a owner_name label")
						}
					} else {
						Expect(pod.GetOwnerReferences()).To(BeEmpty(), "pod has owners, but labels no")
					}
				}
			}
		})
	})
})
