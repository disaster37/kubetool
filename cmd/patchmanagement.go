package cmd

import (
	"context"
	"os"
	"time"

	"github.com/disaster37/kubetools/v1.18/kubetool"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// SetDowntime permit to put node on downtime
// It will look all pod running on node. For each, it try to retrieve some pre-job to execuste from configmap patchmanagement.
// Exit 0: all work fine
// Exit 1: Somethink wrong, we need to skip node
// Exit 2: Somethink wrong, we need to stop patchmanagement because of node is broken
func SetDowntime(c *cli.Context) error {

	cmd, err := newCmd(c)
	if err != nil {
		log.Errorf("Can't connect on kubernetes: %s", err.Error())
		os.Exit(1)
	}

	ctx, cancelFunc := getContext(c)
	if cancelFunc != nil {
		defer cancelFunc()
	}

	nodeName := c.String("node-name")

	err = setDowntime(ctx, cmd, nodeName)
	if err != nil {

		err = uncordonNodeForRecue(cmd, nodeName)
		if err != nil {
			// Rescue failed
			log.Error(err.Error())
			os.Exit(2)
		}

		// Rescue success
		os.Exit(1)
	}

	return nil

}

// UnsetDowntime permit to lauch some step after enable node
func UnsetDowntime(c *cli.Context) error {
	cmd, err := newCmd(c)
	if err != nil {
		log.Errorf("Can't connect on kubernetes: %s", err.Error())
		os.Exit(2)
	}

	ctx, cancelFunc := getContext(c)
	if cancelFunc != nil {
		defer cancelFunc()
	}

	nodeName := c.String("node-name")

	err = unsetDowntime(ctx, cmd, nodeName)
	if err != nil {
		err = uncordonNodeForRecue(cmd, nodeName)
		if err != nil {
			// Rescue failed
			log.Error(err.Error())
			os.Exit(2)
		}

		// Rescue success
		os.Exit(1)
	}

	return nil
}

// Try to uncordon node before exit
func uncordonNodeForRecue(cmd *kubetool.Kubetool, nodeName string) (err error) {
	err = cmd.Uncordon(context.Background(), nodeName)
	if err != nil {
		log.Errorf("Error when try to uncordon node %s on rescue step", nodeName)
		return err
	}

	log.Warningf("Node %s successfully uncordonned in rescue step", nodeName)
	return nil
}

func setDowntime(ctx context.Context, cmd *kubetool.Kubetool, nodeName string) (err error) {
	// check the node status
	isOk, err := cmd.NodeOk(ctx, nodeName)
	if err != nil {
		log.Errorf("Error when check the node state for %s", nodeName)
		return err
	}
	if !isOk {
		log.Errorf("Node %s is not on ready state", nodeName)
		return kubetool.NewErrNodeNotReady(nodeName)
	}

	// Cordon node
	err = cmd.Cordon(ctx, nodeName)
	if err != nil {
		log.Errorf("Error when cordon node %s", nodeName)
		return err
	}

	// List all namespace and lauch pre-job if needed
	namespaces, err := cmd.NamespacesPodsOnNode(ctx, nodeName)
	if err != nil {
		log.Errorf("Error when get all namespace for node %s", nodeName)
		return err
	}
	for _, namespace := range namespaces {
		preScript, err := cmd.PreJobPatchManagement(ctx, namespace)
		if err != nil {
			log.Errorf("Error when try to get pre-job script on %s", namespace)
			return err
		}
		if preScript != "" {
			log.Infof("Pre script found on %s, running it...", namespace)

			// Get list of secrets needed for inject in job
			secrets, err := cmd.Secrets(ctx, namespace)
			if err != nil {
				log.Errorf("Error when try to get list of secrets to inject in job on %s", namespace)
				return err
			}

			// Run job
			err = cmd.RunJob(ctx, namespace, "pre-job", preScript, secrets)
			if err != nil {
				log.Errorf("Error when run pre-job for %s", namespace)
				return err
			}

			log.Infof("Run pre-job successfully for %s", namespace)
		}
	}

	// Drain node
	err = cmd.Drain(ctx, nodeName, 600*time.Second)
	if err != nil {
		log.Errorf("Error when drain node %s", nodeName)
		return err
	}

	log.Infof("Node %s is ready to be patched", nodeName)

	return nil
}

func unsetDowntime(ctx context.Context, cmd *kubetool.Kubetool, nodeName string) (err error) {

	// wait node to be ready
	for {
		isOk, err := cmd.NodeOk(ctx, nodeName)
		if err != nil {
			log.Errorf("Error when get state of node %s: %s", nodeName, err.Error())
			return err
		}
		if !isOk {
			log.Infof("Node %s is not on ready state, we wait ...", nodeName)
			time.Sleep(10 * time.Second)
		} else {
			log.Debugf("Node %s is ready", nodeName)
			break
		}
	}

	// Uncordon the node
	err = cmd.Uncordon(ctx, nodeName)
	if err != nil {
		log.Errorf("Error when uncordon node %s: %s", nodeName, err.Error())
		return err
	}

	// Sleep and wait pods
	time.Sleep(30 * time.Second)
	err = cmd.WaitPodsOnNode(ctx, nodeName)
	if err != nil {
		log.Errorf("Error when wait pods to be started on node %s: %s", nodeName, err.Error())
		return err
	}

	// List all namespace and lauch post-job if needed
	namespaces, err := cmd.NamespacesPodsOnNode(ctx, nodeName)
	if err != nil {
		log.Errorf("Error when get all namespace for node %s: %s", nodeName, err.Error())
		return err
	}
	for _, namespace := range namespaces {
		postScript, err := cmd.PostJobPatchManagement(ctx, namespace)
		if err != nil {
			log.Errorf("Error when try to get post-job script on %s: %s", namespace, err.Error())
			return err
		}
		if postScript != "" {
			log.Infof("Post script found on %s, running it...", namespace)

			// Get list of secrets needed for inject in job
			secrets, err := cmd.Secrets(ctx, namespace)
			if err != nil {
				log.Errorf("Error when try to get list of secrets to inject in job on %s: %s", namespace, err.Error())
				return err
			}

			// Run job
			err = cmd.RunJob(ctx, namespace, "post-job", postScript, secrets)
			if err != nil {
				log.Errorf("Error when run post-job for %s: %s", namespace, err.Error())
				return err
			}

			log.Infof("Run post-job successfully for %s", namespace)
		}
	}

	return nil
}
