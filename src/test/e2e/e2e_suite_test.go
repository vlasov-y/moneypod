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
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestE2E runs the end-to-end (e2e) test suite for the project. These tests execute in an isolated,
// temporary environment to validate project changes with the purposed to be used in CI jobs.
// The default setup requires Kind, builds/loads the Manager Docker image locally, and installs
// CertManager.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "End-to-end")
}

var _ = BeforeSuite(func() {
	By("creating and bootstrapping a kind cluster")
	cmd := exec.Command("task", "kind:bootstrap")
	_, err := run(cmd)
	ExpectWithOffset(1, err).ToNot(HaveOccurred(), "Failed to create the kind cluster")

	By("building and loading docker image to the cluster")
	cmd = exec.Command("task", "docker:build-and-load")
	_, err = run(cmd)
	ExpectWithOffset(1, err).ToNot(HaveOccurred(), "Failed to build and load the image")
})

// run will execute command and grab its output
func run(cmd *exec.Cmd) (string, error) {
	dir, _ := getProjectDir()
	cmd.Dir = dir

	if err := os.Chdir(cmd.Dir); err != nil {
		fmt.Fprintf(GinkgoWriter, "chdir dir: %q\n", err)
	}

	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	command := strings.Join(cmd.Args, " ")
	fmt.Fprintf(GinkgoWriter, "running: %q\n", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(GinkgoWriter, string(output))
		return string(output), fmt.Errorf("%q failed with error: %w", command, err)
	}
	return string(output), nil
}

// getProjectDir will return the directory where the project is
func getProjectDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return wd, fmt.Errorf("failed to get current working directory: %w", err)
	}
	wd = strings.ReplaceAll(wd, "/test/e2e", "")
	return wd, nil
}
