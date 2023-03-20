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

package preprocessor

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	pb "github.com/nephio-project/edge-watcher/protos"
	"k8s.io/apimachinery/pkg/runtime"
)

// EventPreprocessor identifies the  EventType  for incoming events and routes them
// to the EventRouter. In case of List events, it buffers constituting list
// pages and provides the router with consolidated list events
type EventPreprocessor struct {
	logger logr.Logger

	listReceivedMu sync.RWMutex
	listReceived   map[listKey]bool

	router EventRouter
}

type EventRouter interface {
	RouteEvent(key RequestKey) chan *EventMessage
}

type EventReceiver interface {
	ReceiveEvent(ctx context.Context, req *pb.EventRequest) (pb.ResponseType, error)
}

var _ EventReceiver = &EventPreprocessor{}

// New instantiates a new EventPreprocessor which passes the event to a router object
func New(logger logr.Logger, router EventRouter) *EventPreprocessor {
	return &EventPreprocessor{
		listReceived: make(map[listKey]bool),
		router:       router,
		logger:       logger.WithName("EventPreprocessor"),
	}
}

// ReceiveEvent receives pb.EventRequest and passes it to watch/list handler
func (processor *EventPreprocessor) ReceiveEvent(ctx context.Context,
	req *pb.EventRequest) (pb.ResponseType, error) {
	key := getRequestKey(req.Metadata)

	logger := processor.logger.WithName("ReceiveEvent").WithValues("key", key)
	debugLogger := logger.V(1)

	switch *req.Metadata.Type {
	case pb.EventType_List:
		debugLogger.Info("list event received")
		resp, err := processor.handleListEvent(ctx, req)
		if err != nil {
			logger.Error(err, "error in handling list event")
		}
		return resp, err
	case pb.EventType_Added, pb.EventType_Modified, pb.EventType_Deleted:
		debugLogger.Info("watch event received")
		resp, err := processor.handleWatchEvent(ctx, req)
		if err != nil {
			logger.Error(err, "error in handling list event")
		}
		return resp, err
	}

	return pb.ResponseType_RESET, nil
}

func (processor *EventPreprocessor) sendEvent(ctx context.Context, key RequestKey,
	eventType EventType, obj runtime.Object, timestamp time.Time) error {
	logger := processor.logger.WithName("sendEvent")
	debugLogger := logger.V(1)

	debugLogger.Info("sending event to router", "key", key, "eventType", eventType)
	event := &EventMessage{
		Ctx: ctx,
		Err: make(chan error, 10),

		Event: Event{
			Type:      eventType,
			Timestamp: timestamp,
			Key:       key,
			Object:    obj,
		},
	}

	select {
	case processor.router.RouteEvent(key) <- event:
	case <-ctx.Done():
		logger.Error(ctx.Err(), "context cancelled before event is sent")
		return ctx.Err()
	}

	debugLogger.Info("sent event to router", "key", key, "eventType", eventType)

	select {
	case err := <-event.Err:
		if err != nil {
			logger.Error(err, "error received from router", "key", key, "eventType", eventType)
		}
		return err
	case <-ctx.Done():
		logger.Error(ctx.Err(), "context cancelled before acknowledgement is received")
		return ctx.Err()
	}
}
