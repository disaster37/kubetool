package cmd

import (
	"context"
	"fmt"

	"github.com/disaster37/kubetool/v1.23/kubetool"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	discoveryfake "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func (s *TestSuite) TestCleanEvictedNodes() {

	fakeClient := fake.NewSimpleClientset()
	fakeClient.Fake = k8stesting.Fake{}
	fakeDiscovery := fakeClient.Discovery().(*discoveryfake.FakeDiscovery)
	fakeDiscovery.FakedServerVersion = FaikedVersion

	// Mock list pod
	fakeClient.Fake.AddReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
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
	fakeClient.Fake.AddReactor("delete", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, nil
	})

	// Trap all
	fakeClient.Fake.AddReactor("*", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("no reaction implemented for %s", action)
	})
	cmd := kubetool.NewConnexionFromClient(fakeClient)

	err := cleanEvictedPods(context.Background(), cmd)
	assert.NoError(s.T(), err)
}
