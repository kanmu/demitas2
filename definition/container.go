package definition

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/kanmu/demitas2/utils"
	"github.com/mattn/go-shellwords"
	"github.com/valyala/fastjson"
)

type ContainerDefinition struct {
	Content []byte
}

func newContainerDefinition(path string, taskDefPath string) (*ContainerDefinition, error) {
	var content []byte
	var err error

	if _, err = os.Stat(path); err != nil {
		content, err = readContainerDefFromTaskDef(taskDefPath)

		if err != nil {
			return nil, fmt.Errorf("failed to load ECS task definition (instead of ECS container definition): %w: %s", err, taskDefPath)
		}
	} else {
		content, err = utils.ReadJSONorJsonnet(path)

		if err != nil {
			return nil, fmt.Errorf("failed to load ECS container definition: %w: %s", err, path)
		}
	}

	containerDef := &ContainerDefinition{
		Content: content,
	}

	return containerDef, nil
}

func (containerDef *ContainerDefinition) patch(overrides string, command string, image string) error {
	overrides = strings.TrimSpace(overrides)
	patchedContent0, err := jsonpatch.MergePatch(containerDef.Content, []byte(`{"logConfiguration":null}`))

	if err != nil {
		return fmt.Errorf("failed to patch ECS container definition: %w", err)
	}

	if command != "" {
		args, err := shellwords.Parse(command)

		if err != nil {
			return nil
		}

		js, err := json.Marshal(args)

		if err != nil {
			panic(err)
		}

		patchedContent0, err = jsonpatch.MergePatch(patchedContent0, []byte(`{"command":`+string(js)+`}`))

		if err != nil {
			return fmt.Errorf("failed to update 'command' in ECS container definition: %w", err)
		}
	}

	if image != "" {
		if strings.HasPrefix(image, ":") {
			var p fastjson.Parser
			v, err := p.ParseBytes(patchedContent0)

			if err != nil {
				panic(err)
			}

			origImg := string(v.GetStringBytes("image"))
			image = regexp.MustCompile(":[^:]+$").ReplaceAllString(origImg, image)
		}

		patchedContent0, err = jsonpatch.MergePatch(patchedContent0, []byte(`{"image":"`+image+`"}`))

		if err != nil {
			return fmt.Errorf("failed to update 'image' in ECS container definition: %w", err)
		}
	}

	var patchedContent []byte

	if overrides != "" {
		patchedContent, err = jsonpatch.MergePatch(patchedContent0, []byte(overrides))

		if err != nil {
			return fmt.Errorf("failed to patch ECS container definition: %w", err)
		}
	} else {
		patchedContent = patchedContent0
	}

	containerDef.Content = patchedContent

	return nil
}

func readContainerDefFromTaskDef(path string) ([]byte, error) {
	content, err := utils.ReadJSONorJsonnet(path)

	if err != nil {
		return nil, err
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(content)

	if err != nil {
		return nil, err
	}

	containerDef := v.GetObject("containerDefinitions", "0")

	// NOTE: Ignore dependsOn
	containerDef.Del("dependsOn")

	if containerDef == nil {
		return nil, fmt.Errorf("'containerDefinitions.0' is not found in ECS task definition: %s", path)
	}

	return containerDef.MarshalTo(nil), nil
}
