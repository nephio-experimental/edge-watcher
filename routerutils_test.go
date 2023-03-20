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
	"strconv"
	"sync"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/nephio-project/edge-watcher"
	"github.com/nephio-project/edge-watcher/preprocessor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type fakeEventSource struct {
	logger logr.Logger
	router preprocessor.EventRouter
}

func generateEvents(n int, eventType preprocessor.EventType,
	key preprocessor.RequestKey) (list []preprocessor.Event) {
	for i := 0; i < n; i++ {
		object := &unstructured.Unstructured{}
		object.SetName(strconv.Itoa(i))
		event := preprocessor.Event{
			Type:   eventType,
			Key:    key,
			Object: object,
		}
		list = append(list, event)
	}
	return
}

func (f *fakeEventSource) run(ctx context.Context, wg *sync.WaitGroup, events ...preprocessor.Event) {
	defer GinkgoRecover()
	defer wg.Done()

	logger := f.logger.WithName("fakeEventSource")
	debugLogger := logger.V(1)
	defer debugLogger.Info("exit")
	for _, e := range events {
		eventReq := &preprocessor.EventMessage{
			Event: e,
			Ctx:   ctx,
			Err:   make(chan error),
		}

		debugLogger.Info("sending event", "event", e)
		select {
		case <-ctx.Done():
			Fail("context cancelled")
		case f.router.RouteEvent(e.Key) <- eventReq:
		}
		debugLogger.Info("sent event", "event", e)

		select {
		case err := <-eventReq.Err:
			if err != nil {
				Fail(fmt.Sprintf("error received from eventReq: %v", err.Error()))
			}
		case <-ctx.Done():
			Fail("context cancelled")
		}
	}
}

type subscriber struct {
	router *edgewatcher.Router
	logger logr.Logger
}

func (t *subscriber) subscribe(ctx context.Context, opts edgewatcher.EventOptions, subscriberName ...string) (*edgewatcher.SubscriptionReq, error) {
	eventStream := make(chan preprocessor.Event)
	errStream := make(chan error)
	if len(subscriberName) == 0 {
		subscriberName = []string{uuid.NewString()}
	}
	subscribeReq := &edgewatcher.SubscriptionReq{
		Ctx:          ctx,
		Error:        errStream,
		EventOptions: opts,
		SubscriberInfo: edgewatcher.SubscriberInfo{
			SubscriberName: subscriberName[0],
			Channel:        eventStream,
		},
	}

	logger := t.logger.WithName("subscriber").WithValues("eventOptions", opts, "requestId", subscribeReq.SubscriberName)
	debugLogger := logger.V(1)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case t.router.Subscribe() <- subscribeReq:
	}
	Eventually(errStream).Should(Receive(nil))
	debugLogger.Info("sent subscribe request")
	return subscribeReq, nil
}

type receiveOptions struct {
	wg         *sync.WaitGroup
	events     []preprocessor.Event
	eventCount int
	requests   []*edgewatcher.SubscriptionReq
}

// receiveEvents mocks an edgewatcher subscriber which is listening from
// multiple streams
func (t *subscriber) receiveEvents(ctx context.Context, opts receiveOptions) {
	defer GinkgoRecover()
	defer opts.wg.Done()

	if len(opts.requests) == 0 {
		Fail("requests not provided to subscriber.receiveEvents")
	}

	logger := t.logger.WithName("subscriber").WithValues("subscriberName", opts.requests[0].SubscriberName, "eventCount", len(opts.events))
	debugLogger := logger.V(1)

	combinedStream := make(chan preprocessor.Event)

	go func() {
		defer close(combinedStream)
		var wg sync.WaitGroup
		wg.Add(len(opts.requests))
		for _, req := range opts.requests {
			go func(req *edgewatcher.SubscriptionReq) {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case <-req.Ctx.Done():
						return
					case event, ok := <-req.Channel:
						if !ok {
							//debugLogger.Info("stream closed")
							return
						}
						combinedStream <- event
					}
				}
			}(req)
		}
		wg.Wait()
	}()

	for i := 0; i < opts.eventCount; i++ {
		select {
		case e, ok := <-combinedStream:
			debugLogger.Info("received event", "event", e, "id", i)
			Expect(ok).To(BeTrue())
			Expect(opts.events).To(ContainElement(e))
		case <-ctx.Done():
			logger.Error(ctx.Err(), "context cancelled")
			Fail("context cancelled")
		}
	}

	debugLogger.Info("received the events")

	group, gctx := errgroup.WithContext(ctx)

	for _, req := range opts.requests {
		req := req
		group.Go(func() error {
			select {
			case <-gctx.Done():
				return gctx.Err()
			case t.router.Cancel() <- req:
			}
			debugLogger.Info("sent cancellation request", "eventOptions", req.EventOptions, "subscriberName", req.SubscriberName)
			select {
			case <-gctx.Done():
				return gctx.Err()
			case err := <-req.Error:
				if err != nil {
					logger.Error(err, "error received on cancellation request",
						"eventOptions", req.EventOptions, "subscriberName", req.SubscriberName)
				}
				return err
			}
		})
	}

	Expect(group.Wait()).To(BeNil())
}
