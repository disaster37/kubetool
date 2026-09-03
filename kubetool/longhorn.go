package kubetool

import (
	"context"
	"time"

	"emperror.dev/errors"
	longhorn "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utiljson "k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/discovery"
)

// HasLonghornCRD permit to check if current cluster have loghorn CRD
func (k *Kubetool) HasLonghornCRD(ctx context.Context) bool {
	if err := discovery.ServerSupportsVersion(k.client.Discovery(), longhorn.SchemeGroupVersion); err != nil {
		return false
	}

	return true
}

// convertUnstructuredToTyped converts an unstructured object to its typed form
// through a JSON round-trip. The reflection based converter used by
// scheme.Convert and runtime.DefaultUnstructuredConverter cannot decode fields
// using the encoding/json ",string" option (like Longhorn's Volume.Spec.Size
// int64 `json:"size,string"`), which fails with "unrecognized type: int64" on
// real cluster objects. A JSON round-trip honors those tags.
func convertUnstructuredToTyped(u runtime.Object, out interface{}) error {
	data, err := utiljson.Marshal(u)
	if err != nil {
		return err
	}
	return utiljson.Unmarshal(data, out)
}

func (k *Kubetool) CleanPendingBackup(ctx context.Context) (err error) {
	list, err := k.dclient.Resource(longhorn.SchemeGroupVersion.WithResource("backups")).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "Error when list Longhorn backups")
	}

	for _, backupUnstructuredObj := range list.Items {
		// Need to convert to Backup type
		backup := &longhorn.Backup{}
		if err := convertUnstructuredToTyped(&backupUnstructuredObj, backup); err != nil {
			return errors.Wrap(err, "Error when convert to Longhorn Backup type")
		}

		if backup.Status.State == "Pending" {
			err = k.dclient.Resource(longhorn.SchemeGroupVersion.WithResource("backups")).Namespace(backup.Namespace).Delete(ctx, backup.Name, metav1.DeleteOptions{})
			if err != nil {
				return errors.Wrap(err, "Error when delete Longhorn backup on pending state")
			}
			log.Infof("Longhorn backup %s deleted", backup.Name)
		}
	}

	return nil
}

// CleanOrphanBackup deletes the Longhorn backups whose volume does not exist
// anymore and which are older than the given duration. A backup with an
// unparsable creation date, or that cannot be deleted, is logged and skipped
// so one bad record does not stop the whole cleanup.
func (k *Kubetool) CleanOrphanBackup(ctx context.Context, olderThan time.Duration) (err error) {
	listBackup, err := k.dclient.Resource(longhorn.SchemeGroupVersion.WithResource("backups")).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "Error when list Longhorn backups")
	}

	listVolumeUnstructured, err := k.dclient.Resource(longhorn.SchemeGroupVersion.WithResource("volumes")).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "Error when list Longhorn volumes")
	}

	// Keep an index of existing volume names to avoid a nested loop per backup
	existingVolumes := make(map[string]struct{}, len(listVolumeUnstructured.Items))
	for _, volumeUnstructuredObj := range listVolumeUnstructured.Items {
		volume := &longhorn.Volume{}
		if err := convertUnstructuredToTyped(&volumeUnstructuredObj, volume); err != nil {
			return errors.Wrap(err, "Error when convert to Longhorn Volume type")
		}
		existingVolumes[volume.Name] = struct{}{}
	}

	for _, backupUnstructuredObj := range listBackup.Items {
		// Need to convert to Backup type
		backup := &longhorn.Backup{}
		if err := convertUnstructuredToTyped(&backupUnstructuredObj, backup); err != nil {
			return errors.Wrap(err, "Error when convert to Longhorn Backup type")
		}

		// Skip backups whose volume still exists
		if _, isFound := existingVolumes[backup.Status.VolumeName]; isFound {
			log.Debugf("Backup %s still has volume %s, skip", backup.Name, backup.Status.VolumeName)
			continue
		}

		createdAt, err := time.Parse(time.RFC3339, backup.Status.BackupCreatedAt)
		if err != nil {
			log.Warnf("Skip orphan backup %s: could not parse creation date %q: %s", backup.Name, backup.Status.BackupCreatedAt, err.Error())
			continue
		}

		if time.Since(createdAt) <= olderThan {
			log.Infof("Orphan Longhorn backup %s is not yet older than %s (%s)", backup.Name, olderThan.String(), createdAt.String())
			continue
		}

		if err := k.dclient.Resource(longhorn.SchemeGroupVersion.WithResource("backups")).
			Namespace(backup.Namespace).
			Delete(ctx, backup.Name, metav1.DeleteOptions{}); err != nil {
			// Do not stop the whole cleanup on a single delete failure
			log.Errorf("Error when delete orphan Longhorn backup %s: %s", backup.Name, err.Error())
			continue
		}
		log.Infof("Orphan Longhorn backup %s deleted", backup.Name)
	}

	return nil
}
