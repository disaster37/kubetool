package cmd

import (
	"context"
	"fmt"

	"github.com/disaster37/kubetool/v1.32/kubetool"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	discoveryfake "k8s.io/client-go/discovery/fake"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/kubectl/pkg/scheme"
)

func (s *TestSuite) TestCleanEvictedNodes() {
	fakeClient := fake.NewSimpleClientset()
	fakeClient.Fake = k8stesting.Fake{}
	fakeDiscovery := fakeClient.Discovery().(*discoveryfake.FakeDiscovery)
	fakeDiscovery.FakedServerVersion = FaikedVersion

	sh := scheme.Scheme
	dynamicFakeClient := dynamicFake.NewSimpleDynamicClient(sh)

	// Mock list pod
	fakeClient.AddReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		pods := &v1.PodList{
			Items: []v1.Pod{
				{
					Spec: v1.PodSpec{
						NodeName: "fake-node",
					},
					ObjectMeta: meta.ObjectMeta{
						Name:      "fake-pod",
						Namespace: "fake-namespace",
					},
				},
			},
		}
		return true, pods, nil
	})

	// Mock delete pod
	fakeClient.AddReactor("delete", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, nil
	})

	// Trap all
	fakeClient.AddReactor("*", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("no reaction implemented for %s", action)
	})
	cmd := kubetool.NewConnexionFromClient(fakeClient, dynamicFakeClient)

	err := cleanEvictedPods(context.Background(), cmd)
	assert.NoError(s.T(), err)
}
