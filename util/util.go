package util

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
)

// CreateCollection creates a collection
func CreateCollection(cid string) {
	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}

	svc := rekognition.New(sess, aws.NewConfig().WithRegion("us-west-2"))
	params := &rekognition.CreateCollectionInput{
		CollectionId: aws.String(cid), // Required
	}
	resp, err := svc.CreateCollection(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)
}

// DeleteCollection creates a collection
func DeleteCollection(cid string) {
	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}

	svc := rekognition.New(sess, aws.NewConfig().WithRegion("us-west-2"))
	params := &rekognition.DeleteCollectionInput{
		CollectionId: aws.String(cid), // Required
	}
	resp, err := svc.DeleteCollection(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)
}
