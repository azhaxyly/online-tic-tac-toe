package models

type Game struct {
	PlayerX    string
	PlayerO    string
	Turn       string
	Board      [9]string
	IsFinished bool
	Winner     string
	PlayAgainX bool
	PlayAgainO bool
}
