//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"time"
)

type Log struct {
	ID        int32 `sql:"primary_key"`
	UserID    int32
	Message   string
	CreatedAt time.Time
}
