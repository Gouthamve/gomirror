package util

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/ChimeraCoder/anaconda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/s3"
)

// APIKEY is the APIKEY
var APIKEY = "tPvbiL1SICgfYsa6HXlhJ5Bjr"

// APISEC is the APISEC
var APISEC = "ub9M8aRQ8lFeznUDfSZG8JbiyoYpQC7zVwqAzKN8KT8t7SNWWn"

// ATOKEN is the Access Token
var ATOKEN = "2380003249-6poY4mbdHWl494gUyw5RKixwVoJLeTGxClI3xHS"

// ASEC is the Access Secret
var ASEC = "ZIMsjJA3XuHzGnM8R5M65NQrZ9ObjYtui0rmgdpheXbuV"

// TwitterStream streams twitter
func TwitterStream(hash string) {
	anaconda.SetConsumerKey(APIKEY)
	anaconda.SetConsumerSecret(APISEC)

	client := anaconda.NewTwitterApi(ATOKEN, ASEC)

	v := url.Values{}
	v.Set("track", hash)
	s := client.PublicStreamFilter(v)

	for {
		item := <-s.C
		switch status := item.(type) {
		case anaconda.Tweet:
			go addPersontoDb(status.User.ScreenName, status.Entities.Media[0].Media_url)
		default:
			fmt.Println("YOLO")
		}
	}
}

func addPersontoDb(handle, url string) {
	filename := handle + ".jpg"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("OOOOPS:", handle, url)
		return
	}
	defer resp.Body.Close()
	file, err := os.Create(path.Join(os.TempDir(), "com.nhack", filename))
	if err != nil {
		log.Fatal(err)
	}
	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("OOOOPS2:", handle, url)
		return
	}

	s3vc := s3.New(sess, aws.NewConfig().WithRegion("us-west-2"))
	params := &s3.PutObjectInput{
		Bucket: aws.String("nhack1"), // Required
		Key:    aws.String(filename),
		Body:   file,
	}

	_, err = s3vc.PutObject(params)
	if err != nil {
		fmt.Println("here")
		return
	}

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
		ExternalImageId: aws.String(handle),
	}

	resp3, err := svc.IndexFaces(params2)
	if err != nil {
		return
	}

	fmt.Println(resp3)
}
