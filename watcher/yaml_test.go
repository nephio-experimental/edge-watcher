/*
Copyright 2022-2023 The Nephio Authors.

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

package watcher_test

import (
	"fmt"
	"time"

	"github.com/nephio-project/edge-watcher/watcher"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const yamlFormat = `apiVersion: monitor.nephio.org/v1alpha1
kind: WatcherAgent
metadata:
  name: watcheragent-upf-smf
  namespace: %v
  annotations:
    cloud.nephio.org/cluster: %v
spec:
  resyncTimestamp: "%v"
  edgeWatcher:
    addr: %v
    port: "%v"
  watchRequests:
  - group: workload.nephio.org
    version: v1alpha1
    kind: UPFDeployment
    namespace: %v
  - group: workload.nephio.org
    version: v1alpha1
    kind: SMFDeployment
    namespace: %v
  - group: workload.nephio.org
    version: v1alpha1
    kind: AMFDeployment
    namespace: %v
`

var _ = Describe("GetYaml", func() {
	var (
		podIP, port, clusterName, nephioNamespace string
		timeStamp                                 time.Time
	)
	BeforeEach(func() {
		podIP = "10.0.4.0"
		port = "3000"
		clusterName = "cluster1"
		nephioNamespace = "nephio-system"
		timeStamp = time.Now()
	})
	It("should create correct yaml", func() {
		expectedYaml := fmt.Sprintf(yamlFormat, nephioNamespace, clusterName, timeStamp.Format(time.RFC3339), podIP, port, nephioNamespace, nephioNamespace, nephioNamespace)
		Expect(watcher.GetYAML(podIP, port, clusterName, nephioNamespace, timeStamp)).Should(Equal(expectedYaml))
	})
})
