// /*
// Copyright The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

// Code generated by client-gen. DO NOT EDIT.
package managedclusterclient

import (
	"context"
	"strings"

	armcontainerservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var beforeAllFunc func(context.Context)
var afterAllFunc func(context.Context)
var additionalTestCases func()
var newResource *armcontainerservice.ManagedCluster = &armcontainerservice.ManagedCluster{}

var _ = Describe("ManagedClustersClient", Ordered, func() {

	if beforeAllFunc != nil {
		BeforeAll(beforeAllFunc)
	}

	if additionalTestCases != nil {
		additionalTestCases()
	}

	When("creation requests are raised", func() {
		It("should not return error", func(ctx context.Context) {
			newResource, err := realClient.CreateOrUpdate(ctx, resourceGroupName, resourceName, *newResource)
			Expect(err).NotTo(HaveOccurred())
			Expect(newResource).NotTo(BeNil())
			Expect(strings.EqualFold(*newResource.Name, resourceName)).To(BeTrue())
		})
	})

	When("get requests are raised", func() {
		It("should not return error", func(ctx context.Context) {
			newResource, err := realClient.Get(ctx, resourceGroupName, resourceName)
			Expect(err).NotTo(HaveOccurred())
			Expect(newResource).NotTo(BeNil())
		})
	})
	When("invalid get requests are raised", func() {
		It("should return 404 error", func(ctx context.Context) {
			newResource, err := realClient.Get(ctx, resourceGroupName, resourceName+"notfound")
			Expect(err).To(HaveOccurred())
			Expect(newResource).To(BeNil())
		})
	})

	When("update requests are raised", func() {
		It("should not return error", func(ctx context.Context) {
			newResource, err := realClient.CreateOrUpdate(ctx, resourceGroupName, resourceName, *newResource)
			Expect(err).NotTo(HaveOccurred())
			Expect(newResource).NotTo(BeNil())
		})
	})

	When("list requests are raised", func() {
		It("should not return error", func(ctx context.Context) {
			resourceList, err := realClient.List(ctx, resourceGroupName)
			Expect(err).NotTo(HaveOccurred())
			Expect(resourceList).NotTo(BeNil())
			Expect(len(resourceList)).To(Equal(1))
			Expect(*resourceList[0].Name).To(Equal(resourceName))
		})
	})
	When("invalid list requests are raised", func() {
		It("should return error", func(ctx context.Context) {
			resourceList, err := realClient.List(ctx, resourceGroupName+"notfound")
			Expect(err).To(HaveOccurred())
			Expect(resourceList).To(BeNil())
		})
	})

	When("deletion requests are raised", func() {
		It("should not return error", func(ctx context.Context) {
			err = realClient.Delete(ctx, resourceGroupName, resourceName)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	if afterAllFunc != nil {
		AfterAll(afterAllFunc)
	}
})
