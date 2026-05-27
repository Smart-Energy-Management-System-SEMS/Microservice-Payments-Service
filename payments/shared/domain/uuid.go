package domain

import "github.com/google/uuid"

func NewID() uuid.UUID {
	return uuid.New()
}

func ParseID(value string) (uuid.UUID, error) {
	return uuid.Parse(value)
}
