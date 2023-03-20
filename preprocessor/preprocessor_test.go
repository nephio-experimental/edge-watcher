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

package preprocessor_test

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"github.com/nephio-project/edge-watcher/preprocessor"
	pb "github.com/nephio-project/edge-watcher/protos"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type fakeRouter struct {
	streams sync.Map // map[preprocessor.RequestKey]chan *preprocessor.EventMessage
}

func (f *fakeRouter) RouteEvent(key preprocessor.RequestKey) (
	ch chan *preprocessor.EventMessage) {
	chIntf, ok := f.streams.Load(key)
	if !ok {
		ch = make(chan *preprocessor.EventMessage)
		f.streams.Store(key, ch)
	} else {
		ch = chIntf.(chan *preprocessor.EventMessage)
	}
	return
}

type fakeSource struct {
	logger   logr.Logger
	receiver preprocessor.EventReceiver
}

type sourceParams struct {
	ctx        context.Context
	wg         *sync.WaitGroup
	shouldFail bool
	events     []*pb.EventRequest
	response   pb.ResponseType
}

func (f *fakeSource) sendEvents(params sourceParams) {
	defer GinkgoRecover()
	defer params.wg.Done()
	logger := f.logger.WithName("fakeSource.sendEvents")
	debugLogger := logger.V(1)

	for _, event := range params.events {
		resp, err := f.receiver.ReceiveEvent(params.ctx, event)
		debugLogger.Info("response received", "resp", resp, "err", err)
		if resp != params.response {
			Fail(fmt.Sprintf("incorrect response, expected: %v, received: %v", params.response, resp))
		}
		if err != nil {
			if params.shouldFail {
				return
			} else {
				Fail(fmt.Sprintf("RESET response received with err: %v", err.Error()))
			}
		}
	}
	if params.shouldFail {
		Fail("no RESET response received")
	}
}

var _ = Describe("Preprocessor", func() {
	var eventPreprocessor *preprocessor.EventPreprocessor
	var router *fakeRouter
	var fakeServer *fakeSource
	var ctx context.Context
	var cancelCtx context.CancelFunc
	BeforeEach(func() {
		logger := zap.New(func(options *zap.Options) {
			options.Development = true
			options.DestWriter = GinkgoWriter
		})
		router = &fakeRouter{}
		eventPreprocessor = preprocessor.New(logger, router)
		fakeServer = &fakeSource{
			logger:   logger,
			receiver: eventPreprocessor,
		}
		ctx, cancelCtx = context.WithCancel(context.Background())
	})
	Describe("List event", func() {
		Context("when router returns nil", func() {
			It("should send the correct event", func() {
				var sourceWg sync.WaitGroup
				sourceWg.Add(1)
				go fakeServer.sendEvents(sourceParams{
					ctx:        ctx,
					wg:         &sourceWg,
					shouldFail: false,
					events:     []*pb.EventRequest{eventLists["smfdeploy list"].event},
					response:   pb.ResponseType_OK,
				})
				eventReq := <-router.RouteEvent(eventLists["smfdeploy list"].key)
				Expect(*eventLists["smfdeploy list"].finalEvent).To(MatchFields(IgnoreExtras, Fields{
					"Type":      Equal(eventReq.Event.Type),
					"Key":       Equal(eventReq.Event.Key),
					"Timestamp": BeTemporally("==", eventReq.Timestamp),
					"Object":    Equal(eventReq.Event.Object),
				}))
				eventReq.Err <- nil
				sourceWg.Wait()
			})
		})
		Context("when router return error", func() {
			It("should return RESET response", func() {
				var sourceWg sync.WaitGroup
				sourceWg.Add(1)
				go fakeServer.sendEvents(sourceParams{
					ctx:        ctx,
					wg:         &sourceWg,
					shouldFail: true,
					events:     []*pb.EventRequest{eventLists["smfdeploy list"].event},
					response:   pb.ResponseType_RESET,
				})
				eventReq := <-router.RouteEvent(eventLists["smfdeploy list"].key)
				eventReq.Err <- fmt.Errorf("fake router error")
				sourceWg.Wait()
			})
		})
	})

	Describe("Watch events", func() {
		Context("after receiving list", func() {
			BeforeEach(func() {
				var sourceWg sync.WaitGroup
				sourceWg.Add(1)
				go fakeServer.sendEvents(sourceParams{
					ctx:        ctx,
					wg:         &sourceWg,
					shouldFail: false,
					events:     []*pb.EventRequest{eventLists["smfdeploy list"].event},
				})
				listEventReq := <-router.RouteEvent(eventLists["smfdeploy list"].key)
				listEventReq.Err <- nil
				Expect(*eventLists["smfdeploy list"].finalEvent).To(MatchFields(IgnoreExtras, Fields{
					"Type":      Equal(listEventReq.Event.Type),
					"Key":       Equal(listEventReq.Event.Key),
					"Timestamp": BeTemporally("==", listEventReq.Timestamp),
					"Object":    Equal(listEventReq.Event.Object),
				}))
				sourceWg.Wait()
			})
			It("should send add watch event", func() {
				var sourceWg sync.WaitGroup
				sourceWg.Add(1)
				go fakeServer.sendEvents(sourceParams{
					ctx:        ctx,
					wg:         &sourceWg,
					shouldFail: false,
					events:     []*pb.EventRequest{eventLists["smfdeploy add event"].event},
					response:   pb.ResponseType_OK,
				})
				watchEventReq := <-router.RouteEvent(eventLists["smfdeploy add event"].key)
				Expect(*eventLists["smfdeploy add event"].finalEvent).To(MatchFields(IgnoreExtras, Fields{
					"Type":      Equal(watchEventReq.Event.Type),
					"Key":       Equal(watchEventReq.Event.Key),
					"Timestamp": BeTemporally("==", watchEventReq.Timestamp),
					"Object":    Equal(watchEventReq.Event.Object),
				}))
				watchEventReq.Err <- nil
				sourceWg.Wait()
			})
			It("should send modify watch event", func() {
				var sourceWg sync.WaitGroup
				sourceWg.Add(1)
				go fakeServer.sendEvents(sourceParams{
					ctx:        ctx,
					wg:         &sourceWg,
					shouldFail: false,
					events:     []*pb.EventRequest{eventLists["smfdeploy modify event"].event},
					response:   pb.ResponseType_OK,
				})
				watchEventReq := <-router.RouteEvent(eventLists["smfdeploy modify event"].key)
				Expect(*eventLists["smfdeploy modify event"].finalEvent).To(MatchFields(IgnoreExtras, Fields{
					"Type":      Equal(watchEventReq.Event.Type),
					"Key":       Equal(watchEventReq.Event.Key),
					"Timestamp": BeTemporally("==", watchEventReq.Timestamp),
					"Object":    Equal(watchEventReq.Event.Object),
				}))
				watchEventReq.Err <- nil
				sourceWg.Wait()
			})
			It("should send delete watch event", func() {
				var sourceWg sync.WaitGroup
				sourceWg.Add(1)
				go fakeServer.sendEvents(sourceParams{
					ctx:        ctx,
					wg:         &sourceWg,
					shouldFail: false,
					events:     []*pb.EventRequest{eventLists["smfdeploy delete event"].event},
					response:   pb.ResponseType_OK,
				})
				watchEventReq := <-router.RouteEvent(eventLists["smfdeploy delete event"].key)
				Expect(*eventLists["smfdeploy delete event"].finalEvent).To(MatchFields(IgnoreExtras, Fields{
					"Type":      Equal(watchEventReq.Event.Type),
					"Key":       Equal(watchEventReq.Event.Key),
					"Timestamp": BeTemporally("==", watchEventReq.Timestamp),
					"Object":    Equal(watchEventReq.Event.Object),
				}))
				watchEventReq.Err <- nil
				sourceWg.Wait()
			})
			Context("router returns error", func() {
				It("should return RESET response", func() {
					var sourceWg sync.WaitGroup
					sourceWg.Add(1)
					go fakeServer.sendEvents(sourceParams{
						ctx:        ctx,
						wg:         &sourceWg,
						shouldFail: true,
						events:     []*pb.EventRequest{eventLists["smfdeploy add event"].event},
						response:   pb.ResponseType_RESET,
					})
					watchEventReq := <-router.RouteEvent(eventLists["smfdeploy add event"].key)
					watchEventReq.Err <- fmt.Errorf("fake router error")
					sourceWg.Wait()
				})
			})
		})
		Context("without list event", func() {
			It("should return RESET response", func() {
				var sourceWg sync.WaitGroup
				sourceWg.Add(1)
				go fakeServer.sendEvents(sourceParams{
					ctx:        ctx,
					wg:         &sourceWg,
					shouldFail: true,
					events:     []*pb.EventRequest{eventLists["smfdeploy add event"].event},
					response:   pb.ResponseType_RESET,
				})
				sourceWg.Wait()
			})
		})
	})
	AfterEach(func() {
		cancelCtx()
	})
})
