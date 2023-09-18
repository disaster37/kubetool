package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/disaster37/kubetool/v1.28/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var version = "develop"
var commit = ""

func run(args []string) error {

	// Logger setting
	log.SetOutput(os.Stdout)

	// Get home directory
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Warnf("Can't get home directory: %s", err.Error())
		homePath = "/root"
	}

	// CLI settings
	app := cli.NewApp()
	app.Usage = "Extra kubernetes tools box"
	app.Version = fmt.Sprintf("%s-%s", version, commit)
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Usage: "Load configuration from `FILE`",
		},
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "kubeconfig",
			Usage:   "The kube config file",
			EnvVars: []string{"KUBECONFIG"},
			Value:   fmt.Sprintf("%s/.kube/config", homePath),
		}),
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Display debug output",
		},
		altsrc.NewInt64Flag(&cli.Int64Flag{
			Name:  "timeout",
			Usage: "The timeout in second",
			Value: 0,
		}),
		&cli.BoolFlag{
			Name:  "no-color",
			Usage: "No print color",
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:     "set-downtime",
			Usage:    "Run pre action on node and set it on downtime",
			Category: "Patchmanagement",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "node-name",
					Usage: "The node name",
				},
				&cli.BoolFlag{
					Name:  "retry-on-drain-failed",
					Usage: "Retry if drain failed",
					Value: false,
				},
				&cli.IntFlag{
					Name:  "number-retry",
					Usage: "How many retry",
					Value: 3,
				},
			},
			Action: cmd.SetDowntime,
		},
		{
			Name:     "unset-downtime",
			Usage:    "Unset downtime and run post action on node",
			Category: "Patchmanagement",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "node-name",
					Usage: "The node name",
				},
			},
			Action: cmd.UnsetDowntime,
		},
		{
			Name:     "list-master-nodes",
			Usage:    "List master nodes on cluster",
			Category: "Cluster",
			Action:   cmd.GetMasterNodes,
		},
		{
			Name:     "list-worker-nodes",
			Usage:    "List worker nodes on cluster",
			Category: "Cluster",
			Action:   cmd.GetWorkerNodes,
		},
		{
			Name:     "list-nodes-rundeck",
			Usage:    "List all nodes and return them as json Rundeck format",
			Category: "Cluster",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "username",
					Usage: "Username to connect on node with ssh",
				},
				&cli.StringFlag{
					Name:  "cluster-name",
					Usage: "The cluster name to append it on tags. Usefull to filter node on Rundeck",
				},
				&cli.StringFlag{
					Name:  "ssh-key-storage-path",
					Usage: "SSH key storage path to connect on node with ssh",
				},
				&cli.StringFlag{
					Name:  "ssh-password-storage-path",
					Usage: "SSH password storage path to connect on node with ssh",
				},
				&cli.StringFlag{
					Name:  "ssh-authentication",
					Usage: "SSH authentication to connect on node",
					Value: "password",
				},
			},
			Action: cmd.GetNodesForRundeck,
		},
		{
			Name:     "run-pre-job",
			Usage:    "Run pre job from given namespace",
			Category: "Patchmanagement",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "namespace",
					Usage: "Namespace where found pre job to run",
				},
			},
			Action: cmd.RunPreJob,
		},
		{
			Name:     "run-post-job",
			Usage:    "Run post job from given namespace",
			Category: "Patchmanagement",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "namespace",
					Usage: "Namespace where found post job to run",
				},
			},
			Action: cmd.RunPostJob,
		},
		{
			Name:     "clean-evicted-pods",
			Usage:    "Clean all evicted pods that failed",
			Category: "Clean",
			Flags:    []cli.Flag{},
			Action:   cmd.CleanEvictedPods,
		},
	}

	app.Before = func(c *cli.Context) error {

		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		if !c.Bool("no-color") {
			formatter := new(prefixed.TextFormatter)
			formatter.FullTimestamp = true
			formatter.ForceFormatting = true
			log.SetFormatter(formatter)
		}

		if c.String("config") != "" {
			before := altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewYamlSourceFromFlagFunc("config"))
			return before(c)
		}
		return nil
	}

	sort.Sort(cli.CommandsByName(app.Commands))

	err = app.Run(args)
	return err
}

func main() {
	err := run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
