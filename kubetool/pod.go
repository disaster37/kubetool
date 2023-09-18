package kubetool

import (
	"context"
	"sort"
	"time"

	"emperror.dev/errors"
	"github.com/mpvl/unique"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

// NamespacesPodsOnNode return a list of unique Namespace pod hosted on node
func (k *Kubetool) NamespacesPodsOnNode(ctx context.Context, nodeName string) (listNamespace []string, err error) {

	log.Debugf("NodeName: %s", nodeName)

	listNamespace = make([]string, 0, 1)

	pods, err := k.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
		LabelSelector: "patchmanagement=true",
	})
	if err != nil {
		return nil, err
	}
	for _, pod := range pods.Items {
		log.Debugf("Found pod %s on host %s", pod.Name, nodeName)
		listNamespace = append(listNamespace, pod.Namespace)
	}

	// remove duplicate
	sort.Strings(listNamespace)
	unique.Strings(&listNamespace)

	return listNamespace, err
}

// WaitPodsOnNode permit to wait all pods are on ready state
func (k *Kubetool) WaitPodsOnNode(ctx context.Context, nodeName string) (err error) {
	log.Debugf("NodeName: %s", nodeName)

	for {
		pods, err := k.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + nodeName,
		})
		if err != nil {
			return err
		}
		isOk := true
		for _, pod := range pods.Items {
			for _, condition := range pod.Status.Conditions {
				if condition.Type == v1.PodReady {
					if condition.Status != v1.ConditionTrue && condition.Reason != "PodCompleted" {
						log.Debugf("We wait pod %s: %s", pod.Name, condition.Reason)
						isOk = false
						break
					}
				}
			}
		}

		if isOk {
			log.Debugf("All pods are ready")
			break
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	return nil

}

// CleanEvictedPods remove all pods failed because of Evicted
func (k *Kubetool) CleanEvictedPods(ctx context.Context) (err error) {

	pods, err := k.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodFailed && pod.Status.Reason == "Evicted" {
			log.Debugf("Found pod to clean %s/%s", pod.Namespace, pod.Name)

			err = k.client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}

			log.Infof("Delete pod %s/%s successfully", pod.Namespace, pod.Name)
		}
	}

	return nil
}

func (k *Kubetool) DeleteTerminatingPodsOnNode(ctx context.Context, nodeName string, maxTime time.Duration) (err error) {

	// Get pods on node
	pods, err := k.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	if err != nil {
		return err
	}

	// Check if pod on terminating state
	for _, pod := range pods.Items {
		if pod.Spec.TerminationGracePeriodSeconds != nil && *pod.Spec.TerminationGracePeriodSeconds > 0 {
			maxTime = time.Duration(*pod.Spec.TerminationGracePeriodSeconds) * time.Second
		}
		if pod.ObjectMeta.DeletionTimestamp != nil && pod.ObjectMeta.DeletionTimestamp.Add(maxTime).Before(time.Now()) {
			log.Debugf("Force delete pod %s", pod.Name)
			if err = k.client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{
				GracePeriodSeconds: pointer.Int64(0),
			}); err != nil {
				return errors.Wrapf(err, "Error when force delete pod %s", pod.Name)
			}
		}
	}

	return nil
}
