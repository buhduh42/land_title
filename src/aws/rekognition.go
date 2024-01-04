package aws

import (
	"fmt"

	oSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/rekognition/rekognitioniface"
)

type RekognitionClient interface {
	AnalyzeFromDisk(string) error
	AnalyzeFromS3(string, string) error
}

type rekognitionClient struct {
	client rekognitioniface.RekognitionAPI
}

// this should ONLY be called from main()
func NewRekognitionClient(session *Session) (RekognitionClient, error) {
	if session == nil {
		return nil, fmt.Errorf(
			"Passed session to aws.NewRekognitionClient can't be nil.",
		)
	}
	tSess := oSession.Session((*session))
	rClient := rekognition.New(&tSess)
	return BuildRekognitionClient(rClient), nil
}

// convenience function, mostly for dependency injection
// while testing
func BuildRekognitionClient(client rekognitioniface.RekognitionAPI) RekognitionClient {
	return &rekognitionClient{
		client: client,
	}
}

func (r *rekognitionClient) AnalyzeFromDisk(path string) error {
	return nil
}

func (r *rekognitionClient) AnalyzeFromS3(bucket, key string) error {
	return nil
}
