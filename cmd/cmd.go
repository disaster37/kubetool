package cmd

import (
	"context"

	"github.com/disaster37/kubetool/v1.32/kubetool"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// Permit to get connexion on kubernetes
func newCmd(c *cli.Context) (cmd *kubetool.Kubetool, err error) {
	log.Debugf("Use kubeconfig: %s", c.String("kubeconfig"))

	cmd, err = kubetool.NewConnexion(c.String("kubeconfig"))

	return cmd, err
}

// Permit to get context with tiemout if needed
func getContext(c *cli.Context) (ctx context.Context, cancelFunc context.CancelFunc) {
	if c.Duration("timeout") == 0 {
		return c.Context, nil
	}

	return context.WithTimeout(c.Context, c.Duration("timeout"))
}
