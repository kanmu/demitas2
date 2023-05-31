package definition

import (
	"fmt"
	"os"

	"github.com/kanmu/demitas2/utils"
	"github.com/valyala/fastjson"
)

type Overrides struct {
	Content []byte
}

func newOoverrides(path string) (*Overrides, error) {
	_, err := os.Stat(path)

	if err != nil {
		return &Overrides{}, nil
	}

	content, err := utils.EvaluateJsonnet(path)

	if err != nil {
		return nil, fmt.Errorf("failed to parse overrides file: %w: %s", err, path)
	}

	overrides := &Overrides{
		Content: content,
	}

	return overrides, nil
}

func (overrides *Overrides) get(key string) string {
	var p fastjson.Parser
	content, _ := p.ParseBytes(overrides.Content)
	v := content.Get(key)

	if v != nil {
		return v.String()
	} else {
		return ""
	}
}
