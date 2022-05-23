package definition

import (
	"fmt"
	"io/ioutil"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/valyala/fastjson"
	"github.com/winebarrel/demitas2/utils"
)

type EcspressoConfig struct {
	Content []byte
}

func newEcspressoConfig(path string) (*EcspressoConfig, error) {
	content, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("failed to load ecspresso config: %w: %s", err, path)
	}

	js, err := utils.YAMLToJSON(content)

	if err != nil {
		return nil, fmt.Errorf("failed to parse ecspresso config: %w: %s", err, path)
	}

	ecsConf := &EcspressoConfig{
		Content: js,
	}

	return ecsConf, nil
}

func (ecsConf *EcspressoConfig) patch(overrides string) error {
	overrides = strings.TrimSpace(overrides)

	if overrides == "" {
		return nil
	}

	patchedContent, err := jsonpatch.MergePatch(ecsConf.Content, []byte(overrides))

	if err != nil {
		return fmt.Errorf("failed to patch ecspresso config: %w", err)
	}

	ecsConf.Content = patchedContent

	return nil
}

func (ecsConf *EcspressoConfig) Print() {
	ym, _ := utils.JSONToYAML(ecsConf.Content)
	fmt.Printf("# ecspresso config\n%s\n", strings.TrimSpace(string(ym)))
}

func (ecsConf *EcspressoConfig) get(key string) (string, error) {
	var p fastjson.Parser
	v, err := p.ParseBytes(ecsConf.Content)

	if err != nil {
		return "", fmt.Errorf("failed to get '%s' from ecspresso config: %w", key, err)
	}

	bs := v.GetStringBytes(key)

	if bs != nil {
		return string(bs), nil
	} else {
		return "", nil
	}
}
