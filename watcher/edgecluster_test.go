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
	"strconv"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/nephio-project/common-lib/edge/porch"
	"github.com/nephio-project/edge-watcher/watcher"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("CreateWatcherAgentCRs", func() {
	var (
		logger      logr.Logger
		podIP, port string
		timeStamp   time.Time
		porchClient *porch.FakeClient
	)
	testEnv.InitOnRunningSuite()
	BeforeEach(func() {
		logger = zap.New(func(options *zap.Options) {
			options.Development = true
			options.DestWriter = GinkgoWriter
		})

		podIP = "10.0.2.0"
		port = "3000"
		timeStamp = time.Now()
		porchClient = porch.NewFakeClient()
	})
	Context("porch client returns nil error", func() {
		It("should create corresponding porch packages", func(ctx SpecContext) {
			go createEdgeClusters(ctx, 100)
			go func() {
				defer GinkgoRecover()
				params := watcher.Params{
					K8sDynamicClient:     testEnv.DynamicClient,
					PorchClient:          porchClient,
					Port:                 port,
					PodIP:                podIP,
					EdgeClusterNamespace: testEnv.GetNamespace(),
					NephioNamespace:      testEnv.GetNamespace(),
					Timestamp:            timeStamp,
				}
				err := watcher.CreateWatcherAgentCRs(ctx, logger, params)
				Expect(err).To(BeNil())
			}()

			var wg sync.WaitGroup
			wg.Add(100)
			for i := 1; i <= 100; i++ {
				go func(i int) {
					defer GinkgoRecover()
					defer wg.Done()
					clusterName := "edgecluster" + strconv.Itoa(i)
					expectedYaml := fmt.Sprintf(yamlFormat, testEnv.GetNamespace(), clusterName, timeStamp.Format(time.RFC3339), podIP, port, testEnv.GetNamespace(), testEnv.GetNamespace())

					Eventually(func(g Gomega) map[string]string {
						content, err := porchClient.GetContent(ctx, "watcherAgent", clusterName)
						g.Expect(err).To(BeNil())
						return content
					}, ctx, 10*time.Second).Should(HaveKeyWithValue("watcheragent.yaml", expectedYaml))
				}(i)
			}
			logger.Info("waiting for assertions to finish")
			wg.Wait()
		}, SpecTimeout(10*time.Second))
	})
	Context("porch client returns error", func() {
		It("should retry", func(ctx SpecContext) {
			clusterName := "edgecluster1"
			err := porchClient.SetError(ctx, "watcherAgent", clusterName, fmt.Errorf("fake error"))
			Expect(err).To(BeNil())
			go createEdgeClusters(ctx, 1)
			go func() {
				defer GinkgoRecover()
				params := watcher.Params{
					K8sDynamicClient:     testEnv.DynamicClient,
					PorchClient:          porchClient,
					Port:                 port,
					PodIP:                podIP,
					EdgeClusterNamespace: testEnv.GetNamespace(),
					NephioNamespace:      testEnv.GetNamespace(),
					Timestamp:            timeStamp,
				}
				err := watcher.CreateWatcherAgentCRs(ctx, logger, params)
				Expect(err).To(BeNil())
			}()
			Eventually(func(g Gomega) int {
				count, err := porchClient.GetRetryCount(ctx, "watcherAgent", clusterName)
				g.Expect(err).To(BeNil())
				return count
			}, ctx, 10*time.Second).Should(BeNumerically(">", 1))
		})
	})

})
