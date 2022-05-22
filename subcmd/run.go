package subcmd

import (
	"fmt"
	"log"

	"github.com/winebarrel/demitas2"
	"github.com/winebarrel/demitas2/definition"
)

type RunCmd struct {
	Command string `help:"Command to run on a container."`
	Image   string `help:"Container image."`
}

func (cmd *RunCmd) Run(ctx *demitas2.Context) error {
	def, err := definition.Load(ctx.DefinitionOpts, cmd.Command, cmd.Image)

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
