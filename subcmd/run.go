package subcmd

import (
	"fmt"
	"log"

	"github.com/winebarrel/demitas2"
)

type RunCmd struct {
	Profile string `env:"DMTS_PROFILE" short:"p" required:"" help:"Demitas profile name."`
	Command string `help:"Command to run on a container."`
	Image   string `help:"Container image."`
}

func (cmd *RunCmd) Run(ctx *demitas2.Context) error {
	def, err := ctx.DefinitionOpts.Load(cmd.Profile, cmd.Command, cmd.Image)

	if err != nil {
		return err
	}

	if ctx.DryRun {
		def.Print()
		fmt.Println()
	}

	stdout, _, _, err := demitas2.RunTask(ctx.EcspressoCmd, ctx.EcspressoOpts, def)
	taskId := findTaskIdFromLog(stdout)

	defer func() {
		log.Printf("Stopping ECS task... (Please wait for a while): %s", taskId)
		cluster, _ := def.EcspressoConfig.Cluster()
		stopTask(ctx.AwsConfig, cluster, taskId)
	}()

	return err
}
