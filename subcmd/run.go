package subcmd

import (
	"fmt"

	"github.com/winebarrel/demitas2"
	"github.com/winebarrel/demitas2/definition"
)

type RunCmd struct {
	Command string `help:"Command to run on a container."`
	Image   string `help:"Container image."`
}

func (cmd *RunCmd) Run(ctx *demitas2.Context) error {
	def, err := definition.Load(ctx.DefinitionOpts, cmd.Command, cmd.Image)

	if err != nil {
		return err
	}

	if ctx.DryRun {
		def.Print()
		fmt.Println()
	}

	_, _, err = demitas2.RunTask(ctx.EcspressoCmd, ctx.EcspressoOpts, def)

	return err
}
