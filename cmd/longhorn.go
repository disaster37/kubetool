package cmd

import (
	"context"
	"os"
	"time"

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

// CleanLonghornOldOrphanBackup is the cli command to clean orphan backup that are older than.
func CleanLonghornOldOrphanBackup(c *cli.Context) error {
	cmd, err := newCmd(c)
	if err != nil {
		log.Errorf("Can't connect on kubernetes: %s", err.Error())
		os.Exit(1)
	}

	ctx, cancelFunc := getContext(c)
	if cancelFunc != nil {
		defer cancelFunc()
	}

	return cleanLonghornOldOrphanBackup(ctx, cmd, c.Duration("older-than"))
}

func cleanLonghornOldOrphanBackup(ctx context.Context, cmd *kubetool.Kubetool, olderThan time.Duration) error {
	return cmd.CleanOrphanBackup(ctx, olderThan)
}
