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
	"fmt"
)

func (router *Router) sendError(req *SubscriptionReq, err error) error {
	logger := router.log.WithName("sendError").
		WithValues("requestId", req.SubscriptionName,
			"EventsOptions", req.EventOptions,
			"error", err)
	select {
	case <-router.Ctx.Done():
		logger.Error(router.Ctx.Err(), "router context cancelled before sending error")
		if err == nil {
			return router.Ctx.Err()
		}
	case <-req.Ctx.Done():
		logger.Error(router.Ctx.Err(), "request context cancelled before sending error")
		if err == nil {
			return req.Ctx.Err()
		}
	case req.Error <- err:
		logger.V(1).Info("error sent")
	}
	return err
}

func (router *Router) cancelSubscription(req *SubscriptionReq,
	subscribers map[EventOptions]subscriberSet) (
	streams subscriberSet, err error) {
	logger := router.log.WithName("removeSubscriber")
	debugLogger := logger.V(1)

	logger.Info("received cancellation req", "subscriberName", req.SubscriberName,
		"requestId", req.SubscriptionName, "EventsOptions", req.EventOptions)

	if req.Error == nil {
		err := fmt.Errorf("nil error channel received with subscribe request id: %v, EventOptions: %v", req.SubscriberName, req.EventOptions)
		logger.Error(err, "invalid cancellation request")
		return nil, err
	}

	if req.Ctx == nil {
		err := fmt.Errorf("context not provided")
		logger.Error(err, "invalid cancellation request")
		select {
		case req.Error <- err:
		case <-router.Ctx.Done():
		}
		return nil, err
	}

	if req.Type != ClusterSubscriber && req.Type != NfDeploySubscriber {
		// TODO: use typed error
		err := fmt.Errorf("unknown subscriber type: %v", req.Type)
		logger.Error(err, "invalid cancellation request")
		return nil, router.sendError(req, err)
	}

	streams, ok := subscribers[req.EventOptions]
	if !ok {
		// TODO: use typed error
		err := fmt.Errorf("unknown cancel request recieved: %v , options: %v", req.SubscriberName, req.EventOptions)
		logger.Error(err, "invalid cancellation request")
		return nil, router.sendError(req, err)
	}

	subscribeReq, ok := streams.find(req.SubscriberInfo.SubscriberName)
	if !ok {
		// TODO: use typed error
		err := fmt.Errorf("unknown stream received")
		logger.Error(err, "invalid cancellation request")

		return nil, router.sendError(req, err)
	}

	if err := router.sendError(req, nil); err != nil {
		return nil, err
	}

	streams.remove(req.SubscriberInfo.SubscriberName)

	close(subscribeReq.stop)

	debugLogger.Info("removed from subscribers", "requestId", req.SubscriberName, "EventsOptions", req.EventOptions)

	return streams.deepCopy(), nil
}
