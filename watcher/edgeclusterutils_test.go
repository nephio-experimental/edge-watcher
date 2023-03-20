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

package watcher_test

import (
	"context"
	"strconv"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func createEdgeClusters(ctx context.Context, n int) {
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 1; i <= n; i++ {
		go func(i int) {
			defer GinkgoRecover()
			defer wg.Done()

			u := &unstructured.Unstructured{}
			u.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "cloud.nephio.org",
				Version: "v1alpha1",
				Kind:    "EdgeCluster",
			})
			u.SetName("edgecluster" + strconv.Itoa(i))
			u.SetNamespace(testEnv.GetNamespace())
			_, err := testEnv.DynamicClient.Resource(schema.GroupVersionResource{
				Group:    "cloud.nephio.org",
				Version:  "v1alpha1",
				Resource: "edgeclusters",
			}).Namespace(u.GetNamespace()).Create(ctx, u, metav1.CreateOptions{})
			Expect(err).To(BeNil())
		}(i)
	}
	wg.Wait()
}
