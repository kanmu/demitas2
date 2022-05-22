package subcmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/winebarrel/demitas2"
	"github.com/winebarrel/demitas2/definition"
	"github.com/winebarrel/demitas2/utils"
	"go.uber.org/atomic"
)

type ExecCmd struct {
	Command      string `evn:"DMTS_EXEC_COMMAND" required:"" default:"bash" help:"Command to run on a container."`
	Image        string `env:"DMTS_EXEC_IMAGE" default:"public.ecr.aws/lts/ubuntu:latest" help:"Container image."`
	UseTaskImage bool   `env:"DMTS_EXEC_USE_TASK_IMAGE" help:"Use task definition image."`
	SkipStop     bool   `help:"Skip task stop."`
}

func (cmd *ExecCmd) Run(ctx *demitas2.Context) error {
	image := cmd.Image

	if cmd.UseTaskImage {
		image = ""
	}

	def, err := definition.Load(ctx.DefinitionOpts, "sleep infinity", image)
	stopped := atomic.NewBool(false)

	if err != nil {
		return err
	}

	if ctx.DryRun {
		def.Print()
		fmt.Println()
	}

	ecspressoOpts := ctx.EcspressoOpts + " --wait-until=running"
	stdout, _, err := demitas2.RunTask(ctx.EcspressoCmd, ecspressoOpts, def)

	if err != nil {
		return err
	}

	if ctx.DryRun {
		return nil
	}

	taskId := findTaskIdFromLog(stdout)

	if taskId == "" {
		return fmt.Errorf("task ID not found")
	}

	log.Printf("ECS task is running: %s", taskId)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Ignore(syscall.SIGURG)
	config.LoadDefaultConfig(context.Background())
	cluster, _ := def.EcspressoConfig.Cluster()

	teardown := func() {
		if stopped.Load() || taskId == "" {
			return
		}

		stopped.Store(true)

		if cmd.SkipStop {
			log.Printf(`ECS task is still running.

Re-login command:
  aws ecs execute-command --cluster %s --task %s --interactive --command %s

Task stop command:
  aws ecs stop-task --cluster %s --task %s`,
				cluster, taskId, cmd.Command,
				cluster, taskId,
			)

			return
		}

		log.Printf("Stopping ECS task... (Please wait for a while): %s", taskId)
		stopTask(ctx.AwsConfig, cluster, taskId)
	}

	defer teardown()

	go func() {
		<-sig
		teardown()
		os.Exit(130)
	}()

	for {
		err = cmd.executeCommand(cluster, taskId, "id")

		if err == nil {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return cmd.executeShellCommand(cluster, taskId, cmd.Command)
}

func (cmd *ExecCmd) buildExecuteCommand(cluster string, taskId string, command string) []string {
	return []string{
		"aws", "ecs", "execute-command",
		"--cluster", cluster,
		"--task", taskId,
		"--interactive",
		"--command", command,
	}
}

func (cmd *ExecCmd) executeCommand(cluster string, taskId string, command string) error {
	cmdWithArgs := cmd.buildExecuteCommand(cluster, taskId, command)
	_, _, err := utils.RunCommand(cmdWithArgs, true)
	return err
}

func (cmd *ExecCmd) executeShellCommand(cluster string, taskId string, command string) error {
	cmdWithArgs := cmd.buildExecuteCommand(cluster, taskId, command)
	shell := exec.Command(cmdWithArgs[0], cmdWithArgs[1:]...)
	shell.Stdin = os.Stdin
	shell.Stdout = os.Stdout
	shell.Stderr = os.Stderr
	signal.Ignore(os.Interrupt)
	return shell.Run()
}
