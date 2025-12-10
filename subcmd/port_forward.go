package subcmd

import (
	"fmt"
	"time"

	"github.com/kanmu/demitas2"
	"github.com/kanmu/demitas2/utils"
)

type PortForwardCmd struct {
	Profile    string `env:"DMTS_PROFILE" short:"p" help:"Demitas profile name."`
	RemoteHost string `required:"" short:"H" help:"Remote host."`
	RemotePort uint   `required:"" short:"r"  help:"Remote port."`
	LocalPort  uint   `required:"" short:"l"  help:"Local port."`
	Image      string `short:"i" default:"mirror.gcr.io/library/debian:stable-slim" help:"Container image."`
}

func (cmd *PortForwardCmd) Run(ctx *demitas2.Context) error {
	def, err := ctx.DefinitionOpts.Load(cmd.Profile, "sleep infinity", cmd.Image, 0, 0, true)

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

			time.Sleep(3 * time.Second) // wait... :-(
			fmt.Println("Start port forwarding...")
			return ctx.Ecs.StartPortForwardingSessionToRemoteHost(def.Cluster, taskId, containerId, cmd.RemoteHost, cmd.RemotePort, cmd.LocalPort)
		},
		func() {
			fmt.Printf("Stopping task: %s\n", taskId)
			ctx.Ecs.StopTask(def.Cluster, taskId) //nolint:errcheck
		},
	)
}
