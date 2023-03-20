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

package watcher

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// default resync duration for informer from http://shortn/_kSVzCXsNs5
const defaultResyncTime = 10 * time.Hour

func newInClusterWatcher(ctx context.Context, logger logr.Logger,
	client dynamic.Interface, gvr schema.GroupVersionResource,
	gvk schema.GroupVersionKind, namespace string) (
	workqueue.RateLimitingInterface, error) {
	logger = logger.WithName("InClusterWatcher").WithName("Setup")
	debugLogger := logger.V(1)

	debugLogger.Info("initializing ListWatcher", "gvr", gvr,
		"namespace", namespace)
	lw := &cache.ListWatch{
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return client.Resource(gvr).Namespace(namespace).Watch(ctx, options)
		},
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return client.Resource(gvr).Namespace(namespace).List(ctx, options)
		},
	}

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	go func() {
		<-ctx.Done()
		logger.Info("closing workqueue")
		queue.ShutDown()
	}()

	exampleObject := &unstructured.Unstructured{}
	exampleObject.SetGroupVersionKind(gvk)

	informer := cache.NewSharedInformer(lw, exampleObject, defaultResyncTime)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
			if err != nil {
				logger.Error(err, "error in deserializing")
				return
			}
			u := unstructured.Unstructured{Object: data}
			debugLogger.Info("got object", u.GetName(), u.GetNamespace())
			queue.Add(obj)
		},
	})

	go informer.Run(ctx.Done())

	debugLogger.Info("started informer")

	return queue, nil
}
