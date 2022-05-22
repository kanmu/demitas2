package subcmd

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

func findTaskIdFromLog(log string) string {
	r := regexp.MustCompile(`(?s)Waiting for task ID (\S+)`)
	m := r.FindStringSubmatch(log)

	if len(m) < 2 {
		return ""
	}

	return m[1]
}

func stopTask(cfg aws.Config, cluster string, taskId string) error {
	svc := ecs.NewFromConfig(cfg)

	input := &ecs.StopTaskInput{
		Cluster: aws.String(cluster),
		Task:    aws.String(taskId),
	}

	_, err := svc.StopTask(context.Background(), input)

	if err != nil {
		return fmt.Errorf("faild to call StopTask: %s/%s", cluster, taskId)
	}

	return nil
}
