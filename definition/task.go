package definition

import (
	"bytes"
	"fmt"
	"os/user"
	"regexp"
	"strconv"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/kanmu/demitas2/utils"
	"github.com/valyala/fastjson"
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

func (taskDef *TaskDefinition) patch(overrides string, containerDef *ContainerDefinition, cpu uint64, memory uint64) error {
	overrides = strings.TrimSpace(overrides)
	patchedContent := taskDef.Content
	var err error

	if overrides != "" {
		patchedContent, err = jsonpatch.MergePatch(patchedContent, []byte(overrides))

		if err != nil {
			return fmt.Errorf("failed to patch ECS task definition: %w", err)
		}
	}

	if containerDef != nil {
		containerDefinitions := fmt.Sprintf(`{"containerDefinitions":[%s]}`, string(containerDef.Content))
		patchedContent, err = jsonpatch.MergePatch(patchedContent, []byte(containerDefinitions))
	}

	if err != nil {
		return fmt.Errorf("failed to patch containerDefinitions: %w", err)
	}

	if cpu != 0 {
		patchedContent, err = jsonpatch.MergePatch(patchedContent, []byte(`{"cpu":"`+strconv.FormatUint(cpu, 10)+`"}`))

		if err != nil {
			return fmt.Errorf("failed to update 'cpu' in ECS taks definition: %w", err)
		}
	}

	if memory != 0 {
		patchedContent, err = jsonpatch.MergePatch(patchedContent, []byte(`{"memory":"`+strconv.FormatUint(memory, 10)+`"}`))

		if err != nil {
			return fmt.Errorf("failed to update 'memory' in ECS taks definition: %w", err)
		}
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

	username := regexp.MustCompile(`\W+`).ReplaceAllString(currUser.Username, "")
	patch := fmt.Sprintf(`{"family":"dmts-%s-%s"}`, username, string(family))
	patchedContent, err := jsonpatch.MergePatch(content, []byte(patch))

	if err != nil {
		return nil, err
	}

	return patchedContent, nil
}
