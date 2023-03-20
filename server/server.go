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

package server

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/nephio-project/edge-watcher/preprocessor"
	pb "github.com/nephio-project/edge-watcher/protos"
)

const EventTimeoutSeconds = 5

type edgeWatcherServer struct {
	pb.UnsafeWatcherServiceServer
	filter preprocessor.EventReceiver
	log    logr.Logger
}

func (server *edgeWatcherServer) ReportEvent(ctx context.Context,
	request *pb.EventRequest) (*pb.EventResponse, error) {
	logger := server.log.WithName("ReportEvent")
	debugLogger := logger.V(1)

	debugLogger.Info("event received",
		"eventMetadata", request.Metadata,
		"eventObject", request.Object)
	ctx, cancel := context.WithTimeout(ctx, EventTimeoutSeconds*time.Second)
	defer cancel()

	if err := server.validateEventRequest(request); err != nil {
		return nil, err
	}

	response, err := server.filter.ReceiveEvent(ctx, request)
	if err != nil {
		return nil, err
	}
	return &pb.EventResponse{
		Response: &response,
	}, nil
}

func NewEdgeWatcherServer(log logr.Logger, filter preprocessor.EventReceiver) *edgeWatcherServer {
	return &edgeWatcherServer{
		log:    log.WithName("EdgeWatcherServer"),
		filter: filter,
	}
}
