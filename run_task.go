package demitas2

import (
	"fmt"
	"io/ioutil"
	"os"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/mattn/go-shellwords"
	"github.com/winebarrel/demitas2/definition"
	"github.com/winebarrel/demitas2/utils"
)

func RunTask(ecspressoPath string, ecspressoOpts string, def *definition.Definition) (stdout string, stderr string, err error) {
	runInTempDir(func() {
		err = writeTemporaryConfigs(def.EcspressoConfig, def.Service, def.Task)

		if err != nil {
			return
		}

		cmdWithArgs := []string{ecspressoPath, "run"}
		args, err := shellwords.Parse(ecspressoOpts)

		if err != nil {
			return
		}

		if len(args) > 0 {
			cmdWithArgs = append(cmdWithArgs, args...)
		}

		stdout, stderr, err = utils.RunCommand(cmdWithArgs, false)

		if err != nil {
			return
		}
	})

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
	err := ioutil.WriteFile("task-def.json", taskDef.Content, os.FileMode(0o644))

	if err != nil {
		return fmt.Errorf("failed to write ECS task definition: %w", err)
	}

	err = ioutil.WriteFile("service-def.json", svrDef.Content, os.FileMode(0o644))

	if err != nil {
		return fmt.Errorf("failed to write ECS service definition: %w", err)
	}

	ecsConfOverrides := `{"service_definition":"service-def.json","task_definition":"task-def.json"}`
	ecsConfJson, err := jsonpatch.MergePatch(ecsConf.Content, []byte(ecsConfOverrides))

	if err != nil {
		return fmt.Errorf("failed to update temporary ecspresso config: %w", err)
	}

	ecsConfYaml, err := utils.JSONToYAML(ecsConfJson)

	if err != nil {
		return fmt.Errorf("failed to convert temporary ecspresso config to yaml: %w", err)
	}

	err = ioutil.WriteFile("ecspresso.yml", ecsConfYaml, os.FileMode(0o644))

	if err != nil {
		return fmt.Errorf("failed to write ecspresso config: %w", err)
	}

	return nil
}
