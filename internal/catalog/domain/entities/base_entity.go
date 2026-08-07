package entities

import "github.com/google/uuid"

type BaseEntity struct {  //это базовая сущность для всех других сущностей,сущность это объект который описывает реальный мир
	Id uuid.UUID
	Title string
}