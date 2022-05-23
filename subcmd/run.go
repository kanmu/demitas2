package subcmd

import (
	"fmt"

	"github.com/winebarrel/demitas2"
)

type RunCmd struct {
	Profile  string `env:"DMTS_PROFILE" short:"p" required:"" help:"Demitas profile name."`
	Command  string `help:"Command to run on a container."`
	Image    string `help:"Container image."`
	SkipStop bool   `help:"Skip task stop."`
}

func (cmd *RunCmd) Run(ctx *demitas2.Context) error {
	def, err := ctx.DefinitionOpts.Load(cmd.Profile, cmd.Command, cmd.Image)

	if err != nil {
		return err
	}

	if cmd.SkipStop {
		taskId, interrupted, err := ctx.Ecspresso.RunUntilRunning(def, ctx.DryRun)

		if err != nil {
			return err
		}

		if interrupted {
			return nil
		}

		fmt.Printf(`ECS task is still running.

Re-login command:
  aws ecs execute-command --cluster %s --task %s --interactive --command bash

Task stop command:
  aws ecs stop-task --cluster %s --task %s
`,
			def.Cluster, taskId,
			def.Cluster, taskId,
		)

		return nil
	} else {
		taskId, _, err := ctx.Ecspresso.RunUntilStopped(def, ctx.DryRun)

		defer func() {
			ctx.Ecs.StopTask(def.Cluster, taskId)
		}()

		return err
	}
}
