package kubetool

import (
	"context"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/drain"
)

// WorkerNodes permit to return the list of all worker nodes
func (k *Kubetool) WorkerNodes(ctx context.Context) (nodes []string, err error) {
	nodeList, err := k.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: "master!=true"})
	if err != nil {
		return nodes, err
	}

	for _, node := range nodeList.Items {
		nodes = append(nodes, node.Name)
	}

	return nodes, err
}

// MasterNodes permit to return the list of master nodes
func (k *Kubetool) MasterNodes(ctx context.Context) (nodes []string, err error) {
	nodeList, err := k.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: "master=true"})
	if err != nil {
		return nodes, err
	}

	for _, node := range nodeList.Items {
		nodes = append(nodes, node.Name)
	}

	return nodes, err
}

// Nodes permit to return the list of all nodes
func (k *Kubetool) Nodes(ctx context.Context) (nodes []string, err error) {
	nodeList, err := k.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nodes, err
	}

	for _, node := range nodeList.Items {
		nodes = append(nodes, node.Name)
	}

	return nodes, err
}

// Drain permit to drain a node
func (k *Kubetool) Drain(ctx context.Context, nodeName string, timeout time.Duration) (err error) {
	log.Debugf("NodeName: %s", nodeName)

	node, err := k.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	log.Debugf("Node %s found", node.Name)

	drainer := &drain.Helper{
		Ctx:                 ctx,
		Client:              k.client,
		DeleteLocalData:     false,
		IgnoreAllDaemonSets: true,
		Timeout:             timeout,
		GracePeriodSeconds:  -1,
		Out:                 os.Stdout,
		ErrOut:              os.Stderr,
	}

	err = drain.RunNodeDrain(drainer, node.Name)

	return err
}

// Cordon permit to cordon the node
func (k *Kubetool) Cordon(ctx context.Context, nodeName string) (err error) {
	log.Debugf("NodeName: %s", nodeName)

	node, err := k.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	log.Debugf("Node %s found", node.Name)

	cordoner := drain.NewCordonHelper(node)

	if cordoner.UpdateIfRequired(true) {
		log.Debugf("Cordon node %s", node.Name)

		err, err2 := cordoner.PatchOrReplace(k.client, false)
		if err != nil {
			return err
		}
		if err2 != nil {
			return err2
		}
	} else {
		log.Debugf("Node %s already cordoned", node.Name)
	}

	return err
}

// Uncordon permit to ucordon the node
func (k *Kubetool) Uncordon(ctx context.Context, nodeName string) (err error) {
	log.Debugf("NodeName: %s", nodeName)

	node, err := k.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	log.Debugf("Node %s found", node.Name)

	cordoner := drain.NewCordonHelper(node)

	if cordoner.UpdateIfRequired(false) {
		log.Debugf("Uncordon node %s", node.Name)

		err, err2 := cordoner.PatchOrReplace(k.client, false)
		if err != nil {
			return err
		}
		if err2 != nil {
			return err2
		}
	} else {
		log.Debugf("Node %s already uncordoned", node.Name)
	}

	return err
}

// NodeOk permit to check if node is OK
func (k *Kubetool) NodeOk(ctx context.Context, nodeName string) (isOk bool, err error) {
	log.Debugf("NodeName: %s", nodeName)

	node, err := k.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return isOk, err
	}
	log.Debugf("Node %s found", node.Name)

	isOk = false
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
			log.Debugf("Node %s ready", nodeName)
			isOk = true
			break
		}
	}

	return isOk, err

}
