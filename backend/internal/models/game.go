package models

import "time"

type Game struct {
	PlayerX            string
	PlayerO            string
	Turn               string
	Board              [9]string
	IsFinished         bool
	Winner             string
	PlayAgainX         bool
	PlayAgainO         bool
	RematchRequestedAt time.Time
	RematchTimer       *time.Timer
}
