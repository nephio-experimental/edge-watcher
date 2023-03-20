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
	"time"

	"github.com/go-logr/logr"
	"github.com/nephio-project/common-lib/edge/porch"
	"github.com/nephio-project/edge-watcher/preprocessor"
	pb "github.com/nephio-project/edge-watcher/protos"
	"github.com/nephio-project/edge-watcher/server"
	"github.com/nephio-project/edge-watcher/watcher"
	"google.golang.org/grpc"
	"k8s.io/client-go/dynamic"
)

type Params struct {
	GRPCServer           *grpc.Server      `ignored:"true"`
	K8sDynamicClient     dynamic.Interface `ignored:"true"`
	PorchClient          porch.Client      `ignored:"true"`
	PodIP                string            `required:"true" envconfig:"POD_IP"`
	Port                 string            `required:"true" envconfig:"GRPC_PORT"`
	NephioNamespace      string            `required:"true" envconfig:"NEPHIO_NAMESPACE"`
	EdgeClusterNamespace string            `required:"true" split_words:"true"`
}

// New initializes and registers EdgeWatcherService with the given grpc.Server
func New(ctx context.Context, logger logr.Logger, params Params) (EventPublisher, error) {
	watcherParams := watcher.Params{
		K8sDynamicClient:     params.K8sDynamicClient,
		PorchClient:          params.PorchClient,
		PodIP:                params.PodIP,
		Port:                 params.Port,
		NephioNamespace:      params.NephioNamespace,
		EdgeClusterNamespace: params.EdgeClusterNamespace,
		Timestamp:            time.Now(),
	}

	go watcher.CreateWatcherAgentCRs(ctx, logger, watcherParams)

	router := NewRouter(logger, ctx)

	eventFilter := preprocessor.New(logger, router)

	pb.RegisterWatcherServiceServer(params.GRPCServer,
		server.NewEdgeWatcherServer(logger, eventFilter))

	return router, nil
}
