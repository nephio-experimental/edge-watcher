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
	"fmt"

	pb "github.com/nephio-project/edge-watcher/protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (processor *EventPreprocessor) handleListEvent(ctx context.Context,
	req *pb.EventRequest) (pb.ResponseType, error) {
	key := getRequestKey(req.Metadata)
	logger := processor.logger.WithName("handleListEvent").WithValues("key", key)
	debugLogger := logger.V(1)

	obj, err := unmarshal(req.Object)
	if err != nil {
		logger.Error(err, "error in deserializing object")
		return pb.ResponseType_RESET, fmt.Errorf("error in deserializing object: %v", err)
	}

	if obj.GetName() == "" {
		debugLogger.Info("got empty list event")
	} else {
		debugLogger.Info("sending list event to router")
		if err := processor.sendEvent(ctx, key, List, obj, req.EventTimestamp.AsTime()); err != nil {
			logger.Error(err, "error in sending final list to router")
			return pb.ResponseType_RESET, status.Error(codes.Unavailable,
				fmt.Sprintf("error in publishing list event to subscribers: %v", err.Error()))
		}
		debugLogger.Info("sent final list to router")
	}

	listKey := getListKey(key)

	processor.listReceivedMu.Lock()
	defer processor.listReceivedMu.Unlock()
	processor.listReceived[listKey] = true
	debugLogger.Info("set listReceived")

	return pb.ResponseType_OK, nil
}
