package dto

import (
	"encoding/json"
	"io"
)

type BooleanDto struct {
	BoolField bool `json:"boolField"`
}

func (o *BooleanDto) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}
