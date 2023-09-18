package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/disaster37/kubetool/v1.28/kubetool"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// GetMasterNodes permit to list all master nodes
func GetMasterNodes(c *cli.Context) error {

	cmd, err := newCmd(c)
	if err != nil {
		log.Errorf("Can't connect on kubernetes: %s", err.Error())
		os.Exit(1)
	}

	ctx, cancelFunc := getContext(c)
	if cancelFunc != nil {
		defer cancelFunc()
	}

	nodes, err := getMasterNodes(ctx, cmd)
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
		log.Errorf("Can't connect on kubernetes: %s", err.Error())
		os.Exit(1)
	}

	ctx, cancelFunc := getContext(c)
	if cancelFunc != nil {
		defer cancelFunc()
	}

	workers, err := getWorkerNodes(ctx, cmd)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", strings.Join(workers, ";"))

	return nil
}

func getWorkerNodes(ctx context.Context, cmd *kubetool.Kubetool) (workers []string, err error) {
	return cmd.WorkerNodes(ctx)
}

func getMasterNodes(ctx context.Context, cmd *kubetool.Kubetool) (workers []string, err error) {
	return cmd.MasterNodes(ctx)
}
