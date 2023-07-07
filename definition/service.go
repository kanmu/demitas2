package definition

import (
	"fmt"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/kanmu/demitas2/utils"
)

type ServiceDefinition struct {
	Content []byte
}

func newServiceDefinition(path string) (*ServiceDefinition, error) {
	content, err := utils.ReadJSONorJsonnet(path)

	if err != nil {
		return nil, fmt.Errorf("failed to load ECS service definition: %w: %s", err, path)
	}

	svrDef := &ServiceDefinition{
		Content: content,
	}

	return svrDef, nil
}

func (svrDef *ServiceDefinition) patch(overrides string) error {
	overrides = strings.TrimSpace(overrides)

	if overrides == "" {
		return nil
	}

	patchedContent, err := jsonpatch.MergePatch(svrDef.Content, []byte(overrides))

	if err != nil {
		return fmt.Errorf("failed to patch ECS service definition: %w", err)
	}

	patchedContent, err = jsonpatch.MergePatch(patchedContent, []byte(`{"enableExecuteCommand": true}`))

	if err != nil {
		return fmt.Errorf("failed to enable ECS Exec: %w", err)
	}

	svrDef.Content = patchedContent

	return nil
}

func (svrDef *ServiceDefinition) Print() {
	fmt.Printf("# ECS service definition\n%s\n", utils.PrettyJSON(svrDef.Content))
}
