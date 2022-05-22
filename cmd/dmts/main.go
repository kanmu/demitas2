package main

import (
	"context"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/posener/complete"
	"github.com/willabides/kongplete"
	"github.com/winebarrel/demitas2"
	"github.com/winebarrel/demitas2/definition"
	"github.com/winebarrel/demitas2/subcmd"
)

var version string

var cli struct {
	Version       kong.VersionFlag
	EcspressoCmd  string `env:"ECSPRESSO_CMD" required:"" default:"ecspresso" help:"ecspresso command path."`
	EcspressoOpts string `env:"ECSPRESSO_OPTS" short:"X" help:"Options passed to ecspresso."`
	DryRun        bool   `default:"false" help:"Run ecspresso with dry-run."`
	definition.DefinitionOpts
	Run                subcmd.RunCmd                `cmd:"" help:"Run ECS task."`
	Exec               subcmd.ExecCmd               `cmd:"" help:"Run ECS task and execute a command on a container."`
	PortForward        subcmd.PortForwardCmd        `cmd:"" help:"Forward a local port to a container."`
	Profiles           subcmd.ProfilesCmd           `cmd:"" help:"List profiles."`
	InstallCompletions kongplete.InstallCompletions `cmd:"" help:"Install shell completions"`
}

func init() {
	log.SetFlags(0)
}

func main() {
	parser := kong.Must(&cli, kong.Vars{"version": version})

	kongplete.Complete(parser,
		kongplete.WithPredictor("file", complete.PredictFiles("*")),
	)

	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	cfg, err := config.LoadDefaultConfig(context.Background())

	if err != nil {
		panic(err)
	}

	ecspressoOpts := cli.EcspressoOpts

	if cli.DryRun {
		ecspressoOpts += " --dry-run"
	}

	err = ctx.Run(&demitas2.Context{
		EcspressoCmd:   cli.EcspressoCmd,
		EcspressoOpts:  ecspressoOpts,
		DryRun:         cli.DryRun,
		DefinitionOpts: &cli.DefinitionOpts,
		AwsConfig:      cfg,
	})

	ctx.FatalIfErrorf(err)
}
