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

func (processor *EventPreprocessor) handleWatchEvent(ctx context.Context,
	req *pb.EventRequest) (pb.ResponseType, error) {
	key := getRequestKey(req.Metadata)

	logger := processor.logger.WithName("handleWatchEvent").
		WithValues("key", key)
	debugLogger := logger.V(1)

	listKey := getListKey(key)

	processor.listReceivedMu.RLock()
	isListReceived, ok := processor.listReceived[listKey]
	processor.listReceivedMu.RUnlock()

	if !ok || !isListReceived {
		err := fmt.Errorf("list not received")
		logger.Error(err, "error processing watch event")
		return pb.ResponseType_RESET, fmt.Errorf("list not received")
	}

	obj, err := unmarshal(req.Object)
	if err != nil {
		logger.Error(err, "error in deserializing object")
		// TODO: send RESET after continuous drops
		return pb.ResponseType_OK,
			fmt.Errorf("error in deserializing object: %v", err)
	}

	switch *req.Metadata.Type {
	case pb.EventType_Added:
		err = processor.sendEvent(ctx, key, Added, obj, req.EventTimestamp.AsTime())
	case pb.EventType_Modified:
		err = processor.sendEvent(ctx, key, Modified, obj, req.EventTimestamp.AsTime())
	case pb.EventType_Deleted:
		err = processor.sendEvent(ctx, key, Deleted, obj, req.EventTimestamp.AsTime())
	}

	if err != nil {
		logger.Error(err, "error in publishing event to subscribers")
		// TODO: use typed error

		return pb.ResponseType_RESET, status.Error(codes.Unavailable,
			fmt.Sprintf("error in publishing watch event to subscribers: %v",
				err.Error()))
	}

	debugLogger.Info("successfully sent watch event to router")
	return pb.ResponseType_OK, nil
}
