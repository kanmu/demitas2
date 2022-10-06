package subcmd

import (
	"fmt"
	"os"

	"github.com/kanmu/demitas2"
)

type ProfilesCmd struct {
}

func (cmd *ProfilesCmd) Run(ctx *demitas2.Context) error {
	fmt.Printf("# conf-dir: %s\n", ctx.DefinitionOpts.ConfDir)
	files, err := os.ReadDir(ctx.DefinitionOpts.ExpandConfDir())

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
