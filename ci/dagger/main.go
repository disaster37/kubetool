// A generated module for Kubetool functions
//
// This module provides the CI pipeline for the kubetool project:
// build, lint, format, unit tests, acceptance tests on a k3s cluster,
// CodeCov upload, Docker image build/push and goreleaser release.

package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dagger/kubetool/internal/dagger"

	"emperror.dev/errors"
	"github.com/disaster37/dagger-library-go/lib/helper"
)

const (
	gitUsername      string = "ci"
	gitEmail         string = "ci@localhost"
	defaultGitBranch string = "main"
	registry         string = "ghcr.io"
	repository       string = "disaster37/kubetool"
	devTag           string = "dev"
	k3sVersion       string = "v1.34.3-k3s1"
	clusterName      string = "kubetool-acceptance"
)

type Kubetool struct {
	// Src is a directory that contains the projects source code
	// +private
	Src *dagger.Directory

	// +private
	GolangModule *dagger.Golang
}

func New(
	ctx context.Context,
	// a path to a directory containing the source code
	// +required
	src *dagger.Directory,
) (*Kubetool, error) {
	return &Kubetool{
		Src:          src,
		GolangModule: dag.Golang(src),
	}, nil
}

// Ci permit to run all CI steps: build, lint, format, test, acceptance tests,
// CodeCov upload, commit formatted code, Docker image build/push and
// goreleaser release on tag.
func (h *Kubetool) Ci(
	ctx context.Context,

	// Set tru if you are on CI
	// +default=false
	ci bool,

	// The image version to publish
	// +optional
	version string,

	// The registry username
	// +optional
	registryUsername *dagger.Secret,

	// The registry password
	// +optional
	registryPassword *dagger.Secret,

	// The codeCov token
	// +optional
	codeCoveToken *dagger.Secret,

	// The git branch where you should to push
	// You need to provide it when you are on PullRequest or on Tag
	// +optional
	gitBranch string,

	// Set true if current build is a tag
	// It will publish the release with goreleaser
	// +optional
	isTag bool,

	// The pull request number
	// It will publish the image with tag pr<prNumber>
	// +optional
	// +default=0
	prNumber int,

	// The git token
	// +optional
	gitToken *dagger.Secret,
) (dir *dagger.Directory, err error) {
	var stdout string

	// Build
	h.Build(ctx)

	// Lint code
	stdout, err = h.Lint(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "Error when lint project: %s", stdout)
	}

	// Format code
	dir = h.Format(ctx)

	// Test code
	reportFile, err := h.Test(ctx, false, false, "", "", "")
	if err != nil {
		return nil, errors.Wrapf(err, "Error when test project: %s", stdout)
	}

	// Generate coverage report
	dir = dir.WithFile("coverage.out", reportFile)

	// Acceptance tests on k3s cluster
	if err = h.AcceptanceTest(ctx); err != nil {
		return nil, errors.Wrap(err, "Error when run acceptance tests")
	}

	if ci {
		if codeCoveToken == nil {
			return nil, errors.New("You need to provide CodeCov token")
		}
		stdout, err = h.CodeCov(ctx, dir, codeCoveToken)
		if err != nil {
			return nil, errors.Wrapf(err, "Error when upload report on CodeCov: %s", stdout)
		}

		git := dag.GitModule(dir.WithDirectory("ci", h.Src.Directory("ci")), dagger.GitModuleOpts{Ci: "github"}).
			SetConfig(dagger.GitModuleSetConfigOpts{
				Username: gitUsername,
				Email:    gitEmail,
			})

		if isTag {
			gitBranch = defaultGitBranch
		}

		if _, err = git.CommitAndPush(
			ctx,
			gitToken,
			dagger.GitModuleCommitAndPushOpts{
				BranchName: gitBranch,
				GitRepoURL: "https://github.com/disaster37/kubetool.git",
				Message:    "Commit from CI",
			},
		); err != nil {
			return nil, errors.Wrap(err, "Error when commit and push files change")
		}
	}

	// Build docker image
	if err = h.BuildImage(ctx, ci, version, isTag, prNumber, registryUsername, registryPassword); err != nil {
		return nil, errors.Wrap(err, "Error when build and push Docker image")
	}

	// Goreleaser release (binary artifacts to GitHub Release)
	if ci && isTag {
		if gitToken == nil {
			return nil, errors.New("You need to provide git token for goreleaser release")
		}
		stdout, err = h.GoreleaserRelease(ctx, gitToken)
		if err != nil {
			return nil, errors.Wrapf(err, "Error when goreleaser release: %s", stdout)
		}
	}

	return dir, nil
}

// Lint permit to lint code
func (h *Kubetool) Lint(
	ctx context.Context,
) (string, error) {
	return h.GolangModule.Lint(ctx)
}

// Format permit to format the golang code
func (h *Kubetool) Format(
	ctx context.Context,
) *dagger.Directory {
	return h.GolangModule.Format()
}

// Test permit to run unit tests
func (h *Kubetool) Test(
	ctx context.Context,
	//+optional
	short bool,
	//+optional
	shuffle bool,
	//+optional
	run string,
	//+optional
	skip string,
	//+optional
	path string,
) (*dagger.File, error) {
	return dag.Golang(h.Src, dagger.GolangOpts{Base: h.GolangModule.Container()}).Test(dagger.GolangTestOpts{
		Short:         short,
		Shuffle:       shuffle,
		Run:           run,
		Skip:          skip,
		WithGotestsum: true,
		Path:          path,
	}), nil
}

// AcceptanceTest permit to run acceptance tests on a k3s cluster
// with the Docker image built from the project
func (h *Kubetool) AcceptanceTest(
	ctx context.Context,
) (err error) {
	// Start k3s cluster
	k3sCluster := dag.K3S(clusterName, dagger.K3SOpts{Image: fmt.Sprintf("rancher/k3s:%s", k3sVersion)})
	k3sService, err := k3sCluster.Server().Start(ctx)
	if err != nil {
		return errors.Wrap(err, "Error when start k3s cluster")
	}

	// Wait cluster ready
	// The container from the k3s module share the config cache volume with the server
	_, err = k3sCluster.Container().
		WithoutEntrypoint().
		WithServiceBinding("k3s.svc", k3sService).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec(helper.ForgeScript(`
set -e
until [ -f /etc/rancher/k3s/k3s.yaml ]; do sleep 2; done
until kubectl get nodes 2>/dev/null | grep -w Ready >/dev/null; do sleep 5; done
kubectl get nodes -o wide
`)).Sync(ctx)
	if err != nil {
		return errors.Wrap(err, "Error when wait k3s cluster ready")
	}

	// Setup fixtures and get the node name
	// The container from the k3s module share the config cache volume with the server
	stdout, err := k3sCluster.Container().
		WithoutEntrypoint().
		WithServiceBinding("k3s.svc", k3sService).
		WithMountedDirectory("/src", h.Src).
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec(helper.ForgeScript(`
set -e
kubectl create namespace test
until kubectl get serviceaccount -n test default >/dev/null 2>&1; do sleep 2; done
kubectl run test --image=alpine --namespace test --command -- tail -f /dev/null
kubectl apply -f /src/fixture/patchmanagement.yaml -n test
kubectl get pods -n test
`)).
		WithExec([]string{"sh", "-c", "kubectl get nodes -o jsonpath='{.items[0].metadata.name}'"}).
		Stdout(ctx)
	if err != nil {
		return errors.Wrap(err, "Error when setup k3s cluster")
	}
	nodeName := strings.TrimSpace(stdout)

	// Run acceptance tests with the built image
	imageContainer := h.imageContainer()

	acceptanceCtr := imageContainer.
		WithoutEntrypoint().
		WithServiceBinding("k3s.svc", k3sService).
		WithFile("/kubeconfig", k3sCluster.Config(), dagger.ContainerWithFileOpts{Owner: "65532:65532"}).
		WithEnvVariable("KUBECONFIG", "/kubeconfig")

	for _, cmd := range [][]string{
		{"kubetool", "list-worker-nodes"},
		{"kubetool", "list-master-nodes"},
		{"kubetool", "set-downtime", "--node-name", nodeName},
		{"kubetool", "unset-downtime", "--node-name", nodeName},
		{"kubetool", "run-pre-job", "--namespace", "test"},
		{"kubetool", "run-post-job", "--namespace", "test"},
		{"kubetool", "clean-evicted-pods"},
	} {
		if _, err = acceptanceCtr.WithExec(cmd).Sync(ctx); err != nil {
			return errors.Wrapf(err, "Error when run acceptance command %s", cmd)
		}
	}

	return nil
}

// Build permit to build project
func (h *Kubetool) Build(
	ctx context.Context,
) *dagger.Directory {
	return h.GolangModule.Build()
}

// CodeCov permit to upload coverage report on CodeCov
func (h *Kubetool) CodeCov(
	ctx context.Context,

	// Optional directory
	// +optional
	src *dagger.Directory,

	// The Codecov token
	// +required
	token *dagger.Secret,
) (stdout string, err error) {
	if src == nil {
		src = h.Src
	}

	return dag.Codecov().Upload(
		ctx,
		src,
		token,
		dagger.CodecovUploadOpts{
			Files: []string{"coverage.out"},
		},
	)
}

// ImageTag permit to get the image tag from the current context:
// the release number when tag, pr<prNumber> when pull request, dev otherwise
func (h *Kubetool) ImageTag(
	// The version to use when tag
	// +optional
	version string,

	// Set true if current build is a tag
	// +optional
	isTag bool,

	// The pull request number
	// +optional
	// +default=0
	prNumber int,
) string {
	return h.imageTag(version, isTag, prNumber)
}

func (h *Kubetool) imageTag(version string, isTag bool, prNumber int) string {
	if isTag {
		return version
	}
	if prNumber > 0 {
		return fmt.Sprintf("pr%d", prNumber)
	}
	return devTag
}

// imageContainer return the container built from the project Dockerfile
func (h *Kubetool) imageContainer() *dagger.Container {
	return dag.Image().Build(h.Src).GetContainer()
}

// BuildImage permit to build and push the Docker image.
// The image tag is:
//   - the release number when tag (ex: v1.0.0)
//   - pr<prNumber> when pull request (ex: pr42)
//   - dev when main branch
func (h *Kubetool) BuildImage(
	ctx context.Context,

	// Set tru if you are on CI
	// +default=false
	ci bool,

	// The image version to publish
	// +optional
	version string,

	// Set true if current build is a tag
	// +optional
	isTag bool,

	// The pull request number
	// +optional
	// +default=0
	prNumber int,

	// The registry username
	// +optional
	registryUsername *dagger.Secret,

	// The registry password
	// +optional
	registryPassword *dagger.Secret,
) (err error) {
	// lint Dockerfile
	_, err = dag.Image().Lint(ctx, h.Src)
	if err != nil {
		return errors.Wrap(err, "Error when run image Lint")
	}

	imageTag := h.imageTag(version, isTag, prNumber)

	if ci {
		_, err = dag.Image().Build(h.Src).Push(ctx, repository, imageTag, registry, dagger.ImageBuildPushOpts{WithRegistryUsername: registryUsername, WithRegistryPassword: registryPassword})
		if err != nil {
			return errors.Wrapf(err, "Error when push image '%s'", repository)
		}

	} else {
		// Force build image
		imageBuilder := h.imageContainer()
		if _, err = imageBuilder.WithoutEntrypoint().WithExec(helper.ForgeCommand("echo test")).Stdout(ctx); err != nil {
			return errors.Wrapf(err, "Error when build image '%s'", repository)
		}
	}

	return nil
}

// GoreleaserRelease build the binaries and publish them on the GitHub Release
// via goreleaser. The token GitHub is required for authentication.
func (h *Kubetool) GoreleaserRelease(
	ctx context.Context,
	// GitHub token for release creation and asset upload
	// +required
	gitToken *dagger.Secret,
) (string, error) {
	return dag.Goreleaser(
		h.Src,
		dagger.GoreleaserOpts{
			Config: h.Src.File(".goreleaser.yml"),
		},
	).
		WithSecretVariable("GITHUB_TOKEN", gitToken).
		Release().
		WithClean().
		Run(ctx)
}
