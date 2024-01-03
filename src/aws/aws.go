package aws

import (
	"land_title/util"

	"github.com/aws/aws-sdk-go/aws/credentials"
	oSession "github.com/aws/aws-sdk-go/aws/session"
)

const (
	DEF_REGION  string = "us-west-1"
	DEF_PROFILE        = "default"
)

type Session oSession.Session

// TODO will need to expand the shared credentials capabilities, good for now
func NewSession(pRegion, pProfile *string) *Session {
	var region, profile string
	if pRegion == nil {
		region = DEF_REGION
	}
	if pProfile == nil {
		profile = DEF_PROFILE
	}
	creds := credentials.NewSharedCredentials("", profile)
	config := oAWS.NewConfig().WithCredentials(creds)
	config.MergeIn(
		&oAWS.Config{
			Region: util.Ptr(region),
		},
	)
	sesion := oSession.NewSession(config)
	return session(*Session)
}
