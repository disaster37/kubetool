package kubetool

import (
	"context"
	"sort"
	"time"

	"github.com/mpvl/unique"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
