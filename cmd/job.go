package cmd

import (
	"context"
	"os"

	"github.com/disaster37/kubetool/v1.18/kubetool"
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
	postJob, err := cmd.PostJobPatchManagement(ctx, namespace)
	if err != nil {
		return err
	}
	if postJob == "" {
		return errors.Errorf("Post job not found in namespace %s", namespace)
	}

	// Get list of secrets needed for inject in job
	secrets, err := cmd.Secrets(ctx, namespace)
	if err != nil {
		log.Errorf("Error when try to get list of secrets to inject in job on %s: %s", namespace, err.Error())
		return err
	}

	// Run postjob
	err = cmd.RunJob(ctx, namespace, "post-job", postJob, secrets)
	if err != nil {
		return err
	}

	return nil
}

func runPreJob(ctx context.Context, cmd *kubetool.Kubetool, namespace string) (err error) {

	// Get pre job
	postJob, err := cmd.PreJobPatchManagement(ctx, namespace)
	if err != nil {
		return err
	}
	if postJob == "" {
		return errors.Errorf("Pre job not found in namespace %s", namespace)
	}

	// Get list of secrets needed for inject in job
	secrets, err := cmd.Secrets(ctx, namespace)
	if err != nil {
		log.Errorf("Error when try to get list of secrets to inject in job on %s: %s", namespace, err.Error())
		return err
	}

	// Run postjob
	err = cmd.RunJob(ctx, namespace, "pre-job", postJob, secrets)
	if err != nil {
		return err
	}

	return nil
}
