package models

import (
	"fmt"
)

type User struct {
	UUID              string   `json:"UUID"`              //
	Username          string   `json:"username"`          //
	Password          string   `json:"password"`          //
	Email             string   `json:"email"`             //
	Fullname          string   `json:"fullname"`          //
	Company           string   `json:"company"`           //
	Location          string   `json:"location"`          //
	Mobile            string   `json:"mobile"`            //
	URL               string   `json:"url"`               //
	Gravatar          string   `json:"gravatar"`          //
	Created           int64    `json:"created"`           //
	Updated           int64    `json:"updated"`           //
	Repositories      []string `json:"repositories"`      //
	Organizations     []string `json:"organizations"`     // Owner's Organizations
	Teams             []string `json:"teams"`             //
	Starts            []string `json:"starts"`            //
	Comments          []string `json:"comments"`          //
	Memo              []string `json:"memo"`              //
	JoinOrganizations []string `json:"joinorganizations"` // Join's Organizations
	JoinTeams         []string `json:"jointeams"`         //
	//RepositoryObjects []Repository `json:"repositoryobjects"` //
}

func (user *User) Get(username, password string) error {
	// TBD: codes as below just for temporary test
	if username != "mabin123" {
		return fmt.Errorf("User is not exist: %s", username)
	}
	if password != "123456" {
		return fmt.Errorf("Password is not match: %s", password)
	}
	return nil
	/*
		if exist, UUID, err := user.Has(username); err != nil {
			return err
		} else if exist == false && err == nil {
			return fmt.Errorf("User is not exist: %s", username)
		} else if exist == true && err == nil {
			if err := Get(user, UUID); err != nil {
				return err
			} else {
				if user.Password != password {
					return fmt.Errorf("User password error.")
				} else {
					return nil
				}
			}
		}
		return nil
	*/
}
