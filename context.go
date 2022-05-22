package demitas2

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/winebarrel/demitas2/definition"
)

type Context struct {
	EcspressoCmd   string
	EcspressoOpts  string
	DryRun         bool
	DefinitionOpts *definition.DefinitionOpts
	AwsConfig      aws.Config
}
