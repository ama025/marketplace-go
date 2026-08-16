package entities

import "github.com/google/uuid"

type BaseEntity struct {
	Id uuid.UUID
	Title string
}