package definition

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/kanmu/demitas2/utils"
	tilde "gopkg.in/mattes/go-expand-tilde.v1"
)

type DefinitionOpts struct {
	ConfDir            string `env:"DMTS_CONF_DIR" short:"d" required:"" default:"~/.demitas" help:"Config file base dir."`
	Config             string `env:"ECSPRESSO_CONF" required:"" default:"ecspresso.yml" help:"ecspresso config file name."`
	ContainerDef       string `env:"DMTS_CONT_DEF" required:"" default:"ecs-container-def.jsonnet" help:"ECS container definition file name."`
	ConfigOverrides    string `short:"e" help:"JSON/YAML string that overrides ecspresso config."`
	ServiceOverrides   string `short:"s" help:"JSON/YAML string that overrides ECS service definition."`
	TaskOverrides      string `short:"t" help:"JSON/YAML string that overrides ECS task definition."`
	ContainerOverrides string `short:"c" help:"JSON/YAML string that overrides ECS container definition."`
	Cluster            string `env:"DMTS_CLUSTER" help:"ECS cluster name."`
}

type Definition struct {
	EcspressoConfig *EcspressoConfig
	Service         *ServiceDefinition
	Task            *TaskDefinition
	Cluster         string
}

func (opts *DefinitionOpts) ExpandConfDir() string {
	confDir, err := tilde.Expand(opts.ConfDir)

	if err != nil {
		panic(err)
	}

	return confDir
}

func (opts *DefinitionOpts) Load(profile string, command string, image string, cpu uint64, memory uint64) (*Definition, error) {
	confDir := opts.ExpandConfDir()

	if profile != "" {
		confDir = filepath.Join(confDir, profile)
	}

	ecspressoConf, err := loadEcsecspressoConf(confDir, opts)

	if err != nil {
		return nil, err
	}

	serviceDefFile, err := ecspressoConf.get("service_definition")

	if err != nil {
		return nil, err
	}

	if serviceDefFile == "" {
		// NOTE: For compatibility
		serviceDefFile = "ecs-service-def.jsonnet"
	}

	taskDefFile, err := ecspressoConf.get("task_definition")

	if err != nil {
		return nil, err
	}

	if taskDefFile == "" {
		// NOTE: For compatibility
		taskDefFile = "ecs-task-def.jsonnet"
	}

	serviceDef, err := loadServiceDef(confDir, serviceDefFile, opts)

	if err != nil {
		return nil, err
	}

	containerDef, err := loadContainerDef(confDir, taskDefFile, opts, command, image)

	if err != nil {
		return nil, err
	}

	taskDef, err := loadTaskDef(confDir, taskDefFile, containerDef, opts, cpu, memory)

	if err != nil {
		return nil, err
	}

	cluster, err := ecspressoConf.get("cluster")

	if err != nil {
		return nil, err
	}

	return &Definition{
		EcspressoConfig: ecspressoConf,
		Service:         serviceDef,
		Task:            taskDef,
		Cluster:         cluster,
	}, nil
}

func loadEcsecspressoConf(confDir string, opts *DefinitionOpts) (*EcspressoConfig, error) {
	ecspressoConf, err := newEcspressoConfig(filepath.Join(confDir, opts.Config))

	if err != nil {
		return nil, err
	}

	if opts.Cluster != "" {
		js, err := json.Marshal(map[string]string{
			"cluster": opts.Cluster,
		})

		if err != nil {
			panic(err)
		}

		err = ecspressoConf.patch(string(js))

		if err != nil {
			return nil, err
		}
	}

	err = ecspressoConf.patch(opts.ConfigOverrides)

	if err != nil {
		return nil, err
	}

	return ecspressoConf, nil
}

func loadServiceDef(confDir string, serviceDefFile string, opts *DefinitionOpts) (*ServiceDefinition, error) {
	serviceDef, err := newServiceDefinition(filepath.Join(confDir, serviceDefFile))

	if err != nil {
		return nil, err
	}

	err = serviceDef.patch(opts.ServiceOverrides)

	if err != nil {
		return nil, err
	}

	return serviceDef, nil
}

func loadTaskDef(confDir string, taskDefFile string, containerDef *ContainerDefinition, opts *DefinitionOpts, cpu uint64, memory uint64) (*TaskDefinition, error) {
	taskDef, err := newTaskDefinition(filepath.Join(confDir, taskDefFile))

	if err != nil {
		return nil, err
	}

	err = taskDef.patch(opts.TaskOverrides, containerDef, cpu, memory)

	if err != nil {
		return nil, err
	}

	return taskDef, nil
}

func loadContainerDef(confDir string, taskDefFile string, opts *DefinitionOpts, command string, image string) (*ContainerDefinition, error) {
	containerDef, err := newContainerDefinition(filepath.Join(confDir, opts.ContainerDef), filepath.Join(confDir, taskDefFile))

	if err != nil {
		return nil, err
	}

	err = containerDef.patch(opts.ConfigOverrides, command, image)

	if err != nil {
		return nil, err
	}

	return containerDef, nil
}

func (def *Definition) Print() {
	ecspressoConf, err := utils.JSONToYAML(def.EcspressoConfig.Content)

	if err != nil {
		panic(err)
	}

	fmt.Printf(`# ECS task definition
%s
# ECS service definition
%s

# ECS task definition
%s
`,
		ecspressoConf,
		utils.PrettyJSON(def.Service.Content),
		utils.PrettyJSON(def.Task.Content),
	)
}
