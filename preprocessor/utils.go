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
	pb "github.com/nephio-project/edge-watcher/protos"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func getRequestKey(metadata *pb.Metadata) RequestKey {
	key := RequestKey{
		ClusterName: *metadata.ClusterName,
		NFDeploy:    *metadata.NfdeployName,
		Namespace:   *metadata.Request.Namespace,
	}

	switch *metadata.Request.Group {
	case pb.APIGroup_NFDeployNephioOrg:
		key.Group = "workload.nephio.org"
	}

	switch *metadata.Request.Version {
	case pb.Version_v1alpha1:
		key.Version = "v1alpha1"
	}

	switch *metadata.Request.Kind {
	case pb.CRDKind_SMFDeployment:
		key.Kind = "SMFDeployment"
	case pb.CRDKind_UPFDeployment:
		key.Kind = "UPFDeployment"
	case pb.CRDKind_AMFDeployment:
		key.Kind = "AMFDeployment"
	}
	return key
}

func getListKey(requestKey RequestKey) listKey {
	return listKey{
		clusterName: requestKey.ClusterName,
		group:       requestKey.Group,
		version:     requestKey.Version,
		kind:        requestKey.Kind,
		namespace:   requestKey.Namespace,
	}
}

func unmarshal(data []byte) (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}
	if err := u.UnmarshalJSON(data); err != nil {
		return nil, err
	}
	return u, nil
}
