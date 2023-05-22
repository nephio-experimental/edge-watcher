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

package preprocessor_test

import (
	"time"

	"github.com/nephio-project/edge-watcher/preprocessor"
	pb "github.com/nephio-project/edge-watcher/protos"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var smfDeployGVK = schema.GroupVersionKind{
	Group:   "workload.nephio.org",
	Version: "v1alpha1",
	Kind:    "SMFDeployment",
}

var amfDeployGVK = schema.GroupVersionKind{
	Group:   "workload.nephio.org",
	Version: "v1alpha1",
	Kind:    "AMFDeployment",
}

type event struct {
	key        preprocessor.RequestKey
	event      *pb.EventRequest
	finalEvent *preprocessor.Event
}

type objectParams struct {
	name      string
	namespace string
	gvk       schema.GroupVersionKind
}

func getPtr[T string | int32 | bool | pb.CRDKind | pb.APIGroup | pb.Version | pb.EventType](x T) *T {
	return &x
}

var timestamp = time.Now()

var eventLists = map[string]event{
	"smfdeploy list": {
		key: preprocessor.RequestKey{
			ClusterName: "cluster1",
			NFDeploy:    "nfdeploy1",
			Namespace:   "namespace1",
			Group:       "workload.nephio.org",
			Version:     "v1alpha1",
			Kind:        "SMFDeployment",
		},
		event: &pb.EventRequest{
			Metadata: &pb.Metadata{
				Type: getPtr(pb.EventType_List),
				Request: &pb.RequestMetadata{
					Namespace: getPtr("namespace1"),
					Kind:      getPtr(pb.CRDKind_SMFDeployment),
					Group:     getPtr(pb.APIGroup_NFDeployNephioOrg),
					Version:   getPtr(pb.Version_v1alpha1),
				},
				ClusterName:  getPtr("cluster1"),
				NfdeployName: getPtr("nfdeploy1"),
			},
			EventTimestamp: timestamppb.New(timestamp),
			Object: serialize(getObject(objectParams{
				name:      "object1",
				namespace: "namespace1",
				gvk:       smfDeployGVK,
			})),
		},
		finalEvent: &preprocessor.Event{
			Type: preprocessor.List,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy1",
				Namespace:   "namespace1",
				Group:       "workload.nephio.org",
				Version:     "v1alpha1",
				Kind:        "SMFDeployment",
			},
			Timestamp: timestamp,
			Object: getObject(
				objectParams{
					name:      "object1",
					namespace: "namespace1",
					gvk:       smfDeployGVK,
				}),
		},
	},
	"smfdeploy add event": {
		key: preprocessor.RequestKey{
			ClusterName: "cluster1",
			NFDeploy:    "nfdeploy1",
			Namespace:   "namespace1",
			Group:       "workload.nephio.org",
			Version:     "v1alpha1",
			Kind:        "SMFDeployment",
		},
		event: &pb.EventRequest{
			Metadata: &pb.Metadata{
				Type: getPtr(pb.EventType_Added),
				Request: &pb.RequestMetadata{
					Namespace: getPtr("namespace1"),
					Kind:      getPtr(pb.CRDKind_SMFDeployment),
					Group:     getPtr(pb.APIGroup_NFDeployNephioOrg),
					Version:   getPtr(pb.Version_v1alpha1),
				},
				ClusterName:  getPtr("cluster1"),
				NfdeployName: getPtr("nfdeploy1"),
			},
			EventTimestamp: timestamppb.New(timestamp),
			Object: serialize(getObject(objectParams{
				name:      "object1",
				namespace: "namespace1",
				gvk:       smfDeployGVK,
			})),
		},
		finalEvent: &preprocessor.Event{
			Type: preprocessor.Added,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy1",
				Namespace:   "namespace1",
				Group:       "workload.nephio.org",
				Version:     "v1alpha1",
				Kind:        "SMFDeployment",
			},
			Timestamp: timestamp,
			Object: getObject(objectParams{
				name:      "object1",
				namespace: "namespace1",
				gvk:       smfDeployGVK,
			}),
		},
	},
	"smfdeploy modify event": {
		key: preprocessor.RequestKey{
			ClusterName: "cluster1",
			NFDeploy:    "nfdeploy1",
			Namespace:   "namespace1",
			Group:       "workload.nephio.org",
			Version:     "v1alpha1",
			Kind:        "SMFDeployment",
		},
		event: &pb.EventRequest{
			Metadata: &pb.Metadata{
				Type: getPtr(pb.EventType_Modified),
				Request: &pb.RequestMetadata{
					Namespace: getPtr("namespace1"),
					Kind:      getPtr(pb.CRDKind_SMFDeployment),
					Group:     getPtr(pb.APIGroup_NFDeployNephioOrg),
					Version:   getPtr(pb.Version_v1alpha1),
				},
				ClusterName:  getPtr("cluster1"),
				NfdeployName: getPtr("nfdeploy1"),
			},
			EventTimestamp: timestamppb.New(timestamp),
			Object: serialize(getObject(objectParams{
				name:      "object1",
				namespace: "namespace1",
				gvk:       smfDeployGVK,
			})),
		},
		finalEvent: &preprocessor.Event{
			Type: preprocessor.Modified,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy1",
				Namespace:   "namespace1",
				Group:       "workload.nephio.org",
				Version:     "v1alpha1",
				Kind:        "SMFDeployment",
			},
			Timestamp: timestamp,
			Object: getObject(objectParams{
				name:      "object1",
				namespace: "namespace1",
				gvk:       smfDeployGVK,
			}),
		},
	},
	"smfdeploy delete event": {
		key: preprocessor.RequestKey{
			ClusterName: "cluster1",
			NFDeploy:    "nfdeploy1",
			Namespace:   "namespace1",
			Group:       "workload.nephio.org",
			Version:     "v1alpha1",
			Kind:        "SMFDeployment",
		},
		event: &pb.EventRequest{
			Metadata: &pb.Metadata{
				Type: getPtr(pb.EventType_Deleted),
				Request: &pb.RequestMetadata{
					Namespace: getPtr("namespace1"),
					Kind:      getPtr(pb.CRDKind_SMFDeployment),
					Group:     getPtr(pb.APIGroup_NFDeployNephioOrg),
					Version:   getPtr(pb.Version_v1alpha1),
				},
				ClusterName:  getPtr("cluster1"),
				NfdeployName: getPtr("nfdeploy1"),
			},
			EventTimestamp: timestamppb.New(timestamp),
			Object: serialize(getObject(objectParams{
				name:      "object1",
				namespace: "namespace1",
				gvk:       smfDeployGVK,
			})),
		},
		finalEvent: &preprocessor.Event{
			Type: preprocessor.Deleted,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy1",
				Namespace:   "namespace1",
				Group:       "workload.nephio.org",
				Version:     "v1alpha1",
				Kind:        "SMFDeployment",
			},
			Timestamp: timestamp,
			Object: getObject(objectParams{
				name:      "object1",
				namespace: "namespace1",
				gvk:       smfDeployGVK,
			}),
		},
	},
	"amfdeploy list": {
		key: preprocessor.RequestKey{
			ClusterName: "cluster1",
			NFDeploy:    "nfdeploy1",
			Namespace:   "namespace1",
			Group:       "workload.nephio.org",
			Version:     "v1alpha1",
			Kind:        "AMFDeployment",
		},
		event: &pb.EventRequest{
			Metadata: &pb.Metadata{
				Type: getPtr(pb.EventType_List),
				Request: &pb.RequestMetadata{
					Namespace: getPtr("namespace1"),
					Kind:      getPtr(pb.CRDKind_AMFDeployment),
					Group:     getPtr(pb.APIGroup_NFDeployNephioOrg),
					Version:   getPtr(pb.Version_v1alpha1),
				},
				ClusterName:  getPtr("cluster1"),
				NfdeployName: getPtr("nfdeploy1"),
			},
			EventTimestamp: timestamppb.New(timestamp),
			Object: serialize(getObject(objectParams{
				name:      "object1",
				namespace: "namespace1",
				gvk:       amfDeployGVK,
			})),
		},
		finalEvent: &preprocessor.Event{
			Type: preprocessor.List,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy1",
				Namespace:   "namespace1",
				Group:       "workload.nephio.org",
				Version:     "v1alpha1",
				Kind:        "AMFDeployment",
			},
			Timestamp: timestamp,
			Object: getObject(
				objectParams{
					name:      "object1",
					namespace: "namespace1",
					gvk:       amfDeployGVK,
				}),
		},
	},
	"amfdeploy add event": {
		key: preprocessor.RequestKey{
			ClusterName: "cluster1",
			NFDeploy:    "nfdeploy1",
			Namespace:   "namespace1",
			Group:       "workload.nephio.org",
			Version:     "v1alpha1",
			Kind:        "AMFDeployment",
		},
		event: &pb.EventRequest{
			Metadata: &pb.Metadata{
				Type: getPtr(pb.EventType_Added),
				Request: &pb.RequestMetadata{
					Namespace: getPtr("namespace1"),
					Kind:      getPtr(pb.CRDKind_AMFDeployment),
					Group:     getPtr(pb.APIGroup_NFDeployNephioOrg),
					Version:   getPtr(pb.Version_v1alpha1),
				},
				ClusterName:  getPtr("cluster1"),
				NfdeployName: getPtr("nfdeploy1"),
			},
			EventTimestamp: timestamppb.New(timestamp),
			Object: serialize(getObject(objectParams{
				name:      "object1",
				namespace: "namespace1",
				gvk:       amfDeployGVK,
			})),
		},
		finalEvent: &preprocessor.Event{
			Type: preprocessor.Added,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy1",
				Namespace:   "namespace1",
				Group:       "workload.nephio.org",
				Version:     "v1alpha1",
				Kind:        "AMFDeployment",
			},
			Timestamp: timestamp,
			Object: getObject(objectParams{
				name:      "object1",
				namespace: "namespace1",
				gvk:       amfDeployGVK,
			}),
		},
	},
	"amfdeploy modify event": {
		key: preprocessor.RequestKey{
			ClusterName: "cluster1",
			NFDeploy:    "nfdeploy1",
			Namespace:   "namespace1",
			Group:       "workload.nephio.org",
			Version:     "v1alpha1",
			Kind:        "AMFDeployment",
		},
		event: &pb.EventRequest{
			Metadata: &pb.Metadata{
				Type: getPtr(pb.EventType_Modified),
				Request: &pb.RequestMetadata{
					Namespace: getPtr("namespace1"),
					Kind:      getPtr(pb.CRDKind_AMFDeployment),
					Group:     getPtr(pb.APIGroup_NFDeployNephioOrg),
					Version:   getPtr(pb.Version_v1alpha1),
				},
				ClusterName:  getPtr("cluster1"),
				NfdeployName: getPtr("nfdeploy1"),
			},
			EventTimestamp: timestamppb.New(timestamp),
			Object: serialize(getObject(objectParams{
				name:      "object1",
				namespace: "namespace1",
				gvk:       amfDeployGVK,
			})),
		},
		finalEvent: &preprocessor.Event{
			Type: preprocessor.Modified,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy1",
				Namespace:   "namespace1",
				Group:       "workload.nephio.org",
				Version:     "v1alpha1",
				Kind:        "AMFDeployment",
			},
			Timestamp: timestamp,
			Object: getObject(objectParams{
				name:      "object1",
				namespace: "namespace1",
				gvk:       amfDeployGVK,
			}),
		},
	},
	"amfdeploy delete event": {
		key: preprocessor.RequestKey{
			ClusterName: "cluster1",
			NFDeploy:    "nfdeploy1",
			Namespace:   "namespace1",
			Group:       "workload.nephio.org",
			Version:     "v1alpha1",
			Kind:        "AMFDeployment",
		},
		event: &pb.EventRequest{
			Metadata: &pb.Metadata{
				Type: getPtr(pb.EventType_Deleted),
				Request: &pb.RequestMetadata{
					Namespace: getPtr("namespace1"),
					Kind:      getPtr(pb.CRDKind_AMFDeployment),
					Group:     getPtr(pb.APIGroup_NFDeployNephioOrg),
					Version:   getPtr(pb.Version_v1alpha1),
				},
				ClusterName:  getPtr("cluster1"),
				NfdeployName: getPtr("nfdeploy1"),
			},
			EventTimestamp: timestamppb.New(timestamp),
			Object: serialize(getObject(objectParams{
				name:      "object1",
				namespace: "namespace1",
				gvk:       amfDeployGVK,
			})),
		},
		finalEvent: &preprocessor.Event{
			Type: preprocessor.Deleted,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy1",
				Namespace:   "namespace1",
				Group:       "workload.nephio.org",
				Version:     "v1alpha1",
				Kind:        "AMFDeployment",
			},
			Timestamp: timestamp,
			Object: getObject(objectParams{
				name:      "object1",
				namespace: "namespace1",
				gvk:       amfDeployGVK,
			}),
		},
	},
}

func serialize(obj runtime.Object) []byte {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		panic(err)
	}
	u := &unstructured.Unstructured{
		Object: data,
	}
	b, err := u.MarshalJSON()
	if err != nil {
		panic(err)
	}
	return b
}

func getObject(param objectParams) runtime.Object {
	u := &unstructured.Unstructured{}
	u.SetNamespace(param.namespace)
	u.SetName(param.name)
	u.SetGroupVersionKind(param.gvk)
	return u
}
