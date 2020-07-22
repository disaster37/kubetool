package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// GetMasterNodes permit to list all master nodes
func GetMasterNodes(c *cli.Context) error {

	cmd, err := newCmd(c)
	if err != nil {
		log.Error("Can't connect on kubernetes: %s", err.Error)
		os.Exit(1)
	}

	ctx, cancelFunc := getContext(c)
	if cancelFunc != nil {
		defer cancelFunc()
	}

	nodes, err := cmd.MasterNodes(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", strings.Join(nodes, ";"))

	return nil
}

// GetWorkerNodes permit to list all worker nodes
func GetWorkerNodes(c *cli.Context) error {

	cmd, err := newCmd(c)
	if err != nil {
		log.Error("Can't connect on kubernetes: %s", err.Error)
		os.Exit(1)
	}

	ctx, cancelFunc := getContext(c)
	if cancelFunc != nil {
		defer cancelFunc()
	}

	nodes, err := cmd.WorkerNodes(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", strings.Join(nodes, ";"))

	return nil
}
