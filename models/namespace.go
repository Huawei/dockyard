package models

import ()

type Namespace struct {
	Id        int64  `json:"id" orm:"auto"`
	Namespace string `json:"namespace" orm:"unique;varchar(255)"`
	Type      string `json:"type" orm:"varchar(255)"`
}

func (n *Namespace) Get(namespace string) error {
	return nil
}

func (n *Namespace) Save(namespace string) error {
	return nil
}
