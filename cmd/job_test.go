package cmd

import (
	"context"
	"fmt"

	"github.com/disaster37/kubetool/v1.18/kubetool"
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

func (s *TestSuite) TestRunPreJob() {

	fakeClient := fake.NewSimpleClientset()
	fakeClient.Fake = fake.Clientset{}.Fake
	fakeDiscovery := fakeClient.Discovery().(*discoveryfake.FakeDiscovery)
	fakeDiscovery.FakedServerVersion = &version.Info{
		Major: "1",
		Minor: "18",
	}

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

	err := runPreJob(context.Background(), cmd, "fake-namespace")
	assert.NoError(s.T(), err)
}

func (s *TestSuite) TestRunPostJob() {

	fakeClient := fake.NewSimpleClientset()
	fakeClient.Fake = fake.Clientset{}.Fake
	fakeDiscovery := fakeClient.Discovery().(*discoveryfake.FakeDiscovery)
	fakeDiscovery.FakedServerVersion = &version.Info{
		Major: "1",
		Minor: "18",
	}

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

	err := runPostJob(context.Background(), cmd, "fake-namespace")
	assert.NoError(s.T(), err)
}
