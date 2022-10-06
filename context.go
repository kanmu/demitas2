package demitas2

import (
	"github.com/kanmu/demitas2/definition"
	"github.com/kanmu/demitas2/ecscli"
	"github.com/kanmu/demitas2/ecspresso"
)

type Context struct {
	Ecspresso      *ecspresso.Ecspresso
	DryRun         bool
	DefinitionOpts *definition.DefinitionOpts
	Ecs            *ecscli.Driver
}
