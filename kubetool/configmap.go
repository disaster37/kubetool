package kubetool

import (
	"context"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PreJobPatchManagement permit to get pre script from configmap to be executed before drain node
// The configmap `patchmanagement` with key `pre-job`
func (k *Kubetool) PreJobPatchManagement(ctx context.Context, namespace string) (job string, err error) {
	log.Debugf("Namespace: %s", namespace)

	configMap, err := k.client.CoreV1().ConfigMaps(namespace).Get(ctx, "patchmanagement", metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("No pre-job found on %s", namespace)
			return "", nil
		}
		return job, err
	}

	return configMap.Data["pre-job"], err
}

// PostJobPatchManagement permit to get post script from configmap to be executed after uncordon node
// The configmap `patchmanagement` with key `post-job`
func (k *Kubetool) PostJobPatchManagement(ctx context.Context, namespace string) (job string, err error) {
	log.Debugf("Namespace: %s", namespace)

	configMap, err := k.client.CoreV1().ConfigMaps(namespace).Get(ctx, "patchmanagement", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("No post-job found on %s", namespace)
			return "", nil
		}
		return job, err
	}

	return configMap.Data["post-job"], err
}

// Secrets permit to get the list of secrets to inject on pre/post job as environment variable
// The configmap `patchmanagement` with key `secrets`
func (k *Kubetool) Secrets(ctx context.Context, namespace string) (secrets []string, err error) {
	log.Debugf("Namespace: %s", namespace)

	configMap, err := k.client.CoreV1().ConfigMaps(namespace).Get(ctx, "patchmanagement", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("No secrets found on %s", namespace)
			return nil, nil
		}
		return secrets, err
	}
	if _, ok := configMap.Data["secrets"]; !ok {
		log.Debugf("No secrets found on %s", namespace)
		return nil, nil
	}

	return strings.Split(configMap.Data["secrets"], ";"), err
}
