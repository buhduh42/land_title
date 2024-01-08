package aws

/*
import (
	"fmt"

	oSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Client interface{}

type s3Client struct{}

// this should ONLY be called from main()
func NewS3Client(session *Session) (S3Client, error) {
	if session == nil {
		return nil, fmt.Errorf(
			"Passed session to aws.NewRekognitionClient can't be nil.",
		)
	}
	tSess := oSession.Session((*session))
	rClient := s3.New(&tSess)
	return BuildS3Client(rClient), nil
}

// convenience function, mostly for dependency injection
// while testing
func BuildS3Client(client s3iface.) S3Client {
	return &s3Client{
		client: client,
	}
}
*/
