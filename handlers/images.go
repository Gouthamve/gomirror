package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo"
)

var sess *session.Session

func init() {
	var err error
	sess, err = session.NewSession()
	if err != nil {
		panic(err)
	}
}

// IndexFace detects the face
func IndexFace(c echo.Context) error {
	image, err := c.FormFile("image")
	if err != nil {
		return err
	}

	src, err := image.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	filename := time.Now().String() + image.Filename

	s3vc := s3.New(sess, aws.NewConfig().WithRegion("us-west-2"))
	params := &s3.PutObjectInput{
		Bucket: aws.String("nhack1"), // Required
		Key:    aws.String(filename),
		Body:   src,
	}

	resp, err := s3vc.PutObject(params)
	if err != nil {
		fmt.Println("here")
		return err
	}

	fmt.Println(resp)

	svc := rekognition.New(sess, aws.NewConfig().WithRegion("us-west-2"))

	params2 := &rekognition.IndexFacesInput{
		CollectionId: aws.String("nhack1"), // Required
		Image: &rekognition.Image{ // Required
			S3Object: &rekognition.S3Object{
				Bucket: aws.String("nhack1"),
				Name:   aws.String(filename),
			},
		},
		DetectionAttributes: []*string{
			aws.String("DEFAULT"),
		},
		ExternalImageId: aws.String(filename),
	}

	resp2, err := svc.IndexFaces(params2)
	if err != nil {
		return err
	}

	fmt.Println(resp2)
	return c.String(http.StatusOK, "Hello, World!\n")
}

// DetectFace detects the possible face
func DetectFace(c echo.Context) error {
	image, err := c.FormFile("image")
	if err != nil {
		return err
	}

	src, err := image.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	filename := time.Now().String() + image.Filename

	s3vc := s3.New(sess, aws.NewConfig().WithRegion("us-west-2"))
	params := &s3.PutObjectInput{
		Bucket: aws.String("nhack1"), // Required
		Key:    aws.String(filename),
		Body:   src,
	}

	resp, err := s3vc.PutObject(params)
	if err != nil {
		return err
	}

	fmt.Println(resp)

	svc := rekognition.New(sess, aws.NewConfig().WithRegion("us-west-2"))
	params2 := &rekognition.SearchFacesByImageInput{
		CollectionId: aws.String("nhack1"), // Required
		Image: &rekognition.Image{ // Required
			S3Object: &rekognition.S3Object{
				Bucket: aws.String("nhack1"),
				Name:   aws.String(filename),
			},
		},
		FaceMatchThreshold: aws.Float64(1.0),
		MaxFaces:           aws.Int64(1),
	}
	resp2, err := svc.SearchFacesByImage(params2)
	if err != nil {
		return err
	}

	fmt.Println(resp2)
	return c.String(http.StatusOK, "Hello, World!\n")
}
