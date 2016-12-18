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
		Name:           u.Name,
		TwitterID:      u.TwitterID,
		Tweets:         u.Tweets,
		Emotion:        *fm2.Emotions[0].Type,
		EmotionConf:    *fm2.Emotions[0].Confidence,
		Gender:         *fm2.Gender.Value,
		GenderConf:     *fm2.Gender.Confidence,
		Eyeglasses:     *fm2.Eyeglasses.Value,
		EyeglassesConf: *fm2.Eyeglasses.Confidence,
		EyesOpen:       *fm2.EyesOpen.Value,
		EyesOpenConf:   *fm2.EyesOpen.Confidence,
		MouthOpen:      *fm2.MouthOpen.Value,
		MouthOpenConf:  *fm2.MouthOpen.Confidence,
		Mustache:       *fm2.Mustache.Value,
		MustacheConf:   *fm2.Mustache.Confidence,
		Smile:          *fm2.Smile.Value,
		SmileConf:      *fm2.Smile.Confidence,
		Sunglasses:     *fm2.Sunglasses.Value,
		SunglassesConf: *fm2.Sunglasses.Confidence,
		Beard:          *fm2.Beard.Value,
		BeardConf:      *fm2.Beard.Confidence,
	}
	return c.JSON(http.StatusOK, r)
}

type responsey struct {
	Name           string   `json:"name"`
	NameConf       string   `json:"nameConf"`
	TwitterID      string   `json:"twitterId"`
	TwitterIDConf  float64  `json:"twitterIdConf"`
	Tweets         []string `json:"tweets"`
	Emotion        string   `json:"emotion"`
	EmotionConf    float64  `json:"emotionConf"`
	Gender         string   `json:"gender"`
	GenderConf     float64  `json:"genderConf"`
	Age            int      `json:"age"`
	AgeConf        float64  `json:"ageConf"`
	Eyeglasses     bool     `json:"eyeglasses"`
	EyeglassesConf float64  `json:"eyeglassesConf"`
	EyesOpen       bool     `json:"eyesopen"`
	EyesOpenConf   float64  `json:"eyesopenConf"`
	MouthOpen      bool     `json:"mouthopen"`
	MouthOpenConf  float64  `json:"mouthopenConf"`
	Mustache       bool     `json:"mustache"`
	MustacheConf   float64  `json:"mustacheConf"`
	Smile          bool     `json:"smile"`
	SmileConf      float64  `json:"smileConf"`
	Sunglasses     bool     `json:"sunglasses"`
	SunglassesConf float64  `json:"sunglassesConf"`
	Beard          bool     `json:"beard"`
	BeardConf      float64  `json:"beardConf"`
}
