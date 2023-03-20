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

	"github.com/nephio-project/edge-watcher/preprocessor"
)

type SubscriptionType string

const (
	ClusterSubscriber  SubscriptionType = "clusterSubscriber"
	NfDeploySubscriber SubscriptionType = "nfdeploySubscriber"
)

// EventOptions is used for filtering the edge events based on either cluster or
// parent NfDeploy
type EventOptions struct {
	Type SubscriptionType
	// The name of the subscriptionType, which could be a clusterName or NFDeploy
	// name.
	SubscriptionName string
}

// EventPublisher publishes k8s events received from the edgeclusters. It allows
// the clients to subscribe to events using SubscriptionReq.
type EventPublisher interface {
	// Subscribe returns a channel to receive SubscriptionReq from clients
	Subscribe() chan *SubscriptionReq
	// Cancel returns a channel to receive cancellation request in form of
	// SubscriptionReq from clients. Request should have SubscriptionReq.Ctx,
	// SubscriptionReq.Error and SubscriptionReq.SubscriberInfo.SubscriberName.
	Cancel() chan *SubscriptionReq
}

// SubscriberInfo uniquely identifies a subscriber
type SubscriberInfo struct {
	// SubscriberName is used for uniquely identifying the subscriber
	SubscriberName string
	// Channel is used for sending the preprocessor.Event to the subscriber
	Channel chan preprocessor.Event

	// stop is used by edgwatcher determine if the subscription is still valid
	stop chan struct{}
}

type SubscriptionReq struct {
	Ctx context.Context
	// Error is used to send any non-retryable error occurred during the
	// subscription/cancellation process and nil is sent in case of successful
	// subscription/cancellation.
	Error chan error

	EventOptions
	SubscriberInfo
}
