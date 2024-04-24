package model

import (
	"encoding/json"
	"io"
)

type Profile struct {
	ID             int64    `json:"id"`
	FirstName      string   `json:"firstName"`
	LastName       string   `json:"lastName"`
	ProfilePicture string   `json:"profilePicture"`
	UserID         int64    `json:"userID"`
	Follows        []Follow `json:"follows"`
}

type Profiles []*Profile

func (o *Profiles) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}
