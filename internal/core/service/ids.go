package service

import (
	"github.com/google/uuid"
)

func defaultInstanceID() string {
	return uuid.NewString()
}
