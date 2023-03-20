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

package server_test

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	pb "github.com/nephio-project/edge-watcher/protos"
	. "github.com/onsi/ginkgo/v2"
)

func getPtr[T pb.EventType | pb.CRDKind | pb.APIGroup | pb.Version | string](x T) *T {
	return &x
}

type filterResponse struct {
	resp pb.ResponseType
	err  error
}

type fakeReceiver struct {
	ctx    context.Context
	logger logr.Logger

	requestStream  chan *pb.EventRequest
	responseStream chan filterResponse
}

type fakeClient struct {
	pb.WatcherServiceClient
}

func (receiver *fakeReceiver) ReceiveEvent(ctx context.Context, req *pb.EventRequest) (pb.ResponseType, error) {
	defer GinkgoRecover()
	logger := receiver.logger.WithName("ReceiveEvent").WithValues()
	debugLogger := logger.V(1)

	debugLogger.Info("sending request on requestStream")
	select {
	case <-ctx.Done():
		Fail(fmt.Sprintf("request context cancelled: %v", ctx.Err()))
	case <-receiver.ctx.Done():
		Fail(fmt.Sprintf("receiver context cancelled: %v", receiver.ctx.Err()))
	case receiver.requestStream <- req:
		debugLogger.Info("sent request on requestStream")
	}

	debugLogger.Info("waiting for response on responseStream")
	select {
	case <-ctx.Done():
		Fail(fmt.Sprintf("request context cancelled: %v", ctx.Err()))
	case <-receiver.ctx.Done():
		Fail(fmt.Sprintf("receiver context cancelled: %v", receiver.ctx.Err()))
	case resp, ok := <-receiver.responseStream:
		if !ok {
			Fail("responseStream closed")
		}
		debugLogger.Info("response received", "resp", resp.resp, "error", resp.err)
		return resp.resp, resp.err
	}

	return 0, nil
}
