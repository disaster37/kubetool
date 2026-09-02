package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/disaster37/kubetool/v1.32/kubetool"
	longhorn "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/kubectl/pkg/scheme"
)

// toUnstructured converts a typed object to an unstructured one, as the
// dynamic client would return from a real cluster.
func toUnstructured(t testing.TB, obj runtime.Object) runtime.Object {
	t.Helper()
	m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	require.NoError(t, err)
	return &unstructured.Unstructured{Object: m}
}

// listBackupNames returns the names of the backups currently known by the client.
func listBackupNames(t testing.TB, client dynamic.Interface) []string {
	t.Helper()
	list, err := client.Resource(longhorn.SchemeGroupVersion.WithResource("backups")).Namespace("").List(context.Background(), meta.ListOptions{})
	require.NoError(t, err)
	names := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		names = append(names, item.GetName())
	}
	return names
}

func backupObject(name, volumeName, createdAt string) *longhorn.Backup {
	return &longhorn.Backup{
		TypeMeta: meta.TypeMeta{
			APIVersion: longhorn.SchemeGroupVersion.String(),
			Kind:       "Backup",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: name,
		},
		Status: longhorn.BackupStatus{
			VolumeName:      volumeName,
			BackupCreatedAt: createdAt,
		},
	}
}

func (s *TestSuite) TestcleanLonghornPendingBackup() {
	sh := scheme.Scheme
	require.NoError(s.T(), longhorn.AddToScheme(sh))

	dynamicFakeClient := dynamicFake.NewSimpleDynamicClient(sh,
		toUnstructured(s.T(), &longhorn.Backup{
			TypeMeta:   meta.TypeMeta{APIVersion: longhorn.SchemeGroupVersion.String(), Kind: "Backup"},
			ObjectMeta: meta.ObjectMeta{Name: "pending-backup"},
			Status:     longhorn.BackupStatus{State: "Pending"},
		}),
		toUnstructured(s.T(), &longhorn.Backup{
			TypeMeta:   meta.TypeMeta{APIVersion: longhorn.SchemeGroupVersion.String(), Kind: "Backup"},
			ObjectMeta: meta.ObjectMeta{Name: "completed-backup"},
			Status:     longhorn.BackupStatus{State: "Completed"},
		}),
	)

	cmd := kubetool.NewConnexionFromClient(nil, dynamicFakeClient)

	err := cleanLonghornPendingBackup(context.Background(), cmd)
	assert.NoError(s.T(), err)

	// Only the pending backup must be deleted
	assert.ElementsMatch(s.T(), []string{"completed-backup"}, listBackupNames(s.T(), dynamicFakeClient))
}

func (s *TestSuite) TestcleanLonghornOrphanBackup() {
	sh := scheme.Scheme
	require.NoError(s.T(), longhorn.AddToScheme(sh))

	now := time.Now().Format(time.RFC3339)
	old := time.Now().Add(-100 * 24 * time.Hour).Format(time.RFC3339)

	objects := []runtime.Object{
		toUnstructured(s.T(), backupObject("backup-volume", "volume1", now)),
		toUnstructured(s.T(), backupObject("old-backup-not-older-than", "older", now)),
		toUnstructured(s.T(), backupObject("old-backup", "older", old)),
	}
	for _, name := range []string{"volume1", "volume2", "volume3"} {
		objects = append(objects, toUnstructured(s.T(), &longhorn.Volume{
			TypeMeta:   meta.TypeMeta{APIVersion: longhorn.SchemeGroupVersion.String(), Kind: "Volume"},
			ObjectMeta: meta.ObjectMeta{Name: name},
		}))
	}

	dynamicFakeClient := dynamicFake.NewSimpleDynamicClient(sh, objects...)
	cmd := kubetool.NewConnexionFromClient(nil, dynamicFakeClient)

	err := cleanLonghornOldOrphanBackup(context.Background(), cmd, 24*time.Hour)
	assert.NoError(s.T(), err)

	// Only the backup whose volume does not exist anymore and which is old
	// enough must be deleted
	assert.ElementsMatch(s.T(), []string{"backup-volume", "old-backup-not-older-than"}, listBackupNames(s.T(), dynamicFakeClient))
}
