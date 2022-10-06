# demitas2

[![build](https://github.com/kanmu/demitas2/actions/workflows/build.yml/badge.svg)](https://github.com/kanmu/demitas2/actions/workflows/build.yml)

Wrapper for ecspresso that creates task definitions at run time.

## Requirement

* [ecspresso](https://github.com/kayac/ecspresso)
* [AWS CLI](https://aws.amazon.com/cli/)
* [Session Manager plugin](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html)

## Installation

```
brew install kanmu/demitas2/demitas2
```

## Usage

```
Usage: dmts --ecspresso-cmd="ecspresso" --conf-dir="~/.demitas" --config="ecspresso.yml" --container-def="ecs-container-def.jsonnet" <command>

Flags:
  -h, --help                       Show context-sensitive help.
      --version
      --ecspresso-cmd="ecspresso"
                                   ecspresso command path ($ECSPRESSO_CMD).
  -X, --ecspresso-opts=STRING      Options passed to ecspresso
                                   ($ECSPRESSO_OPTS).
      --dry-run                    Run ecspresso with dry-run.
  -P, --aws-profile=STRING         AWS profile name ($AWS_PROFILE)
  -d, --conf-dir="~/.demitas"      Config file base dir ($DMTS_CONF_DIR).
      --config="ecspresso.yml"     ecspresso config file name ($ECSPRESSO_CONF).
      --container-def="ecs-container-def.jsonnet"
                                   ECS container definition file name
                                   ($DMTS_CONT_DEF).
  -e, --config-overrides=STRING    JSON/YAML string that overrides ecspresso
                                   config.
  -s, --service-overrides=STRING
                                   JSON/YAML string that overrides ECS service
                                   definition.
  -t, --task-overrides=STRING      JSON/YAML string that overrides ECS task
                                   definition.
  -c, --container-overrides=STRING
                                   JSON/YAML string that overrides ECS container
                                   definition.

Commands:
  run --ecspresso-cmd="ecspresso" --conf-dir="~/.demitas" --config="ecspresso.yml" --container-def="ecs-container-def.jsonnet"
    Run ECS task.

  exec --ecspresso-cmd="ecspresso" --conf-dir="~/.demitas" --config="ecspresso.yml" --container-def="ecs-container-def.jsonnet" --command="bash"
    Run ECS task and execute a command on a container.

  port-forward --ecspresso-cmd="ecspresso" --conf-dir="~/.demitas" --config="ecspresso.yml" --container-def="ecs-container-def.jsonnet" --remote-host=STRING --remote-port=UINT --local-port=UINT
    Forward a local port to a container.

  profiles --ecspresso-cmd="ecspresso" --conf-dir="~/.demitas" --config="ecspresso.yml" --container-def="ecs-container-def.jsonnet"
    List profiles.

  install-completions --ecspresso-cmd="ecspresso" --conf-dir="~/.demitas" --config="ecspresso.yml" --container-def="ecs-container-def.jsonnet"
    Install shell completions

Run "dmts <command> --help" for more information on a command.
```

## Install shell completions

```
dmts install-completions >> ~/.zshrc
```
