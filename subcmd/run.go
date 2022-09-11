package subcmd

import (
	"fmt"

	"github.com/winebarrel/demitas2"
)

type RunCmd struct {
	Profile string `env:"DMTS_PROFILE" short:"p" help:"Demitas profile name."`
	Command string `help:"Command to run on a container."`
	Image   string `help:"Container image."`
	Cpu     uint64 `help:"Task CPU limit."`
	Memory  uint64 `help:"Task memory limit."`
	Detach  bool   `help:"Detach when the task starts."`
}

func (cmd *RunCmd) Run(ctx *demitas2.Context) error {
	def, err := ctx.DefinitionOpts.Load(cmd.Profile, cmd.Command, cmd.Image, cmd.Cpu, cmd.Memory)

	if err != nil {
		return err
	}

	if cmd.Detach {
		taskId, interrupted, err := ctx.Ecspresso.RunUntilRunning(def, ctx.DryRun)

		if err != nil {
			return err
		}

		if interrupted {
			return nil
		}

		fmt.Printf(`ECS task is still running.

Login command:
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
