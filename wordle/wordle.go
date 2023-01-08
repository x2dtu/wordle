package wordle

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getWord() string {
	index := rand.Intn(len(words))
	return words[index]
}

type Wordle struct {
	Target           string
	Guesses          int
	PreviousGuesses  []string
	GameOver         bool
	EnteredGibberish bool
}

func New() *Wordle {
	w := Wordle{
		Target:          getWord(),
		Guesses:         0,
		PreviousGuesses: make([]string, 0),
		GameOver:        false,
	}
	return &w
}
