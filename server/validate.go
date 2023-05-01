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
	"fmt"

	pb "github.com/nephio-project/edge-watcher/protos"
)

func (server *edgeWatcherServer) validateRequestMetadata(
	metadata *pb.RequestMetadata) error {
	if metadata.Namespace == nil {
		return fmt.Errorf("RequestMetadata.Namespace not found")
	}

	if metadata.Kind == nil {
		return fmt.Errorf("RequestMetadata.Kind not found")
	}
	switch *metadata.Kind {
	case pb.CRDKind_UPFDeployment, pb.CRDKind_SMFDeployment:
	default:
		return fmt.Errorf("unknown RequestMetadata.Kind received: %v",
			*metadata.Kind)
	}

	if metadata.Group == nil {
		return fmt.Errorf("RequestMetadata.Group not found")
	}
	switch *metadata.Group {
	case pb.APIGroup_NFDeployNephioOrg:
	default:
		return fmt.Errorf("unknown RequestMetadata.Group received: %v",
			*metadata.Group)
	}

	if metadata.Version == nil {
		return fmt.Errorf("RequestMetadata.Version not found")
	}
	switch *metadata.Version {
	case pb.Version_v1alpha1:
	default:
		return fmt.Errorf("unknown RequestMetadata.Version received: %v",
			*metadata.Version)
	}

	return nil
}

func (server *edgeWatcherServer) validateEventRequest(req *pb.EventRequest) error {
	if req.Metadata == nil {
		return fmt.Errorf("EventRequest.Metadata not found")
	}

	if req.Metadata.Type == nil {
		return fmt.Errorf("EventRequest.Metadata.Type not found")
	}
	switch *req.Metadata.Type {
	case pb.EventType_Added, pb.EventType_Modified, pb.EventType_Deleted, pb.EventType_List:
	default:
		return fmt.Errorf("unknown EventRequest.Metadata.Type: %v",
			req.Metadata.Type)
	}

	if req.Metadata.Request == nil {
		return fmt.Errorf("EventRequest.Metadata.RequestMetadata not found")
	}

	if err := server.validateRequestMetadata(req.Metadata.Request); err != nil {
		return err
	}

	if req.Metadata.ClusterName == nil {
		return fmt.Errorf("EventRequest.Metadata.ClusterName not found")
	}

	if req.Metadata.NfdeployName == nil {
		return fmt.Errorf("EventRequest.Metadata.NfdeployName not found")
	}

	if req.EventTimestamp == nil {
		return fmt.Errorf("EventRequest.EventTimestamp not found")
	}

	if req.Object == nil || len(req.Object) == 0 {
		return fmt.Errorf("EventRequest.Object not found")
	}

	return nil
}
