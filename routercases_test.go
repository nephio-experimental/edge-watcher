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

package edgewatcher_test

import (
	"github.com/nephio-project/edge-watcher/preprocessor"
)

var events = map[string][]preprocessor.Event{
	"cluster1": {
		{
			Type: preprocessor.List,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy12",
			},
		},
		{
			Type: preprocessor.List,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy13",
			},
		},
		{
			Type: preprocessor.List,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy14",
			},
		},
		{
			Type: preprocessor.List,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy15",
			},
		},
	},
	"nfdeploy1": {
		{

			Type: preprocessor.List,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster12",
				NFDeploy:    "nfdeploy1",
			},
		},
		{
			Type: preprocessor.List,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster13",
				NFDeploy:    "nfdeploy1",
			},
		},
		{
			Type: preprocessor.List,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster14",
				NFDeploy:    "nfdeploy1",
			},
		},
	},
	"common": {
		{
			Type: preprocessor.List,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy1",
			},
		},
		{

			Type: preprocessor.List,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy1",
			},
		},
		{

			Type: preprocessor.List,
			Key: preprocessor.RequestKey{
				ClusterName: "cluster1",
				NFDeploy:    "nfdeploy1",
			},
		},
	},
}
