package subcmd

import (
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

	taskId, _, err := ctx.Ecspresso.RunUntilStopped(def, ctx.DryRun)

	defer func() {
		ctx.Ecs.StopTask(def.Cluster, taskId)
	}()

	return err
}
