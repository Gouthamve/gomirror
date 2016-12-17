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
			if len(status.Entities.Media) == 0 {
				fmt.Println("Tweet w/o image detected")
				continue
			}
			go addPersontoDb(status.User, status.Entities.Media[0].Media_url, client)
		default:
			fmt.Println("YOLO")
		}
	}
}

func addPersontoDb(u anaconda.User, urly string, cl *anaconda.TwitterApi) {
	handle := u.ScreenName
	fmt.Println(handle, urly)
	user, err := GetUser(handle)
	if err != nil {
		fmt.Println(err)
		return
	}
	if user.TwitterID != "" && user.InDB {
		return
	}

	filename := handle + ".jpg"
	resp, err := http.Get(urly)
	if err != nil {
		fmt.Println("OOOOPS:", handle, urly)
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
		fmt.Println("OOOOPS2:", handle, urly)
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
	user = UserModel{
		TwitterID: handle,
		Name:      u.Name,
		InDB:      true,
	}

	if len(resp3.FaceRecords) == 0 {
		user.InDB = false
	}

	v := url.Values{}
	v.Set("user_id", u.IdStr)
	v.Set("count", "10")
	twts, err := cl.GetUserTimeline(v)
	if err != nil {
		fmt.Println(err)
		return
	}

	tweetIds := make([]string, len(twts))
	for i := 0; i < len(twts); i++ {
		tweetIds[i] = twts[i].IdStr
	}

	user.Tweets = tweetIds
	if err := SaveUser(user); err != nil {
		fmt.Println(err)
	}
}
