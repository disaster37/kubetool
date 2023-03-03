package kubetool

import (
	"context"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetJobSpec permit to return the job spec
func (k *Kubetool) GetJobSpec(ctx context.Context, namespace string) (job *Job, err error) {
	log.Debugf("Namespace: %s", namespace)

	configMap, err := k.client.CoreV1().ConfigMaps(namespace).Get(ctx, "patchmanagement", metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("No pre-job found on %s", namespace)
			return nil, nil
		}
		return nil, err
	}

	job = &Job{
		PreJob:  configMap.Data["pre-job"],
		PostJob: configMap.Data["post-job"],
		Image:   configMap.Data["image"],
	}

	if job.Image == "" {
		job.Image = "redhat/ubi8-minimal:latest"
	}

	if _, ok := configMap.Data["secrets"]; !ok {
		log.Debugf("No secrets found on %s", namespace)
	}

	job.SecretNames = strings.Split(configMap.Data["secrets"], ";")

	return job, nil
}
