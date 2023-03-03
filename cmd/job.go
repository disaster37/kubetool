package cmd

import (
	"context"
	"os"
	"time"

	"github.com/disaster37/kubetool/v1.23/kubetool"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// RunPostJob permit to run post job on given namespace
func RunPostJob(c *cli.Context) error {
	if c.String("namespace") == "" {
		return errors.New("--namespace must be provided")
	}
	cmd, err := newCmd(c)
	if err != nil {
		log.Errorf("Can't connect on kubernetes: %s", err.Error())
		os.Exit(1)
	}

	ctx, cancelFunc := getContext(c)
	if cancelFunc != nil {
		defer cancelFunc()
	}

	err = runPostJob(ctx, cmd, c.String("namespace"))
	if err != nil {
		return err
	}

	log.Infof("Post job running successfully")
	return nil

}

// RunPreJob permit to run post job on given namespace
func RunPreJob(c *cli.Context) error {

	if c.String("namespace") == "" {
		return errors.New("--namespace must be provided")
	}
	cmd, err := newCmd(c)
	if err != nil {
		log.Errorf("Can't connect on kubernetes: %s", err.Error())
		os.Exit(1)
	}

	ctx, cancelFunc := getContext(c)
	if cancelFunc != nil {
		defer cancelFunc()
	}

	err = runPreJob(ctx, cmd, c.String("namespace"))
	if err != nil {
		return err
	}

	log.Infof("Pre job running successfully")
	return nil

}

func runPostJob(ctx context.Context, cmd *kubetool.Kubetool, namespace string) (err error) {

	// Get post job
	jobSpec, err := cmd.GetJobSpec(ctx, namespace)
	if err != nil {
		return err
	}
	if jobSpec == nil || jobSpec.PostJob == "" {
		return errors.Errorf("Post job not found in namespace %s", namespace)
	}

	// Run postjob
	ctxWithTimeout, _ := context.WithTimeout(ctx, time.Minute*30)
	err = cmd.RunJob(ctxWithTimeout, namespace, "post-job", jobSpec.PostJob, jobSpec.Image, jobSpec.SecretNames)
	if err != nil {
		return err
	}

	return nil
}

func runPreJob(ctx context.Context, cmd *kubetool.Kubetool, namespace string) (err error) {

	// Get pre job
	jobSpec, err := cmd.GetJobSpec(ctx, namespace)
	if err != nil {
		return err
	}
	if jobSpec == nil || jobSpec.PreJob == "" {
		return errors.Errorf("Pre job not found in namespace %s", namespace)
	}

	// Run postjob
	ctxWithTimeout, _ := context.WithTimeout(ctx, time.Minute*30)
	err = cmd.RunJob(ctxWithTimeout, namespace, "pre-job", jobSpec.PreJob, jobSpec.Image, jobSpec.SecretNames)
	if err != nil {
		return err
	}

	return nil
}
