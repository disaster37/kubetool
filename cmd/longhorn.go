package cmd

import (
	"context"
	"os"

	"github.com/disaster37/kubetool/v1.32/kubetool"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// CleanLonghornPendingBackups is a CLI command that connects to the Kubernetes
// cluster and cleans up any Longhorn backups that are in a pending state.
// It retrieves the necessary context and command tool, checks for connection
// errors, and calls the internal function to perform the cleanup.
func CleanLonghornPendingBackups(c *cli.Context) error {

	cmd, err := newCmd(c)
	if err != nil {
		log.Errorf("Can't connect on kubernetes: %s", err.Error())
		os.Exit(1)
	}

	ctx, cancelFunc := getContext(c)
	if cancelFunc != nil {
		defer cancelFunc()
	}

	return cleanLonghornPendingBackup(ctx, cmd)
}

func cleanLonghornPendingBackup(ctx context.Context, cmd *kubetool.Kubetool) error {
	return cmd.CleanPendingBackup(ctx)
}
