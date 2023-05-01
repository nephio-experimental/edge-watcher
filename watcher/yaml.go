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

package watcher

import (
	"time"

	"github.com/nephio-project/watcher-agent/api/v1alpha1"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	watcherAgentKind       = "WatcherAgent"
	watcherAgentAPIVersion = "monitor.nephio.org/v1alpha1"
	watcherAgentName       = "watcheragent-upf-smf"
	watcherAgentAnnotation = "cloud.nephio.org/cluster"
	upfDeployGroup         = "workload.nephio.org"
	upfDeployKind          = "UPFDeployment"
	upfDeployVersion       = "v1alpha1"
	smfDeployGroup         = "workload.nephio.org"
	smfDeployKind          = "SMFDeployment"
	smfDeployVersion       = "v1alpha1"
)

type watcherAgent struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	Spec              v1alpha1.WatcherAgentSpec `json:"spec" yaml:"spec"`
}

func GetYAML(IP, port, clusterName, nephioNamespace string, timestamp time.Time) (string, error) {
	obj := watcherAgent{
		ResourceMeta: yaml.ResourceMeta{
			TypeMeta: yaml.TypeMeta{
				Kind:       watcherAgentKind,
				APIVersion: watcherAgentAPIVersion,
			},
			ObjectMeta: yaml.ObjectMeta{
				NameMeta: yaml.NameMeta{
					Name:      watcherAgentName,
					Namespace: nephioNamespace,
				},
				Annotations: map[string]string{
					watcherAgentAnnotation: clusterName,
				},
			},
		},
		Spec: v1alpha1.WatcherAgentSpec{
			ResyncTimestamp: metav1.NewTime(timestamp).Rfc3339Copy(),
			EdgeWatcher: v1alpha1.EdgeWatcher{
				Addr: IP,
				Port: port,
			},
			WatchRequests: []v1alpha1.WatchRequest{
				{
					Group:     upfDeployGroup,
					Kind:      upfDeployKind,
					Namespace: nephioNamespace,
					Version:   upfDeployVersion,
				},
			},
		},
	}
	b, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(b), err
}
