package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// RundeckNodeEntry represent node entry for Rundeck
type RundeckNodeEntry struct {
	NodeName          string `json:"nodename"`
	Hostname          string `json:"hostname,omitempty"`
	Username          string `json:"username,omitempty"`
	Tags              string `json:"tags,omitempty"`
	SshKeyStoragePath string `json:"ssh-key-storage-path,omitempty"`
}

// GetNodesForRundeck permit to list all nodes and return Rundeck node entry format
func GetNodesForRundeck(c *cli.Context) error {

	cmd, err := newCmd(c)
	if err != nil {
		log.Errorf("Can't connect on kubernetes: %s", err.Error())
		os.Exit(1)
	}

	ctx, cancelFunc := getContext(c)
	if cancelFunc != nil {
		defer cancelFunc()
	}

	result := map[string]RundeckNodeEntry{}

	// Process master nodes
	nodes, err := getMasterNodes(ctx, cmd)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		result[node] = RundeckNodeEntry{
			NodeName:          node,
			Hostname:          node,
			SshKeyStoragePath: c.String("ssh-key-storage-path"),
			Username:          c.String("username"),
			Tags:              fmt.Sprintf("%s,master", c.String("cluster-name")),
		}
	}

	// Process worker nodes
	nodes, err = getWorkerNodes(ctx, cmd)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		result[node] = RundeckNodeEntry{
			NodeName:          node,
			Hostname:          node,
			SshKeyStoragePath: c.String("ssh-key-storage-path"),
			Username:          c.String("username"),
			Tags:              fmt.Sprintf("%s,worker", c.String("cluster-name")),
		}
	}

	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	fmt.Println(string(b))

	return nil
}
