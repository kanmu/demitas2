package subcmd

import (
	"fmt"

	"github.com/winebarrel/demitas2"
	"github.com/winebarrel/demitas2/utils"
)

const (
	StoneImage = "public.ecr.aws/winebarrel/stone"
)

type PortForwardCmd struct {
	Profile    string `env:"DMTS_PROFILE" short:"p" required:"" help:"Demitas profile name."`
	RemoteHost string `required:"" short:"H" help:"Remote host."`
	RemotePort uint   `required:"" short:"r"  help:"Remote port."`
	LocalPort  uint   `required:"" short:"l"  help:"Local port."`
}

func (cmd *PortForwardCmd) Run(ctx *demitas2.Context) error {
	command := fmt.Sprintf("%s:%d %d", cmd.RemoteHost, cmd.RemotePort, cmd.RemotePort)
	def, err := ctx.DefinitionOpts.Load(cmd.Profile, command, StoneImage)

	if err != nil {
		return err
	}

	taskId, interrupted, err := ctx.Ecspresso.RunUntilRunning(def, ctx.DryRun)

	if err != nil {
		return err
	}

	if ctx.DryRun {
		return nil
	}

	if taskId == "" {
		return fmt.Errorf("task ID not found")
	}

	return utils.TrapInt(
		func() error {
			if interrupted {
				return nil
			}

			containerId, err := ctx.Ecs.GetContainerId(def.Cluster, taskId)

			if err != nil {
				return fmt.Errorf("failed to get ID from container: %w", err)
			}

			fmt.Println("Start port forwarding...")
			return ctx.Ecs.StartPortForwardingSession(def.Cluster, taskId, containerId, cmd.RemotePort, cmd.LocalPort)
		},
		func() {
			fmt.Printf("Stopping task: %s\n", taskId)
			ctx.Ecs.StopTask(def.Cluster, taskId)
		},
	)
}
