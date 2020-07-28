package cmd

import (
	"context"

	"github.com/disaster37/kubetools/v1.18/kubetool"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func (s *TestSuite) TestGetWorkerNodes() {

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

	cmd := kubetool.NewConnexionFromClient(fakeClient)

	nodes, err := getWorkerNodes(context.TODO(), cmd)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "worker1", nodes[0])
}

func (s *TestSuite) TestGetMasterNodes() {

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

	cmd := kubetool.NewConnexionFromClient(fakeClient)

	nodes, err := getMasterNodes(context.TODO(), cmd)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "master1", nodes[0])

}
