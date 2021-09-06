package cmd

import (
	"context"
	"os"

	"github.com/disaster37/kubetool/v1.20/kubetool"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// CleanEvictedPods permit to remove failed pods because of evicted
func CleanEvictedPods(c *cli.Context) error {
	cmd, err := newCmd(c)
	if err != nil {
		log.Errorf("Can't connect on kubernetes: %s", err.Error())
		os.Exit(1)
	}

	ctx, cancelFunc := getContext(c)
	if cancelFunc != nil {
		defer cancelFunc()
	}

	err = cleanEvictedPods(ctx, cmd)
	if err != nil {
		return err
	}

	log.Infof("Clean evicted pods finished successfully")
	return nil
}

func cleanEvictedPods(ctx context.Context, cmd *kubetool.Kubetool) (err error) {

	// Get post job
	err = cmd.CleanEvictedPods(ctx)
	if err != nil {
		return err
	}

	return nil
}
