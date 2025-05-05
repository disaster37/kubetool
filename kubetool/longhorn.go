package kubetool

import (
	"context"

	"emperror.dev/errors"
	longhorn "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/kubectl/pkg/scheme"
)

// HasLonghornCRD permit to check if current cluster have loghorn CRD
func (k *Kubetool) HasLonghornCRD(ctx context.Context) bool {

	if err := discovery.ServerSupportsVersion(k.client.Discovery(), longhorn.SchemeGroupVersion); err != nil {
		return false
	}

	return true
}

func (k *Kubetool) CleanPendingBackup(ctx context.Context) (err error) {

	list, err := k.dclient.Resource(longhorn.SchemeGroupVersion.WithResource("backups")).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "Error when list Longhorn backups")
	}

	s := scheme.Scheme
	if err := longhorn.AddToScheme(s); err != nil {
		panic(err)
	}

	for _, backupUnstructuredObj := range list.Items {
		// Need to convert to Backup type
		backup := &longhorn.Backup{}
		if err := s.Convert(&backupUnstructuredObj, backup, nil); err != nil {
			return errors.Wrap(err, "Error when convert to Longhorn Backup type")
		}

		if backup.Status.State == "Pending" {
			err = k.dclient.Resource(longhorn.SchemeGroupVersion.WithResource("backups")).Namespace(backup.Namespace).Delete(ctx, backup.Name, metav1.DeleteOptions{})
			if err != nil {
				return errors.Wrap(err, "Error when delete Longhorn backup on pending state")
			}
			logrus.Infof("Longhorn backup %s deleted", backup.Name)
		}
	}

	return nil
}
