package model

import (
	"encoding/json"
	"errors"
	"io"
)

// Follow represents a follow relationship between two users
type Follow struct {
	ProfileID  uint32 `json:"profileId"`
	FollowerID uint32 `json:"followerId"`
}

// NewFollow creates a new Follow instance with validation
func NewFollow(profileID, followerID uint32) (*Follow, error) {
	if profileID == 0 {
		return nil, errors.New("invalid profileid")
	}
	if followerID == 0 {
		return nil, errors.New("invalid followerid")
	}
	return &Follow{ProfileID: profileID, FollowerID: followerID}, nil
}

type Follows []*Follow

func (o *Follows) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

// sa naseg psw projekta
// Equal checks if two Follow instances are equal
func (f *Follow) Equal(other *Follow) bool {
	return f.ProfileID == other.ProfileID && f.FollowerID == other.FollowerID
}

func (o *Follow) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
