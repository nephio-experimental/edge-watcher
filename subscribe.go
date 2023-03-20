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

func (router *Router) createSubscription(req *SubscriptionReq,
	subscribers map[EventOptions]subscriberSet) (
	streams subscriberSet, err error) {
	logger := router.log.WithName("createSubscription")
	debugLogger := logger.V(1)
	logger.Info("received subscription request", "subscriberName", req.SubscriberName,
		"requestId", req.SubscriptionName, "EventsOptions", req.EventOptions)

	if req.Error == nil {
		err := fmt.Errorf("nil error channel received with subscribe request id: %v, EventOptions: %v", req.SubscriberName, req.EventOptions)
		logger.Error(err, "invalid subscribe request")
		return nil, err
	}

	if req.Ctx == nil {
		err := fmt.Errorf("context not provided")
		logger.Error(err, "invalid subscribe request")
		select {
		case req.Error <- err:
		case <-router.Ctx.Done():
		}
		return nil, err
	}

	if req.Type != ClusterSubscriber && req.Type != NfDeploySubscriber {
		// TODO: use typed error
		err := fmt.Errorf("unknown subscriber type: %v", req.Type)
		logger.Error(err, "invalid subscribe request")
		return nil, router.sendError(req, err)
	}

	if req.Channel == nil {
		// TODO: use typed error
		err := fmt.Errorf("nil channel received in the request")
		logger.Error(err, "invalid subscribe request")
		return nil, router.sendError(req, err)
	}

	streams, ok := subscribers[req.EventOptions]
	if !ok {
		streams = make(map[string]SubscriberInfo)
		subscribers[req.EventOptions] = streams
	}

	if _, ok := streams.find(req.SubscriberInfo.SubscriberName); ok {
		// TODO: use typed error
		err := fmt.Errorf("duplicate subscription request, SubscriberName: %v", req.SubscriberInfo.SubscriberName)
		logger.Error(err, "invalid subscribe request")
		return nil, router.sendError(req, err)
	}

	req.stop = make(chan struct{})

	streams.add(req.SubscriberInfo.SubscriberName, req.SubscriberInfo)
	debugLogger.Info("added the subscriber", "requestId", req.SubscriberName, "EventsOptions", req.EventOptions)
	return streams.deepCopy(), nil
}
