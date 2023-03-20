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

package edgewatcher

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/go-logr/logr"
	"github.com/nephio-project/edge-watcher/preprocessor"
)

type subscriberSet map[string]SubscriberInfo // maps SubscriberName to SubscriberInfo

func (s subscriberSet) add(key string, val SubscriberInfo) {
	s[key] = val
}

func (s subscriberSet) find(key string) (val SubscriberInfo, _ bool) {
	val, ok := s[key]
	return val, ok
}

func (s subscriberSet) remove(key string) {
	delete(s, key)
}

func (s subscriberSet) deepCopy() subscriberSet {
	ret := make(map[string]SubscriberInfo)
	for key, val := range s {
		ret[key] = val
	}
	return ret
}

// Router maintains subscriptions for receiving events
type Router struct {
	Ctx context.Context

	subscribers sync.Map // map[EventOptions]subscriberSet

	// subscriptions channel is exposed to edgewatcher's clients to receive
	// subscription requests
	subscriptions chan *SubscriptionReq

	// cancellations channel is exposed to edgewatcher's clients to receive
	// cancellation requests
	cancellations chan *SubscriptionReq

	log logr.Logger

	eventStreams sync.Map // map[preprocessor.RequestKey]chan *preprocessor.EventMessage

}

func (router *Router) Subscribe() chan *SubscriptionReq {
	return router.subscriptions
}

func (router *Router) Cancel() chan *SubscriptionReq {
	return router.cancellations
}

var _ preprocessor.EventRouter = &Router{}

var _ EventPublisher = &Router{}

// RouteEvent returns the appropriate channel for sending the events to the
// router based on the RequestKey
func (router *Router) RouteEvent(key preprocessor.RequestKey) (
	stream chan *preprocessor.EventMessage) {
	logger := router.log.WithName("RouteEvent")
	debugLogger := logger.V(1)

	if streamIntf, ok := router.eventStreams.Load(key); !ok {
		stream = make(chan *preprocessor.EventMessage)
		router.eventStreams.Store(key, stream)
		go router.route(key, stream)
		debugLogger.Info("started listening events", "key", key)
	} else {
		stream, ok = streamIntf.(chan *preprocessor.EventMessage)
		if !ok {
			panic(fmt.Errorf("value of type:%v found in Router.eventStreams",
				reflect.TypeOf(streamIntf)))
		}
	}
	return
}

func (router *Router) route(key preprocessor.RequestKey, source chan *preprocessor.EventMessage) {
	logger := router.log.WithName("route").WithValues("RequestKey", key)
	debugLogger := logger.V(1)

	defer func() {
		debugLogger.Info("exit")
	}()
loop:
	for {
		select {
		case <-router.Ctx.Done():
			return
		case event, ok := <-source:
			if !ok {
				debugLogger.Info("event source stream closed")
				router.eventStreams.Delete(key)
				return
			}

			select {
			case <-event.Ctx.Done():
				continue loop
			default:
			}

			debugLogger.Info("event received", "eventType", event.Type,
				"objectKey", event.Key, "object", event.Object)

			err := router.publishEvent(event.Ctx, event.Event)
			if err != nil {
				logger.Error(err, "error in publishing the event")
			} else {
				debugLogger.Info("published event", "event", event.Event)
			}

			select {
			case <-event.Ctx.Done():
			case <-router.Ctx.Done():
			case event.Err <- err:
			}
		}
	}
}

func (router *Router) manageSubscriptions() {
	logger := router.log.WithName("manageSubscriptions")
	subscribers := make(map[EventOptions]subscriberSet)

	for {
		select {
		case <-router.Ctx.Done():
			return
		case req := <-router.subscriptions:
			streams, err := router.createSubscription(req, subscribers)
			if err == nil {
				router.subscribers.Store(req.EventOptions, streams)
				if err := router.sendError(req, nil); err != nil {
					logger.Error(err, "unable to send nil err to subscriber")
				}
			}
		case req := <-router.cancellations:
			streams, err := router.cancelSubscription(req, subscribers)
			if err == nil {
				router.subscribers.Store(req.EventOptions, streams)
			}
		}
	}
}

// NewRouter returns an initialised Router object which can accept subscriptions
// to get events
func NewRouter(logger logr.Logger, ctx context.Context) *Router {
	r := &Router{
		subscriptions: make(chan *SubscriptionReq),
		cancellations: make(chan *SubscriptionReq),
		Ctx:           ctx,
		log:           logger.WithName("Router"),
	}
	go r.manageSubscriptions()
	return r
}
