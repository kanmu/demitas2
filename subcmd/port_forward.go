package subcmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/winebarrel/demitas2"
	"github.com/winebarrel/demitas2/definition"
	"github.com/winebarrel/demitas2/utils"
	"go.uber.org/atomic"
)

const (
	StoneImage = "public.ecr.aws/winebarrel/stone"
)

type PortForwardCmd struct {
	RemoteHost string `required:"" short:"H" help:"Remote host."`
	RemotePort uint   `required:"" short:"r"  help:"Remote port."`
	LocalPort  uint   `required:"" short:"l"  help:"Local port."`
}

func (cmd *PortForwardCmd) Run(ctx *demitas2.Context) error {
	command := fmt.Sprintf("%s:%d %d", cmd.RemoteHost, cmd.RemotePort, cmd.RemotePort)
	def, err := definition.Load(ctx.DefinitionOpts, command, StoneImage)
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

		log.Printf("Stopping ECS task... (Please wait for a while): %s", taskId)
		stopTask(ctx.AwsConfig, cluster, taskId)
	}

	defer teardown()

	go func() {
		<-sig
		teardown()
		os.Exit(130)
	}()

	containerId, err := cmd.getContainerId(ctx.AwsConfig, cluster, taskId)

	if err != nil {
		return fmt.Errorf("failed to get ID from container: %w", err)
	}

	log.Print("Start port forwarding...")

	return cmd.startSession(cluster, taskId, containerId)
}

func (cmd *PortForwardCmd) getContainerId(cfg aws.Config, cluster string, taskId string) (string, error) {
	svc := ecs.NewFromConfig(cfg)

	input := &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   []string{taskId},
	}

	output, err := svc.DescribeTasks(context.Background(), input)

	if err != nil {
		return "", fmt.Errorf("faild to call DescribeTasks: %s/%s", taskId, cluster)
	}

	if len(output.Tasks) == 0 {
		return "", fmt.Errorf("task not found: %s/%s", taskId, cluster)
	}

	task := output.Tasks[0]

	if len(task.Containers) == 0 {
		return "", fmt.Errorf("container not found: %s/%s", taskId, cluster)
	}

	return *task.Containers[0].RuntimeId, nil
}

func (cmd *PortForwardCmd) startSession(cluster string, taskId string, containerId string) error {
	target := fmt.Sprintf("ecs:%s_%s_%s", cluster, taskId, containerId)
	params := fmt.Sprintf(`{"portNumber":["%d"],"localPortNumber":["%d"]}`, cmd.RemotePort, cmd.LocalPort)

	cmdWithArgs := []string{
		"aws", "ssm", "start-session",
		"--target", target,
		"--document-name", "AWS-StartPortForwardingSession",
		"--parameters", params,
	}

	_, _, err := utils.RunCommand(cmdWithArgs, true)

	return err
}
