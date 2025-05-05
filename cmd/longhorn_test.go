package cmd

import (
	"context"
	"fmt"

	"github.com/disaster37/kubetool/v1.32/kubetool"
	longhorn "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	"github.com/stretchr/testify/assert"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	discoveryfake "k8s.io/client-go/discovery/fake"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/kubectl/pkg/scheme"
)

func (s *TestSuite) TestcleanLonghornPendingBackup() {

	fakeClient := fake.NewSimpleClientset()
	fakeClient.Fake = k8stesting.Fake{}
	fakeDiscovery := fakeClient.Discovery().(*discoveryfake.FakeDiscovery)
	fakeDiscovery.FakedServerVersion = FaikedVersion

	sh := scheme.Scheme
	if err := longhorn.AddToScheme(sh); err != nil {
		panic(err)
	}
	dynamicFakeClient := dynamicFake.NewSimpleDynamicClient(sh)

	// Mock list backups
	fakeClient.AddReactor("list", "backups", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		backups := &longhorn.BackupList{
			Items: []longhorn.Backup{
				{
					ObjectMeta: meta.ObjectMeta{
						Name: "pending-backup",
					},
					Status: longhorn.BackupStatus{
						State: "Pending",
					},
				},
				{
					ObjectMeta: meta.ObjectMeta{
						Name: "completed-backup",
					},
					Status: longhorn.BackupStatus{
						State: "Completed",
					},
				},
			},
		}
		return true, backups, nil
	})

	// Mock delete backup
	fakeClient.AddReactor("delete", "backups", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, nil
	})

	// Trap all
	fakeClient.AddReactor("*", "*", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, fmt.Errorf("no reaction implemented for %s", action)
	})
	cmd := kubetool.NewConnexionFromClient(fakeClient, dynamicFakeClient)

	err := cleanLonghornPendingBackup(context.Background(), cmd)
	assert.NoError(s.T(), err)
}
