package ecscli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/winebarrel/demitas2/utils"
)

type Driver struct {
	client *ecs.Client
}

func NewDriver(cfg aws.Config) *Driver {
	return &Driver{
		client: ecs.NewFromConfig(cfg),
	}
}

func (dri *Driver) StopTask(cluster string, taskId string) error {
	input := &ecs.StopTaskInput{
		Cluster: aws.String(cluster),
		Task:    aws.String(taskId),
	}

	_, err := dri.client.StopTask(context.Background(), input)

	if err != nil {
		return fmt.Errorf("faild to call StopTask: %s/%s", cluster, taskId)
	}

	return nil
}

func (dri *Driver) GetContainerId(cluster string, taskId string) (string, error) {
	input := &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   []string{taskId},
	}

	output, err := dri.client.DescribeTasks(context.Background(), input)

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

func (dri *Driver) StartPortForwardingSession(cluster string, taskId string, containerId string, remotePort uint, localPort uint) error {
	target := fmt.Sprintf("ecs:%s_%s_%s", cluster, taskId, containerId)
	params := fmt.Sprintf(`{"portNumber":["%d"],"localPortNumber":["%d"]}`, remotePort, localPort)

	// TODO: Use AWS-StartPortForwardingSessionToRemoteHost
	//       cf. https://dev.classmethod.jp/articles/aws-ssm-support-remote-host-port-forward/
	cmdWithArgs := []string{
		"aws", "ssm", "start-session",
		"--target", target,
		"--document-name", "AWS-StartPortForwardingSession",
		"--parameters", params,
	}

	var err error

	for i := 0; i < 30; i++ {
		var stdout string

		// NOTE: https://github.com/winebarrel/demitas2/issues/2
		stdout, _, _, err = utils.RunCommand(cmdWithArgs, true)

		if err != nil {
			break
		}

		fmt.Fprintf(os.Stderr, "Faild to start session: %s\nRetrying...\n", strings.TrimSpace(stdout))
		time.Sleep(1 * time.Second)
	}

	return err
}

func buildExecuteCommand(cluster string, taskId string, command string) []string {
	return []string{
		"aws", "ecs", "execute-command",
		"--cluster", cluster,
		"--task", taskId,
		"--interactive",
		"--command", command,
	}
}

func (dri *Driver) ExecuteCommand(cluster string, taskId string, command string) error {
	cmdWithArgs := buildExecuteCommand(cluster, taskId, command)
	_, stderr, _, err := utils.RunCommand(cmdWithArgs, true)

	if err != nil {
		if stderr != "" {
			err = fmt.Errorf("%w: %s", err, stderr)
		}

		return err
	}

	return nil
}

func (dri *Driver) ExecuteInteractiveCommand(cluster string, taskId string, command string) error {
	cmdWithArgs := buildExecuteCommand(cluster, taskId, command)
	shell := exec.Command(cmdWithArgs[0], cmdWithArgs[1:]...)
	shell.Stdin = os.Stdin
	shell.Stdout = os.Stdout
	shell.Stderr = os.Stderr
	signal.Ignore(os.Interrupt)
	return shell.Run()
}
