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

package edgewatcher_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/google/uuid"
	edgewatcher "github.com/nephio-project/edge-watcher"
	"github.com/nephio-project/edge-watcher/preprocessor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("Router", func() {
	var router *edgewatcher.Router
	var testClient *subscriber
	var nfdeploySource, clusterSource *fakeEventSource
	var clusterOption, nfdeployOption edgewatcher.EventOptions
	var ctx context.Context
	var cancelCtx context.CancelFunc

	BeforeEach(func() {
		rand.Seed(GinkgoRandomSeed())

		logger := zap.New(func(options *zap.Options) {
			options.Development = true
			options.DestWriter = GinkgoWriter
		})

		ctx, cancelCtx = context.WithCancel(context.Background())
		router = edgewatcher.NewRouter(logger, ctx)

		nfdeploySource = &fakeEventSource{
			logger: logger.WithName("nfdeploySource"),
			router: router,
		}

		clusterSource = &fakeEventSource{
			logger: logger.WithName("clusterSource"),
			router: router,
		}

		testClient = &subscriber{
			logger: logger,
			router: router,
		}

		clusterOption = edgewatcher.EventOptions{
			Type:             edgewatcher.ClusterSubscriber,
			SubscriptionName: "cluster1",
		}
		nfdeployOption = edgewatcher.EventOptions{
			Type:             edgewatcher.NfDeploySubscriber,
			SubscriptionName: "nfdeploy1",
		}

	})

	Describe("Route events", func() {
		Context("with different subscriber types", func() {
			It("should send correct event to each subscriber", func() {
				clusterReq, err := testClient.subscribe(ctx, clusterOption)
				Expect(err).To(BeNil())
				nfdeployReq, err := testClient.subscribe(ctx, nfdeployOption)
				Expect(err).To(BeNil())
				var sourceWg sync.WaitGroup
				sourceWg.Add(2)
				go clusterSource.run(ctx, &sourceWg, events["cluster1"]...)
				go nfdeploySource.run(ctx, &sourceWg, events["nfdeploy1"]...)
				var clientWg sync.WaitGroup
				clientWg.Add(2)

				go testClient.receiveEvents(ctx, receiveOptions{
					wg:         &clientWg,
					events:     events["cluster1"],
					eventCount: len(events["cluster1"]),
					requests:   []*edgewatcher.SubscriptionReq{clusterReq},
				})
				go testClient.receiveEvents(ctx, receiveOptions{
					wg:         &clientWg,
					events:     events["nfdeploy1"],
					eventCount: len(events["nfdeploy1"]),
					requests:   []*edgewatcher.SubscriptionReq{nfdeployReq},
				})
				clientWg.Wait()
				sourceWg.Wait()
			})
		})
		Context("with multiple subscriber of Cluster type", func() {
			It("should duplicate events", func() {
				clusterReq1, err := testClient.subscribe(ctx, clusterOption)
				Expect(err).To(BeNil())
				clusterReq2, err := testClient.subscribe(ctx, clusterOption)
				Expect(err).To(BeNil())
				var sourceWg sync.WaitGroup
				sourceWg.Add(1)
				go clusterSource.run(ctx, &sourceWg, events["cluster1"]...)
				var clientWg sync.WaitGroup
				clientWg.Add(2)
				go testClient.receiveEvents(ctx, receiveOptions{
					wg:         &clientWg,
					events:     events["cluster1"],
					eventCount: len(events["cluster1"]),
					requests:   []*edgewatcher.SubscriptionReq{clusterReq1},
				})
				go testClient.receiveEvents(ctx, receiveOptions{
					wg:         &clientWg,
					events:     events["cluster1"],
					eventCount: len(events["cluster1"]),
					requests:   []*edgewatcher.SubscriptionReq{clusterReq2},
				})
				clientWg.Wait()
				sourceWg.Wait()
			})
		})

		Context("with multiple subscriber of Nfdeploy type", func() {
			It("should duplicate events", func() {
				nfdeployReq1, err := testClient.subscribe(ctx, nfdeployOption)
				Expect(err).To(BeNil())
				nfdeployReq2, err := testClient.subscribe(ctx, nfdeployOption)
				Expect(err).To(BeNil())
				var sourceWg sync.WaitGroup
				sourceWg.Add(1)
				go nfdeploySource.run(ctx, &sourceWg, events["nfdeploy1"]...)
				var clientWg sync.WaitGroup
				clientWg.Add(2)
				go testClient.receiveEvents(ctx, receiveOptions{
					wg:         &clientWg,
					events:     events["nfdeploy1"],
					eventCount: len(events["nfdeploy1"]),
					requests:   []*edgewatcher.SubscriptionReq{nfdeployReq1},
				})
				go testClient.receiveEvents(ctx, receiveOptions{
					wg:         &clientWg,
					events:     events["nfdeploy1"],
					eventCount: len(events["nfdeploy1"]),
					requests:   []*edgewatcher.SubscriptionReq{nfdeployReq2},
				})
				clientWg.Wait()
				sourceWg.Wait()
			})
		})

		Context("fuzzy tests", func() {
			It("should route events", func() {
				clusterEvents := generateEvents(500, preprocessor.Added, preprocessor.RequestKey{
					ClusterName: clusterOption.SubscriptionName,
					NFDeploy:    "nfdeploy2",
					Kind:        "UPFDeploy",
				})
				clusterEvents = append(clusterEvents, generateEvents(500, preprocessor.Added, preprocessor.RequestKey{
					ClusterName: clusterOption.SubscriptionName,
					NFDeploy:    "nfdeploy2",
					Kind:        "SMFDeploy",
				})...)
				nfdeployEvents := generateEvents(500, preprocessor.Added, preprocessor.RequestKey{
					ClusterName: "cluster2",
					NFDeploy:    nfdeployOption.SubscriptionName,
					Kind:        "UPFDeploy",
				})
				nfdeployEvents = append(nfdeployEvents, generateEvents(500, preprocessor.Added, preprocessor.RequestKey{
					ClusterName: "cluster2",
					NFDeploy:    nfdeployOption.SubscriptionName,
					Kind:        "SMFDeploy",
				})...)

				var wg sync.WaitGroup
				wg.Add(100)

				for i := 0; i < 100; i++ {
					nfdeployReq, err := testClient.subscribe(ctx, nfdeployOption)
					Expect(err).To(BeNil())

					clusterReq, err := testClient.subscribe(ctx, clusterOption, nfdeployReq.SubscriberName)
					Expect(err).To(BeNil())

					eventCount := rand.Intn(2000) + 1
					go testClient.receiveEvents(ctx, receiveOptions{
						wg:         &wg,
						events:     append(nfdeployEvents, clusterEvents...),
						eventCount: eventCount,
						requests:   []*edgewatcher.SubscriptionReq{nfdeployReq, clusterReq},
					})
				}

				wg.Add(2)
				go nfdeploySource.run(ctx, &wg, nfdeployEvents...)
				go clusterSource.run(ctx, &wg, clusterEvents...)

				wg.Wait()
			})
		})
	})

	Describe("Router sends error", func() {
		Context("with duplicate SubscriberInfo request", func() {
			It("should return error", func() {
				clusterReq1, err := testClient.subscribe(ctx, clusterOption)
				Expect(err).To(BeNil())
				clusterReq2 := &edgewatcher.SubscriptionReq{
					Ctx:          context.Background(),
					Error:        make(chan error),
					EventOptions: clusterOption,
					SubscriberInfo: edgewatcher.SubscriberInfo{
						SubscriberName: clusterReq1.SubscriberName,
						Channel:        make(chan preprocessor.Event),
					},
				}
				router.Subscribe() <- clusterReq2
				expectedErr := fmt.Sprintf("duplicate subscription request, SubscriberName: %v", clusterReq1.SubscriberName)
				Eventually(clusterReq2.Error).Should(Receive(MatchError(expectedErr)))
			})
		})
		Context("with invalid cancel request", func() {
			It("should return error", func() {
				cancelReq := &edgewatcher.SubscriptionReq{
					Ctx:          ctx,
					Error:        make(chan error),
					EventOptions: clusterOption,
					SubscriberInfo: edgewatcher.SubscriberInfo{
						SubscriberName: uuid.NewString(),
						Channel:        make(chan preprocessor.Event),
					},
				}
				router.Cancel() <- cancelReq
				Eventually(cancelReq.Error).Should(Receive(MatchError(fmt.Errorf("unknown cancel request recieved: %v , options: %v", cancelReq.SubscriberName, cancelReq.EventOptions))))
			})
		})

		Context("with unknown subscriber type", func() {
			It("should return error", func() {
				subscribeReq := &edgewatcher.SubscriptionReq{
					Ctx:   ctx,
					Error: make(chan error),
					EventOptions: edgewatcher.EventOptions{
						Type: "customType",
					},
					SubscriberInfo: edgewatcher.SubscriberInfo{
						Channel:        make(chan preprocessor.Event),
						SubscriberName: uuid.NewString(),
					},
				}
				router.Subscribe() <- subscribeReq
				Eventually(subscribeReq.Error).Should(Receive(MatchError(fmt.Errorf("unknown subscriber type: %v", subscribeReq.Type))))
			})
		})
	})

	AfterEach(func() {
		cancelCtx()
	})

})
