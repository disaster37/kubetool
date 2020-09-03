package kubetool

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RunJob permit to execute script as Job in kubernetes cluster
func (k *Kubetool) RunJob(ctx context.Context, namespace string, jobName string, job string, secrets []string) (err error) {
	if job == "" {
		log.Info("Empty job, skip it")
		return err
	}

	longJobName := fmt.Sprintf("patchmanagement-%s", jobName)
	backOffLimit := int32(4)
	deleteOption := meta.DeletePropagationForeground

	// Check if old job already exist
	jobObj, err := k.client.BatchV1().Jobs(namespace).Get(ctx, longJobName, meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			log.Debugf("Job %s not found on %s", longJobName, namespace)
			jobObj = nil
		} else {
			return err
		}
	}
	if jobObj != nil {
		log.Debugf("Found old job %s, try to remove it", longJobName)
		err := k.client.BatchV1().Jobs(namespace).Delete(ctx, longJobName, meta.DeleteOptions{PropagationPolicy: &deleteOption})
		if err != nil {
			return err
		}

		// We wait job is deleted
		for {
			_, err = k.client.BatchV1().Jobs(namespace).Get(ctx, longJobName, meta.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					log.Debugf("Job %s is deleted on %s", longJobName, namespace)
					break
				} else {
					return err
				}
			}

			log.Debugf("We wait job %s on %s be deleted", longJobName, namespace)
			time.Sleep(5 * time.Second)
		}
	}

	// Compte secret reference
	secretList := make([]core.EnvFromSource, 0, len(secrets))

	for _, secret := range secrets {
		secretList = append(secretList, core.EnvFromSource{
			SecretRef: &core.SecretEnvSource{LocalObjectReference: core.LocalObjectReference{Name: secret}},
		})
	}

	jobObj = &batch.Job{
		TypeMeta: meta.TypeMeta{
			Kind: "Job",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: longJobName,
		},
		Spec: batch.JobSpec{
			BackoffLimit: &backOffLimit,
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Name: jobName,
				},
				Spec: core.PodSpec{
					RestartPolicy: "Never",
					Containers: []core.Container{
						{
							Name:  jobName,
							Image: "centos:7",
							Command: []string{
								"/bin/sh",
							},
							Args:    []string{"-c", job},
							EnvFrom: secretList,
							Resources: core.ResourceRequirements{
								Limits: core.ResourceList{
									"cpu":    resource.MustParse("1000m"),
									"memory": resource.MustParse("512Mi"),
								},
								Requests: core.ResourceList{
									"cpu":    resource.MustParse("1000m"),
									"memory": resource.MustParse("512Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	jobObj, err = k.client.BatchV1().Jobs(namespace).Create(ctx, jobObj, meta.CreateOptions{})
	if err != nil {
		return err
	}

	getLogs := func(jobName string) error {
		podList, err := k.client.CoreV1().Pods(namespace).List(ctx, meta.ListOptions{LabelSelector: "job-name=" + jobObj.Name})
		if err != nil {
			return err
		}
		for _, pod := range podList.Items {
			req := k.client.CoreV1().Pods(namespace).GetLogs(pod.Name, &core.PodLogOptions{})
			podLogs, err := req.Stream(ctx)
			if err != nil {
				return err
			}
			defer podLogs.Close()
			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, podLogs)
			if err != nil {
				return err
			}
			log.Infof("Pod %s:\n%s", pod.Name, buf.String())
		}

		return nil
	}

	// Wait job completion
	for {
		jobObj, err = k.client.BatchV1().Jobs(namespace).Get(ctx, longJobName, meta.GetOptions{})
		if err != nil {
			return err
		}

		for _, condition := range jobObj.Status.Conditions {
			if condition.Type == batch.JobFailed && condition.Status == core.ConditionTrue {
				err = getLogs(jobObj.Name)
				if err != nil {
					log.Errorf("Error when try to read pods logs: %s", err.Error())
				}
				return errors.Errorf("Job %s failed: %s", longJobName, condition.Reason)
			} else if condition.Type == batch.JobComplete && condition.Status == core.ConditionTrue {
				log.Debugf("Job %s/%s completed successfully", namespace, longJobName)
				err = getLogs(jobObj.Name)
				if err != nil {
					log.Errorf("Error when try to read pods logs: %s", err.Error())
				}
				return nil
			}
		}

		time.Sleep(5 * time.Second)
	}

}
