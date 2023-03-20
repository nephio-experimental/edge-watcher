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
	"net"

	pb "github.com/nephio-project/edge-watcher/protos"
	"github.com/nephio-project/edge-watcher/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const bufSize = 1024 * 1024

var _ = Describe("Server", func() {
	var ctx context.Context
	var cancel context.CancelFunc
	var client *fakeClient
	var request *pb.EventRequest
	var filter *fakeReceiver
	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		listener := bufconn.Listen(bufSize)
		logger := zap.New(func(options *zap.Options) {
			options.Development = true
			options.DestWriter = GinkgoWriter
		})

		filter = &fakeReceiver{
			ctx:    ctx,
			logger: logger.WithName("fakeReceiver"),

			requestStream:  make(chan *pb.EventRequest),
			responseStream: make(chan filterResponse),
		}

		s := grpc.NewServer()
		pb.RegisterWatcherServiceServer(s, server.NewEdgeWatcherServer(logger, filter))

		go func() {
			<-ctx.Done()
			s.GracefulStop()
		}()

		go func() {
			defer GinkgoRecover()
			err := s.Serve(listener)
			Expect(err).To(BeNil())
		}()

		conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(
			func(ctx context.Context, addr string) (net.Conn, error) {
				return listener.Dial()
			}), grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).To(BeNil())

		go func() {
			<-ctx.Done()
			conn.Close()
		}()
		client = &fakeClient{
			WatcherServiceClient: pb.NewWatcherServiceClient(conn),
		}

		u := &unstructured.Unstructured{}
		u.SetName("object1")

		objectJSON, err := u.MarshalJSON()
		Expect(err).To(BeNil())
		request = &pb.EventRequest{
			Metadata: &pb.Metadata{
				Type: getPtr(pb.EventType_Added),
				Request: &pb.RequestMetadata{
					Namespace: getPtr("upf"),
					Kind:      getPtr(pb.CRDKind_UPFDeploy),
					Group:     getPtr(pb.APIGroup_NFDeployNephioOrg),
					Version:   getPtr(pb.Version_v1alpha1),
				},
				ClusterName:  getPtr("cluster1"),
				NfdeployName: getPtr("nfdeploy1"),
			},
			EventTimestamp: timestamppb.Now(),
			Object:         objectJSON,
		}
	})

	AfterEach(func() {
		cancel()
	})

	Describe("invalid requests", func() {
		Context("with EventRequest.Metadata missing", func() {
			It("should return the error", func() {
				request.Metadata = nil
				_, err := client.ReportEvent(ctx, request)
				st, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(st.Message()).To(Equal("EventRequest.Metadata not found"))
			})
		})
		Context("with EventRequest.Metadata.Type missing", func() {
			It("should return the error", func() {
				request.Metadata.Type = nil
				_, err := client.ReportEvent(ctx, request)
				st, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(st.Message()).To(Equal("EventRequest.Metadata.Type not found"))
			})
		})
		Context("with unknown EventRequest.Metadata.Type", func() {
			It("should return the error", func() {
				request.Metadata.Type = getPtr(pb.EventType(10))
				_, err := client.ReportEvent(ctx, request)
				st, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(st.Message()).To(Equal("unknown EventRequest.Metadata.Type: 10"))
			})
		})
		Context("with EventRequest.Metadata.RequestMetadata missing", func() {
			It("should return the error", func() {
				request.Metadata.Request = nil
				_, err := client.ReportEvent(ctx, request)
				st, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(st.Message()).To(Equal("EventRequest.Metadata.RequestMetadata not found"))
			})
		})
		Context("with EventRequest.Metadata.RequestMetadata.Namespace missing", func() {
			It("should return the error", func() {
				request.Metadata.Request.Namespace = nil
				_, err := client.ReportEvent(ctx, request)
				st, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(st.Message()).To(Equal("RequestMetadata.Namespace not found"))
			})
		})
		Context("with unknown RequestMetadata.Kind", func() {
			It("should return the error", func() {
				request.Metadata.Request.Kind = getPtr(pb.CRDKind(10))
				_, err := client.ReportEvent(ctx, request)
				st, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(st.Message()).To(Equal("unknown RequestMetadata.Kind received: 10"))
			})
		})
		Context("with unknown RequestMetadata.Group", func() {
			It("should return the error", func() {
				request.Metadata.Request.Group = getPtr(pb.APIGroup(10))
				_, err := client.ReportEvent(ctx, request)
				st, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(st.Message()).To(Equal("unknown RequestMetadata.Group received: 10"))
			})
		})
		Context("with unknown RequestMetadata.Version", func() {
			It("should return the error", func() {
				request.Metadata.Request.Version = getPtr(pb.Version(10))
				_, err := client.ReportEvent(ctx, request)
				st, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(st.Message()).To(Equal("unknown RequestMetadata.Version received: 10"))
			})
		})
		Context("with EventRequest.Metadata.ClusterName missing", func() {
			It("should return the error", func() {
				request.Metadata.ClusterName = nil
				_, err := client.ReportEvent(ctx, request)
				st, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(st.Message()).To(Equal("EventRequest.Metadata.ClusterName not found"))
			})
		})
		Context("with EventRequest.Metadata.NfdeployName missing", func() {
			It("should return the error", func() {
				request.Metadata.NfdeployName = nil
				_, err := client.ReportEvent(ctx, request)
				st, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(st.Message()).To(Equal("EventRequest.Metadata.NfdeployName not found"))
			})
		})
		Context("with EventRequest.EventTimestamp missing", func() {
			It("should return the error", func() {
				request.EventTimestamp = nil
				_, err := client.ReportEvent(ctx, request)
				st, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(st.Message()).To(Equal("EventRequest.EventTimestamp not found"))
			})
		})

		Context("with EventRequest.Object missing", func() {
			It("should return the error", func() {
				request.Object = nil
				_, err := client.ReportEvent(ctx, request)
				st, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(st.Message()).To(Equal("EventRequest.Object not found"))
			})
		})
	})
	Describe("valid requests", func() {
		Context("filter returns error", func() {
			It("should send error", func() {
				go func() {
					select {
					case <-ctx.Done():
					case filter.responseStream <- filterResponse{err: fmt.Errorf("fake error")}:
					}
				}()
				go func() {
					select {
					case <-ctx.Done():
					case <-filter.requestStream:
					}
				}()
				_, err := client.ReportEvent(ctx, request)
				st, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(st.Message()).To(Equal("fake error"))
			})
		})
		Context("filter doesn't return error", func() {
			It("should return the correct Response", func() {
				go func() {
					select {
					case <-ctx.Done():
					case filter.responseStream <- filterResponse{resp: pb.ResponseType_OK, err: nil}:
					}
				}()

				go func() {
					select {
					case <-ctx.Done():
					case <-filter.requestStream:
					}
				}()
				resp, err := client.ReportEvent(ctx, request)
				Expect(err).To(BeNil())
				Expect(resp.Response).To(PointTo(Equal(pb.ResponseType_OK)))
			})
		})
	})
})
