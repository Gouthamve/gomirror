package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

// UserModel is the user model
type UserModel struct {
	TwitterID string   `json:"twitterID"`
	Name      string   `json:"name"`
	InDB      bool     `json:"indb"`
	Tweets    []string `json:"tweets"`
}

var db *bolt.DB

func init() {
	var err error
	db, err = bolt.Open("tw.db", 0666, nil)
	if err != nil {
		log.Fatalln(err)
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("users"))
		if err != nil {
			fmt.Println(err)
			return err
		}

		return nil
	}); err != nil {
		log.Println(err)
		return
	}
}

// DBClose closes db
func DBClose() {
	db.Close()
}

// SaveUser saves user
func SaveUser(u UserModel) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return errors.New("Bucket does not exist")
		}

		val, err := json.Marshal(u)
		if err != nil {
			return err
		}

		err = b.Put([]byte(u.TwitterID), val)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// GetUser gets user
func GetUser(twitterID string) (UserModel, error) {
	var u UserModel

	if err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		if bucket == nil {
			return errors.New("Users bucket does not exist")
		}

		val := bucket.Get([]byte(twitterID))
		if val == nil {
			return nil
		}

		if err := json.Unmarshal(val, &u); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return UserModel{}, err
	}

	return u, nil
}
