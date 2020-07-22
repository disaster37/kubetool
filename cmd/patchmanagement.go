package cmd

import (
	"context"
	"os"
	"time"

	"github.com/disaster37/kubetools/v1.18/kubetool"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// InitPatchManagement permit to lauch all step before start patch management on node
// Exit 0: all work fine
// Exit 1: Somethink wrong, we need to skip node
// Exit 2: Somethink wrong, we need to stop patchmanagement because of node is broken
func InitPatchManagement(c *cli.Context) error {

	cmd, err := newCmd(c)
	if err != nil {
		log.Error("Can't connect on kubernetes: %s", err.Error)
		os.Exit(1)
	}

	ctx, cancelFunc := getContext(c)
	if cancelFunc != nil {
		defer cancelFunc()
	}

	nodeName := c.String("node-name")

	// check the node status
	isOk, err := cmd.NodeOk(ctx, nodeName)
	if err != nil {
		log.Error("Error when check the node state for %s: %s", nodeName, err.Error())
		os.Exit(1)
	}
	if !isOk {
		log.Errorf("Node %s is not on ready state", nodeName)
		os.Exit(1)
	}

	// Cordon node
	err = cmd.Cordon(ctx, nodeName)
	if err != nil {
		log.Errorf("Error when cordon node %s: %s", nodeName, err.Error())
		os.Exit(1)
	}

	// List all namespace and lauch pre-job if needed
	namespaces, err := cmd.NamespacesPodsOnNode(ctx, nodeName)
	if err != nil {
		log.Errorf("Error when get all namespace for node %s: %s", nodeName, err.Error())
		// try to uncordon node before exit for rescue it
		uncordonNodeForRecue(cmd, nodeName)
	}
	for _, namespace := range namespaces {
		preScript, err := cmd.PreJobPatchManagement(ctx, namespace)
		if err != nil {
			log.Errorf("Error when try to get pre-job script on %s: %s", namespace, err.Error())
			// try to uncordon node before exit for rescue it
			uncordonNodeForRecue(cmd, nodeName)
		}
		if preScript != "" {
			log.Infof("Pre script found on %s, running it...", namespace)

			// Get list of secrets needed for inject in job
			secrets, err := cmd.Secrets(ctx, namespace)
			if err != nil {
				log.Errorf("Error when try to get list of secrets to inject in job on %s: %s", namespace, err.Error())
				// try to uncordon node before exit for rescue it
				uncordonNodeForRecue(cmd, nodeName)
			}

			// Run job
			err = cmd.RunJob(ctx, namespace, "pre-job", preScript, secrets)
			if err != nil {
				log.Errorf("Error when run pre-job for %s: %s", namespace, err.Error())
				// try to uncordon node before exit for rescue it
				uncordonNodeForRecue(cmd, nodeName)
			}

			log.Infof("Run pre-job successfully for %s", namespace)
		}
	}

	// Drain node
	err = cmd.Drain(ctx, nodeName, 600*time.Second)
	if err != nil {
		log.Errorf("Error when drain node %s: %s", nodeName, err.Error())

		// try to uncordon node before exit for rescue it
		uncordonNodeForRecue(cmd, nodeName)
	}

	log.Infof("Node %s is ready to be patched", nodeName)

	return nil
}

// FinalizePatchManagement permit to finalize the patch management
func FinalizePatchManagement(c *cli.Context) error {
	cmd, err := newCmd(c)
	if err != nil {
		log.Error("Can't connect on kubernetes: %s", err.Error())
		os.Exit(2)
	}

	ctx, cancelFunc := getContext(c)
	if cancelFunc != nil {
		defer cancelFunc()
	}

	nodeName := c.String("node-name")

	// wait node to be ready
	for {
		isOk, err := cmd.NodeOk(ctx, nodeName)
		if err != nil {
			log.Errorf("Error when get state of node %s: %s", nodeName, err.Error())
			os.Exit(2)
		}
		if !isOk {
			log.Errorf("Node %s is not on ready state, we wait ...", nodeName)
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
		os.Exit(2)
	}

	// Sleep and wait pods
	time.Sleep(60 * time.Second)
	err = cmd.WaitPodsOnNode(ctx, nodeName)
	if err != nil {
		log.Errorf("Error when wait pods to be started on node %s: %s", nodeName, err.Error())
		os.Exit(2)
	}

	// List all namespace and lauch post-job if needed
	namespaces, err := cmd.NamespacesPodsOnNode(ctx, nodeName)
	if err != nil {
		log.Errorf("Error when get all namespace for node %s: %s", nodeName, err.Error())
		os.Exit(1)
	}
	for _, namespace := range namespaces {
		postScript, err := cmd.PostJobPatchManagement(ctx, namespace)
		if err != nil {
			log.Errorf("Error when try to get post-job script on %s: %s", namespace, err.Error())
			os.Exit(1)
		}
		if postScript != "" {
			log.Infof("Post script found on %s, running it...", namespace)

			// Get list of secrets needed for inject in job
			secrets, err := cmd.Secrets(ctx, namespace)
			if err != nil {
				log.Errorf("Error when try to get list of secrets to inject in job on %s: %s", namespace, err.Error())
				os.Exit(1)
			}

			// Run job
			err = cmd.RunJob(ctx, namespace, "post-job", postScript, secrets)
			if err != nil {
				log.Errorf("Error when run post-job for %s: %s", namespace, err.Error())
				os.Exit(1)
			}

			log.Infof("Run post-job successfully for %s", namespace)
		}
	}

	return nil
}

// Try to uncordon node before exit
// Exit 1: the node successfully to be uncordonned
// Exit 2: The node failed to be uncordonned
func uncordonNodeForRecue(cmd *kubetool.Kubetool, nodeName string) {
	err := cmd.Uncordon(context.Background(), nodeName)
	if err != nil {
		log.Errorf("Error when try to uncordon node %s before to exit, to try to rescue it: %s", nodeName, err.Error())
		os.Exit(2)
	}

	log.Warningf("Node %s successfully uncordonned in rescue step")
	os.Exit(1)
}
