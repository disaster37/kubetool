package cmd

import (
	"context"

	"github.com/disaster37/kubetool/v1.32/kubetool"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubectl/pkg/scheme"
)

func (s *TestSuite) TestGetWorkerNodes() {
	sh := scheme.Scheme
	dynamicFakeClient := dynamicFake.NewSimpleDynamicClient(sh)

	fakeClient := fake.NewSimpleClientset(
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "master1",
				Labels: map[string]string{
					"master": "true",
				},
			},
		},
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "worker1",
				Labels: map[string]string{
					"worker": "true",
				},
			},
		},
	)

	cmd := kubetool.NewConnexionFromClient(fakeClient, dynamicFakeClient)

	nodes, err := getWorkerNodes(context.TODO(), cmd)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "worker1", nodes[0])
}

func (s *TestSuite) TestGetMasterNodes() {
	sh := scheme.Scheme
	dynamicFakeClient := dynamicFake.NewSimpleDynamicClient(sh)

	fakeClient := fake.NewSimpleClientset(
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "master1",
				Labels: map[string]string{
					"master": "true",
				},
			},
		},
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "worker1",
				Labels: map[string]string{
					"worker": "true",
				},
			},
		},
	)

	cmd := kubetool.NewConnexionFromClient(fakeClient, dynamicFakeClient)

	nodes, err := getMasterNodes(context.TODO(), cmd)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "master1", nodes[0])
}
