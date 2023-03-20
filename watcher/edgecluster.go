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
	"github.com/nephio-project/common-lib/edge/porch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	edgeClusterGroup    = "cloud.nephio.org"
	edgeClusterVersion  = "v1alpha1"
	edgeClusterKind     = "EdgeCluster"
	edgeClusterResource = "edgeclusters"
	watcherAgentCRFile  = "watcheragent.yaml"
	porchPackageName    = "watcherAgent"
)

type Params struct {
	K8sDynamicClient     dynamic.Interface
	PorchClient          porch.Client
	PodIP                string
	Port                 string
	NephioNamespace      string
	EdgeClusterNamespace string
	Timestamp            time.Time
}

// CreateWatcherAgentCRs watches EdgeCluster resource in EdgeClusterNamespace
// and creates watcherAgentCR in the deploy-repo of each newly created
// EdgeCluster. It blocks until ctx is cancelled.
func CreateWatcherAgentCRs(ctx context.Context, logger logr.Logger,
	params Params) error {
	logger = logger.WithName("CreateWatcherAgentCRs")
	debugLogger := logger.V(1)

	queue, err := newInClusterWatcher(ctx, logger, params.K8sDynamicClient,
		schema.GroupVersionResource{
			Group:    edgeClusterGroup,
			Version:  edgeClusterVersion,
			Resource: edgeClusterResource,
		},
		schema.GroupVersionKind{
			Group:   edgeClusterGroup,
			Version: edgeClusterVersion,
			Kind:    edgeClusterKind,
		},
		params.EdgeClusterNamespace,
	)

	if err != nil {
		logger.Error(err, "error in starting inClusterWatcher")
		return err
	}

	debugLogger.Info("created inClusterWatcher")

	for {
		debugLogger.Info("getting new edgecluster object")

		objIntf, shutdown := queue.Get()
		if shutdown {
			logger.Info("edgecluster queue closed")
			break
		}
		data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(objIntf)
		if err != nil {
			logger.Error(err, "error in converting received object to unstructured.Unstructured")
			queue.Forget(objIntf)
			continue
		}

		debugLogger.Info("converted received object to unstructured.Unstructured")

		obj := &unstructured.Unstructured{Object: data}

		watcherAgentYAML, err := GetYAML(params.PodIP, params.Port, obj.GetName(),
			params.NephioNamespace, params.Timestamp)
		if err != nil {
			logger.Error(err, "error in creating watcherAgent yaml")
			return err
		}

		debugLogger.Info("created watcherAgentYaml",
			"edgeClusterName", obj.GetName(), "edgeClusterNamespace",
			obj.GetNamespace())

		err = params.PorchClient.ApplyPackage(ctx, map[string]string{
			watcherAgentCRFile: watcherAgentYAML,
		}, porchPackageName, obj.GetName())

		if err != nil {
			logger.Error(err, "error in creating porch package",
				"edgeClusterName", obj.GetName(), "edgeClusterNamespace",
				obj.GetNamespace())
			queue.AddRateLimited(obj)
		}

		debugLogger.Info("created porch package",
			"edgeClusterName", obj.GetName(), "edgeClusterNamespace",
			obj.GetNamespace())

		queue.Done(obj)
	}

	logger.Info("closing inClusterWatcher")
	return nil
}
