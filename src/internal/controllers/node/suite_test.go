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

package node

import (
	"context"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/vlasov-y/moneypod/internal/types"
	. "github.com/vlasov-y/moneypod/test/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	c          client.Client
	cancel     context.CancelFunc
	ctx        context.Context
	err        error
	rc         *rest.Config
	reconciler *NodeReconciler
	recorder   *record.FakeRecorder
	testEnv    *envtest.Environment
)

func TestNode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node Controller")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	recorder = record.NewFakeRecorder(5)
	ctx, cancel = context.WithCancel(context.Background())

	// +kubebuilder:scaffold:scheme

	By("bootstrapping envtest environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crds")},
		ErrorIfCRDPathMissing: false,
	}

	By("fetching binaries for envtest")
	cmd := exec.Command("task", "tools:setup-envtest")
	err = Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to install setupEnvtest")
	cwd, err := GetProjectDir()
	ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to get project dir")
	data, err := os.ReadFile(path.Join(cwd, ".task/envtest-path"))
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to read envtest-path file")
	testEnv.BinaryAssetsDirectory = strings.TrimSpace(string(data))

	By("starting the environment")
	rc, err = testEnv.Start()
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to start envtest")
	ExpectWithOffset(2, rc).NotTo(BeNil(), "empty envtest config")

	By("creating a client.Client")
	c, err = client.New(rc, client.Options{Scheme: scheme})
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to create a client.Client")
	ExpectWithOffset(2, c).NotTo(BeNil(), "empty client.Client")

	By("creating a reconciler")
	reconciler = &NodeReconciler{
		Reconciler: Reconciler{
			Client:   c,
			Config:   rc,
			Scheme:   scheme,
			Recorder: recorder,
		},
	}
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err = testEnv.Stop()
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to stop envtest")
})
