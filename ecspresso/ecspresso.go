package ecspresso

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/kanmu/demitas2/definition"
	"github.com/kanmu/demitas2/utils"
	"github.com/mattn/go-shellwords"
)

const (
	taskDefFile      = "task-def.json"
	serviceDefFile   = "service-def.json"
	ecspressoCfgFile = "ecspresso.yml"
)

type Ecspresso struct {
	path    string
	options string
}

func NewEcspresso(path string, opts string) (*Ecspresso, error) {
	out, err := exec.Command(path, "version").CombinedOutput()

	if err != nil {
		return nil, fmt.Errorf("faild to execute ecspresso: %w: %s", err, out)
	}

	return &Ecspresso{
		path:    path,
		options: opts,
	}, nil
}

func (ecsp *Ecspresso) RunUntilRunning(def *definition.Definition, dryRun bool) (taskId string, interrupted bool, err error) {
	return ecsp.run(def, dryRun, true)
}

func (ecsp *Ecspresso) RunUntilStopped(def *definition.Definition, dryRun bool) (taskId string, interrupted bool, err error) {
	return ecsp.run(def, dryRun, false)
}

func (ecsp *Ecspresso) run(def *definition.Definition, dryRun bool, untilRunning bool) (taskId string, interrupted bool, err error) {
	opts := ecsp.options

	if untilRunning {
		opts += " --wait-until=running"
	}

	if dryRun {
		opts += " --dry-run"
	}

	var stdout string

	runInTempDir(func() {
		err = writeTemporaryConfigs(def.EcspressoConfig, def.Service, def.Task)

		if err != nil {
			return
		}

		cmdWithArgs := []string{ecsp.path, "run"}
		args, err := shellwords.Parse(opts)

		if err != nil {
			return
		}

		if len(args) > 0 {
			cmdWithArgs = append(cmdWithArgs, args...)
		}

		stdout, _, interrupted, err = utils.RunCommand(cmdWithArgs, false)

		if err != nil {
			return
		}
	})

	if dryRun {
		def.Print()
		fmt.Println()
		return
	}

	taskId = findTaskIdFromLog(stdout)

	return
}

func runInTempDir(callback func()) {
	pwd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	tmp, err := ioutil.TempDir("", "demitas2")

	if err != nil {
		panic(err)
	}

	defer func() {
		_ = os.Chdir(pwd)
		os.RemoveAll(tmp)
	}()

	err = os.Chdir(tmp)

	if err != nil {
		panic(err)
	}

	callback()
}

func writeTemporaryConfigs(ecsConf *definition.EcspressoConfig, svrDef *definition.ServiceDefinition, taskDef *definition.TaskDefinition) error {
	err := ioutil.WriteFile(taskDefFile, taskDef.Content, os.FileMode(0o644))

	if err != nil {
		return fmt.Errorf("failed to write ECS task definition: %w", err)
	}

	err = ioutil.WriteFile(serviceDefFile, svrDef.Content, os.FileMode(0o644))

	if err != nil {
		return fmt.Errorf("failed to write ECS service definition: %w", err)
	}

	ecsConfOverrides := `{"service_definition":"` + serviceDefFile + `","task_definition":"` + taskDefFile + `"}`
	ecsConfJson, err := jsonpatch.MergePatch(ecsConf.Content, []byte(ecsConfOverrides))

	if err != nil {
		return fmt.Errorf("failed to update temporary ecspresso config: %w", err)
	}

	ecsConfYaml, err := utils.JSONToYAML(ecsConfJson)

	if err != nil {
		return fmt.Errorf("failed to convert temporary ecspresso config to yaml: %w", err)
	}

	err = ioutil.WriteFile(ecspressoCfgFile, ecsConfYaml, os.FileMode(0o644))

	if err != nil {
		return fmt.Errorf("failed to write ecspresso config: %w", err)
	}

	return nil
}

func findTaskIdFromLog(log string) string {
	r := regexp.MustCompile(`(?s)Waiting for task ID (\S+)`)
	m := r.FindStringSubmatch(log)

	if len(m) < 2 {
		return ""
	}

	return m[1]
}
