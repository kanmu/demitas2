package subcmd

import (
	"fmt"
	"io/ioutil"

	"github.com/winebarrel/demitas2"
)

type ProfilesCmd struct {
}

func (cmd *ProfilesCmd) Run(ctx *demitas2.Context) error {
	fmt.Printf("# conf-dir: %s\n", ctx.DefinitionOpts.ConfDir)
	files, err := ioutil.ReadDir(ctx.DefinitionOpts.ExpandConfDir())

	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	for _, f := range files {
		if f.IsDir() {
			fmt.Println(f.Name())
		}
	}

	return nil
}
