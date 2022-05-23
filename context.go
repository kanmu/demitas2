package demitas2

import (
	"github.com/winebarrel/demitas2/definition"
	"github.com/winebarrel/demitas2/ecscli"
	"github.com/winebarrel/demitas2/ecspresso"
)

type Context struct {
	Ecspresso      *ecspresso.Ecspresso
	DryRun         bool
	DefinitionOpts *definition.DefinitionOpts
	Ecs            *ecscli.Driver
}
