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
	"time"

	"k8s.io/apimachinery/pkg/runtime"
)

type EventType string

const (
	List     EventType = "LIST"
	Added    EventType = "ADDED"
	Modified EventType = "MODIFIED"
	Deleted  EventType = "DELETED"
)

type listKey struct {
	clusterName string
	group       string
	version     string
	kind        string
	namespace   string
}

type RequestKey struct {
	ClusterName string
	Group       string
	Version     string
	Kind        string
	Namespace   string
	NFDeploy    string
}

type EventMessage struct {
	Ctx context.Context
	Err chan error

	Event
}

type Event struct {
	Type      EventType
	Timestamp time.Time
	Key       RequestKey
	Object    runtime.Object
}
