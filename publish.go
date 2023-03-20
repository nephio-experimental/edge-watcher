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

	"github.com/nephio-project/edge-watcher/preprocessor"
	"golang.org/x/sync/errgroup"
)

func (router *Router) getSubscribers(key EventOptions) (
	subscribers subscriberSet, exists bool) {
	valIntf, ok := router.subscribers.Load(key)
	if !ok {
		return nil, false
	}

	val, ok := valIntf.(subscriberSet)
	if !ok {
		panic(fmt.Errorf("unknown value of type: %v found in Router.subscriber",
			reflect.TypeOf(valIntf)))
	}

	return val, true
}

func (router *Router) publishEvent(ctx context.Context, e preprocessor.Event) error {
	logger := router.log.WithName("publishEvent").WithValues("ObjectKey", e.Key)
	debugLogger := logger.V(1)

	clusterSubscriberOption := EventOptions{
		Type:             ClusterSubscriber,
		SubscriptionName: e.Key.ClusterName,
	}

	nfDeploySubscriberOption := EventOptions{
		Type:             NfDeploySubscriber,
		SubscriptionName: e.Key.NFDeploy,
	}

	clusterStreams, clusterSubscribers := router.getSubscribers(clusterSubscriberOption)

	group, groupCtx := errgroup.WithContext(ctx)

	if clusterSubscribers {
		debugLogger.Info("send to cluster subscribers", "event", e, "options", clusterSubscriberOption)
		for _, stream := range clusterStreams {
			stream := stream
			group.Go(func() error {
				debugLogger.Info("sending to cluster subscriber", "event", e, "subscriberId", stream.SubscriberName)
				select {
				case <-stream.stop:
					debugLogger.Info("subscription cancelled", "subscriberId", stream.SubscriberName)
					return nil
				case <-groupCtx.Done():
					return groupCtx.Err()
				case <-router.Ctx.Done():
					return router.Ctx.Err()
				case stream.Channel <- e:
					debugLogger.Info("sent", "event", e, "subscriberId", stream.SubscriberName)
					return nil
				}
			})
		}
	}

	nfDeployStreams, nfDeploySubscribers := router.getSubscribers(nfDeploySubscriberOption)

	if nfDeploySubscribers {
		debugLogger.Info("send to nfdeploy subscribers", "event", e, "options", nfDeploySubscriberOption)
		for _, stream := range nfDeployStreams {
			stream := stream
			group.Go(func() error {
				debugLogger.Info("sending to nfdeploy subscriber", "event", e,
					"subscriberId", stream.SubscriberName)
				select {
				case <-stream.stop:
					debugLogger.Info("subscription cancelled", "subscriberId", stream.SubscriberName)
					return nil
				case <-groupCtx.Done():
					return groupCtx.Err()
				case <-router.Ctx.Done():
					return router.Ctx.Err()
				case stream.Channel <- e:
					debugLogger.Info("sent", "event", e, "subscriberId", stream.SubscriberName)
					return nil
				}
			})
		}
	}

	if err := group.Wait(); err != nil {
		logger.Error(err, "error in publishing the event", "event", e)
		return err
	}

	if !clusterSubscribers && !nfDeploySubscribers {
		return fmt.Errorf("no subscribers found for the event: %v", e)
	}

	return nil

}
