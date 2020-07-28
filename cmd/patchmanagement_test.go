package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/disaster37/kubetools/v1.18/kubetool"
	"github.com/stretchr/testify/assert"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	discoveryfake "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

// When node is not ready
// It must return error
func (s *TestSuite) TestSetDowntimeWhenNodeNotReady() {

	fakeClient := &fake.Clientset{}
	// Mock get node
	fakeClient.Fake.AddReactor("get", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{},
		}

		return true, node, nil
	})

	// Trap all
	fakeClient.Fake.AddReactor("*", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("no reaction implemented for %s", action)
	})
	cmd := kubetool.NewConnexionFromClient(fakeClient)

	err := setDowntime(context.TODO(), cmd, "fake-node")
	assert.Error(s.T(), err, "Node fake-node is not on ready state")
}

// When cordon node failed
// It must return error
func (s *TestSuite) TestSetDowntimeWhenCordonFailed() {

	fakeClient := &fake.Clientset{}

	// Mock get node
	fakeClient.Fake.AddReactor("get", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}

		return true, node, nil
	})

	// Mock cordon node
	fakeClient.Fake.AddReactor("patch", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {

		return true, nil, fmt.Errorf("Cordon failed")
	})

	// Trap all
	fakeClient.Fake.AddReactor("*", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("no reaction implemented for %s", action)
	})
	cmd := kubetool.NewConnexionFromClient(fakeClient)

	err := setDowntime(context.TODO(), cmd, "fake-node")
	assert.Error(s.T(), err, "Cordon failed")

}

// When node ready, no pod on node, and all success
// It must return no error
func (s *TestSuite) TestSetDowntimeWhenNoPodsAndDrainSuccess() {

	fakeClient := &fake.Clientset{}

	// Mock get node
	fakeClient.Fake.AddReactor("get", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}

		return true, node, nil
	})

	// Mock cordon node
	fakeClient.Fake.AddReactor("patch", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
			Spec: v1.NodeSpec{Unschedulable: true},
		}

		return true, node, nil
	})

	// Mock list pod on node
	fakeClient.Fake.AddReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		pods := &v1.PodList{
			Items: []v1.Pod{},
		}
		return true, pods, nil
	})

	// Trap all
	fakeClient.Fake.AddReactor("*", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("no reaction implemented for %s", action)
	})
	cmd := kubetool.NewConnexionFromClient(fakeClient)

	err := setDowntime(context.TODO(), cmd, "fake-node")
	assert.NoError(s.T(), err)

}

// When node ready, some pods on node, ans all successs
// It must return no error
func (s *TestSuite) TestSetDowntimeWhenPodsAndDrainSuccess() {

	fakeClient := fake.NewSimpleClientset()
	fakeClient.Fake = fake.Clientset{}.Fake
	fakeDiscovery := fakeClient.Discovery().(*discoveryfake.FakeDiscovery)
	fakeDiscovery.FakedServerVersion = &version.Info{
		Major: "1",
		Minor: "18",
	}

	// Mock get node
	fakeClient.Fake.AddReactor("get", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}

		return true, node, nil
	})

	// Mock patch node
	fakeClient.Fake.AddReactor("patch", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
			Spec: v1.NodeSpec{Unschedulable: true},
		}

		return true, node, nil
	})

	// Mock list pods
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
				{
					Spec: v1.PodSpec{
						NodeName: "fake-node2",
					},
					ObjectMeta: meta.ObjectMeta{
						Name:      "fake-pod2",
						Namespace: "fake-namespace2",
					},
				},
			},
		}
		return true, pods, nil
	})

	// Mock get pod
	fakeClient.Fake.AddReactor("get", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {

		return true, nil, errors.NewNotFound(v1.Resource("pods"), "pod")
	})

	// Mock get configmap
	fakeClient.Fake.AddReactor("get", "configmaps", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.NewNotFound(v1.Resource("configmaps"), "patchmanagement")
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

	err := setDowntime(context.TODO(), cmd, "fake-node")
	assert.NoError(s.T(), err)

}

// When node ready, some pods on node and drain failed
// It must return error
func (s *TestSuite) TestSetDowntimeWhenPodsAndDrainFailed() {

	fakeClient := fake.NewSimpleClientset()
	fakeClient.Fake = fake.Clientset{}.Fake
	fakeDiscovery := fakeClient.Discovery().(*discoveryfake.FakeDiscovery)
	fakeDiscovery.FakedServerVersion = &version.Info{
		Major: "1",
		Minor: "18",
	}

	// Mock get node
	fakeClient.Fake.AddReactor("get", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}

		return true, node, nil
	})

	// Mock cordon node
	fakeClient.Fake.AddReactor("patch", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
			Spec: v1.NodeSpec{Unschedulable: true},
		}

		return true, node, nil
	})

	// Mock list pods
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
				{
					Spec: v1.PodSpec{
						NodeName: "fake-node2",
					},
					ObjectMeta: meta.ObjectMeta{
						Name:      "fake-pod2",
						Namespace: "fake-namespace2",
					},
				},
			},
		}
		return true, pods, nil
	})

	// Mock get pod
	fakeClient.Fake.AddReactor("get", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {

		return true, nil, errors.NewNotFound(v1.Resource("pods"), "pod")

	})

	// Mock get configmap
	fakeClient.Fake.AddReactor("get", "configmaps", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.NewNotFound(v1.Resource("configmaps"), "patchmanagement")
	})

	// Mock delete pod
	fakeClient.Fake.AddReactor("delete", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("Failed to delete pod")
	})

	// Trap all
	fakeClient.Fake.AddReactor("*", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("no reaction implemented for %s", action)
	})
	cmd := kubetool.NewConnexionFromClient(fakeClient)

	err := setDowntime(context.TODO(), cmd, "fake-node")
	assert.Error(s.T(), err, "Failed to delete pod")

}

// When node reay, some pods on node, pre job found and all success
// It must return no error
func (s *TestSuite) TestSetDowntimeWhenPodsAndPrejobWitSecretAndDrainSuccess() {

	fakeClient := fake.NewSimpleClientset()
	fakeClient.Fake = fake.Clientset{}.Fake
	fakeDiscovery := fakeClient.Discovery().(*discoveryfake.FakeDiscovery)
	fakeDiscovery.FakedServerVersion = &version.Info{
		Major: "1",
		Minor: "18",
	}

	// Mock get node
	fakeClient.Fake.AddReactor("get", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}

		return true, node, nil
	})

	// Mock cordon node
	fakeClient.Fake.AddReactor("patch", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
			Spec: v1.NodeSpec{Unschedulable: true},
		}

		return true, node, nil
	})

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

	// Mock get pod
	fakeClient.Fake.AddReactor("get", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {

		return true, nil, errors.NewNotFound(v1.Resource("pods"), "pod")
	})

	// Mock get configmap
	fakeClient.Fake.AddReactor("get", "configmaps", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {

		configmap := &v1.ConfigMap{
			ObjectMeta: meta.ObjectMeta{
				Name:      "patchmanagement",
				Namespace: "fake-namespace",
			},
			Data: map[string]string{
				"pre-job": "fake pre-job",
				"secrets": "fake-secret",
			},
		}

		return true, configmap, nil
	})

	// Mock delete pod
	fakeClient.Fake.AddReactor("delete", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, nil
	})

	// Mock get jobs
	// First time for old job, and next that job is finished
	countCallJob := 0
	fakeClient.Fake.AddReactor("get", "jobs", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		if countCallJob == 0 {
			// Call when look if jon already exist
			countCallJob++
			return true, nil, errors.NewNotFound(v1.Resource("jobs"), "job")
		} else {
			// Cal after create job
			job := &batch.Job{
				ObjectMeta: meta.ObjectMeta{
					Name:      "patchmanagement-pre-job",
					Namespace: "fake-namespace",
				},
				Status: batch.JobStatus{
					Conditions: []batch.JobCondition{
						{
							Type:   batch.JobComplete,
							Status: v1.ConditionTrue,
						},
					},
				},
			}

			return true, job, nil
		}

	})

	// Mock create job
	fakeClient.Fake.AddReactor("create", "jobs", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		createAction := action.(k8stesting.CreateAction)

		return true, createAction.GetObject(), nil
	})

	// Trap all
	fakeClient.Fake.AddReactor("*", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("no reaction implemented for %s", action)
	})
	cmd := kubetool.NewConnexionFromClient(fakeClient)

	err := setDowntime(context.TODO(), cmd, "fake-node")
	assert.NoError(s.T(), err)
}

// When uncordon node failed
// It must return error
func (s *TestSuite) TestUnsetDowntimeWhenUncordonFailed() {

	fakeClient := &fake.Clientset{}

	// Mock get node
	fakeClient.Fake.AddReactor("get", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}

		return true, node, nil
	})

	// Mock cordon node
	fakeClient.Fake.AddReactor("patch", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {

		return true, nil, fmt.Errorf("Uncordon failed")
	})

	// Trap all
	fakeClient.Fake.AddReactor("*", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("no reaction implemented for %s", action)
	})
	cmd := kubetool.NewConnexionFromClient(fakeClient)

	err := unsetDowntime(context.TODO(), cmd, "fake-node")
	assert.Error(s.T(), err, "Uncordon failed")
}

// When node ready, no pod on node, and all success
// It must return no error
func (s *TestSuite) TestUnsetDowntimeWhenNoPodsAndUncordonSuccess() {

	fakeClient := &fake.Clientset{}

	// Mock get node
	fakeClient.Fake.AddReactor("get", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}

		return true, node, nil
	})

	// Mock uncordon node
	fakeClient.Fake.AddReactor("patch", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
			Spec: v1.NodeSpec{Unschedulable: true},
		}

		return true, node, nil
	})

	// Mock list pod on node
	fakeClient.Fake.AddReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		pods := &v1.PodList{
			Items: []v1.Pod{},
		}
		return true, pods, nil
	})

	// Trap all
	fakeClient.Fake.AddReactor("*", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("no reaction implemented for %s", action)
	})
	cmd := kubetool.NewConnexionFromClient(fakeClient)

	err := unsetDowntime(context.TODO(), cmd, "fake-node")
	assert.NoError(s.T(), err)

}

// When node ready, some pods on node, and all successs
// It must return no error
func (s *TestSuite) TestUnsetDowntimeWhenPodsAndUncordonSuccess() {

	fakeClient := fake.NewSimpleClientset()
	fakeClient.Fake = fake.Clientset{}.Fake
	fakeDiscovery := fakeClient.Discovery().(*discoveryfake.FakeDiscovery)
	fakeDiscovery.FakedServerVersion = &version.Info{
		Major: "1",
		Minor: "18",
	}

	// Mock get node
	fakeClient.Fake.AddReactor("get", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}

		return true, node, nil
	})

	// Mock patch node
	fakeClient.Fake.AddReactor("patch", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
			Spec: v1.NodeSpec{Unschedulable: true},
		}

		return true, node, nil
	})

	// Mock list pods
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
				{
					Spec: v1.PodSpec{
						NodeName: "fake-node2",
					},
					ObjectMeta: meta.ObjectMeta{
						Name:      "fake-pod2",
						Namespace: "fake-namespace2",
					},
				},
			},
		}
		return true, pods, nil
	})

	// Mock get pod
	fakeClient.Fake.AddReactor("get", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {

		return true, nil, errors.NewNotFound(v1.Resource("pods"), "pod")
	})

	// Mock get configmap
	fakeClient.Fake.AddReactor("get", "configmaps", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.NewNotFound(v1.Resource("configmaps"), "patchmanagement")
	})

	// Trap all
	fakeClient.Fake.AddReactor("*", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("no reaction implemented for %s", action)
	})
	cmd := kubetool.NewConnexionFromClient(fakeClient)

	err := unsetDowntime(context.TODO(), cmd, "fake-node")
	assert.NoError(s.T(), err)

}

// When node ready, some pods on node, post job found and all success
// It must return no error
func (s *TestSuite) TestUnsetDowntimeWhenPodsAndPostjobWitSecretAndUncordonSuccess() {

	fakeClient := fake.NewSimpleClientset()
	fakeClient.Fake = fake.Clientset{}.Fake
	fakeDiscovery := fakeClient.Discovery().(*discoveryfake.FakeDiscovery)
	fakeDiscovery.FakedServerVersion = &version.Info{
		Major: "1",
		Minor: "18",
	}

	// Mock get node
	fakeClient.Fake.AddReactor("get", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}

		return true, node, nil
	})

	// Mock uncordon node
	fakeClient.Fake.AddReactor("patch", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
			Spec: v1.NodeSpec{Unschedulable: false},
		}

		return true, node, nil
	})

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

	// Mock get pod
	fakeClient.Fake.AddReactor("get", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {

		return true, nil, errors.NewNotFound(v1.Resource("pods"), "pod")
	})

	// Mock get configmap
	fakeClient.Fake.AddReactor("get", "configmaps", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {

		configmap := &v1.ConfigMap{
			ObjectMeta: meta.ObjectMeta{
				Name:      "patchmanagement",
				Namespace: "fake-namespace",
			},
			Data: map[string]string{
				"post-job": "fake post-job",
				"secrets":  "fake-secret",
			},
		}

		return true, configmap, nil
	})

	// Mock get jobs
	// First time for old job, and next that job is finished
	countCallJob := 0
	fakeClient.Fake.AddReactor("get", "jobs", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		if countCallJob == 0 {
			// Call when look if jon already exist
			countCallJob++
			return true, nil, errors.NewNotFound(v1.Resource("jobs"), "job")
		} else {
			// Cal after create job
			job := &batch.Job{
				ObjectMeta: meta.ObjectMeta{
					Name:      "patchmanagement-post-job",
					Namespace: "fake-namespace",
				},
				Status: batch.JobStatus{
					Conditions: []batch.JobCondition{
						{
							Type:   batch.JobComplete,
							Status: v1.ConditionTrue,
						},
					},
				},
			}

			return true, job, nil
		}

	})

	// Mock create job
	fakeClient.Fake.AddReactor("create", "jobs", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		createAction := action.(k8stesting.CreateAction)

		return true, createAction.GetObject(), nil
	})

	// Trap all
	fakeClient.Fake.AddReactor("*", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("no reaction implemented for %s", action)
	})
	cmd := kubetool.NewConnexionFromClient(fakeClient)

	err := unsetDowntime(context.TODO(), cmd, "fake-node")
	assert.NoError(s.T(), err)
}

func (s *TestSuite) TestUncordonNodeForRecueWhenSuccess() {

	fakeClient := fake.NewSimpleClientset()
	fakeClient.Fake = fake.Clientset{}.Fake
	fakeDiscovery := fakeClient.Discovery().(*discoveryfake.FakeDiscovery)
	fakeDiscovery.FakedServerVersion = &version.Info{
		Major: "1",
		Minor: "18",
	}

	// Mock get node
	fakeClient.Fake.AddReactor("get", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
			Spec: v1.NodeSpec{Unschedulable: true},
		}

		return true, node, nil
	})

	// Mock uncordon node
	fakeClient.Fake.AddReactor("patch", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
			Spec: v1.NodeSpec{Unschedulable: false},
		}

		return true, node, nil
	})

	err := uncordonNodeForRecue(kubetool.NewConnexionFromClient(fakeClient), "fake-node")
	assert.NoError(s.T(), err)

}

func (s *TestSuite) TestUncordonNodeForRecueWhenFailed() {

	fakeClient := fake.NewSimpleClientset()
	fakeClient.Fake = fake.Clientset{}.Fake
	fakeDiscovery := fakeClient.Discovery().(*discoveryfake.FakeDiscovery)
	fakeDiscovery.FakedServerVersion = &version.Info{
		Major: "1",
		Minor: "18",
	}

	// Mock get node
	fakeClient.Fake.AddReactor("get", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		node := &v1.Node{
			ObjectMeta: meta.ObjectMeta{
				Name:              "fake-node",
				CreationTimestamp: meta.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
			Spec: v1.NodeSpec{Unschedulable: true},
		}

		return true, node, nil
	})

	// Mock uncordon node
	fakeClient.Fake.AddReactor("patch", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {

		return true, nil, errors.NewInternalError(fmt.Errorf("Uncordon failed"))
	})

	err := uncordonNodeForRecue(kubetool.NewConnexionFromClient(fakeClient), "fake-node")
	assert.Error(s.T(), err, "Uncordon failed")

}
