package models

import (
	"github.com/containerops/wrench/db"
)

func init() {
	db.RegisterModel(new(Repository), new(Tag), new(Image))
}
