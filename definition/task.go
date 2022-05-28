package definition

import (
	"bytes"
	"fmt"
	"os/user"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/valyala/fastjson"
	"github.com/winebarrel/demitas2/utils"
)

type TaskDefinition struct {
	Content []byte
}

func newTaskDefinition(path string) (*TaskDefinition, error) {
	content, err := utils.ReadJSONorJsonnet(path)

	if err != nil {
		return nil, fmt.Errorf("failed to load ECS task definition: %w: %s", err, path)
	}

	patchedContent, err := patchContainerDefInLoad(content)

	if err != nil {
		return nil, fmt.Errorf("failed to patch ECS container definition in load: %w: %s", err, path)
	}

	taskDef := &TaskDefinition{
		Content: patchedContent,
	}

	return taskDef, nil
}

func (taskDef *TaskDefinition) patch(overrides string, containerDef *ContainerDefinition) error {
	overrides = strings.TrimSpace(overrides)
	patchedContent := taskDef.Content
	var err error

	if overrides != "" {
		patchedContent, err = jsonpatch.MergePatch(patchedContent, []byte(overrides))

		if err != nil {
			return fmt.Errorf("failed to patch ECS task definition: %w", err)
		}
	}

	containerDefinitions := fmt.Sprintf(`{"containerDefinitions":[%s]}`, string(containerDef.Content))
	patchedContent, err = jsonpatch.MergePatch(patchedContent, []byte(containerDefinitions))

	if err != nil {
		return fmt.Errorf("failed to patch containerDefinitions: %w", err)
	}

	taskDef.Content = patchedContent

	return nil
}

func patchContainerDefInLoad(content []byte) ([]byte, error) {
	var p fastjson.Parser
	v, err := p.ParseBytes(content)

	if err != nil {
		return nil, err
	}

	family := v.GetStringBytes("family")

	if family == nil {
		return nil, fmt.Errorf("'family' not found in task definition")
	}

	if bytes.HasPrefix(family, []byte("dmts-")) {
		return content, nil
	}

	currUser, err := user.Current()

	if err != nil {
		panic(err)
	}

	patch := fmt.Sprintf(`{"family":"dmts-%s-%s"}`, currUser.Username, string(family))
	patchedContent, err := jsonpatch.MergePatch(content, []byte(patch))

	if err != nil {
		return nil, err
	}

	return patchedContent, nil
}
