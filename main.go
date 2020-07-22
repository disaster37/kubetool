package main

import (
	"os"
	"sort"

	"github.com/disaster37/kubetools/v1.18/cmd"
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
			Name:     "pre-patchmanagement",
			Usage:    "Run pre patchmanagement action on node",
			Category: "Patchmanagement",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "node-name",
					Usage: "The node name",
				},
			},
			Action: cmd.InitPatchManagement,
		},
		{
			Name:     "post-patchmanagement",
			Usage:    "Run post patchmanagement action on nodes",
			Category: "Patchmanagemeent",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "node-name",
					Usage: "The node name",
				},
			},
			Action: cmd.FinalizePatchManagement,
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
