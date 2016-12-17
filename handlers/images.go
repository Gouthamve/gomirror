package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gouthamve/gomirror/util"
	"github.com/labstack/echo"
)

var sess *session.Session

var collectionID = "nhack1"

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
		CollectionId: aws.String(collectionID), // Required
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
		fmt.Println(err)
		return err
	}

	src, err := image.Open()
	if err != nil {
		fmt.Println(err)
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
		fmt.Println(err)
		return err
	}

	fmt.Println(resp)

	svc := rekognition.New(sess, aws.NewConfig().WithRegion("us-west-2"))
	params2 := &rekognition.SearchFacesByImageInput{
		CollectionId: aws.String(collectionID), // Required
		Image: &rekognition.Image{ // Required
			S3Object: &rekognition.S3Object{
				Bucket: aws.String("nhack1"),
				Name:   aws.String(filename),
			},
		},
		FaceMatchThreshold: aws.Float64(1.0),
		MaxFaces:           aws.Int64(1),
	}
	faces, err := svc.SearchFacesByImage(params2)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fm := faces.FaceMatches
	if len(fm) == 0 {
		return c.String(http.StatusNotFound, "No image in database matches\n")
	}

	twittrID := fm[0].Face.ExternalImageId
	u, err := util.GetUser(*twittrID)
	if err != nil {
		return err
	}

	if u.TwitterID == "" {
		return c.String(http.StatusNotFound, "User in AWS but not in local")
	}

	params3 := &rekognition.DetectFacesInput{
		Image: &rekognition.Image{ // Required
			S3Object: &rekognition.S3Object{
				Bucket: aws.String("nhack1"),
				Name:   aws.String(filename),
			},
		},
		Attributes: []*string{
			aws.String("ALL"),
		},
	}
	faces3, err := svc.DetectFaces(params3)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fm2 := faces3.FaceDetails[0]

	r := responsey{
		Name:       u.Name,
		TwitterID:  u.TwitterID,
		Tweets:     u.Tweets,
		Emotion:    *fm2.Emotions[0].Type,
		Gender:     *fm2.Gender.Value,
		Eyeglasses: *fm2.Eyeglasses.Value,
		EyesOpen:   *fm2.EyesOpen.Value,
		MouthOpen:  *fm2.MouthOpen.Value,
		Mustache:   *fm2.Mustache.Value,
		Smile:      *fm2.Smile.Value,
		Sunglasses: *fm2.Sunglasses.Value,
		Beard:      *fm2.Beard.Value,
	}
	return c.JSON(http.StatusOK, r)
}

type responsey struct {
	Name       string   `json:"name"`
	TwitterID  string   `json:"twitterId"`
	Tweets     []string `json:"tweets"`
	Emotion    string   `json:"emotion"`
	Gender     string   `json:"gender"`
	Age        int      `json:"age"`
	Eyeglasses bool     `json:"eyeglasses"`
	EyesOpen   bool     `json:"eyesopen"`
	MouthOpen  bool     `json:"mouthopen"`
	Mustache   bool     `json:"mustache"`
	Smile      bool     `json:"smile"`
	Sunglasses bool     `json:"sunglasses"`
	Beard      bool     `json:"beard"`
}
