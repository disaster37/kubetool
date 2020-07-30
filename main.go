package main

import (
	"os"
	"sort"

	"github.com/disaster37/kubetool/v1.18/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func run(args []string) error {

	// Logger setting
	log.SetOutput(os.Stdout)

	// CLI settings
	app := cli.NewApp()
	app.Usage = "Extra kubernetes tools box"
	app.Version = "develop"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Usage: "Load configuration from `FILE`",
		},
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "kubeconfig",
			Usage:   "The kube config file",
			EnvVars: []string{"KUBECONFIG"},
			Value:   "$HOME/.kube/config",
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
			},
			Action: cmd.SetDowntime,
		},
		{
			Name:     "unset-downtime",
			Usage:    "Unset downtime and run post action on node",
			Category: "Patchmanagemeent",
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
					Name:  "ssh-authentication",
					Usage: "SSH authentication to connect on node",
					Value: "password",
				},
			},
			Action: cmd.GetNodesForRundeck,
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

	err := app.Run(args)
	return err
}

func main() {
	err := run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
