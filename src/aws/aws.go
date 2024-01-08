package aws

import (
	"fmt"
	"landtitle/util"

	oAWS "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	oSession "github.com/aws/aws-sdk-go/aws/session"
)

const (
	DEF_REGION  string = "us-west-1"
	DEF_PROFILE        = "default"
)

type Session oSession.Session

// TODO will need to expand the shared credentials capabilities, good for now
func NewSession(pRegion, pProfile *string) (*Session, error) {
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
	sess, err := oSession.NewSession(config)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, fmt.Errorf(
			"AWS SDK returned a nil pointer when calling NewSession, can't continue.",
		)
	}
	temp_sess := Session(*sess)
	return &temp_sess, nil
}
