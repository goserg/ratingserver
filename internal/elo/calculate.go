package elo

import "math"

type Points float64

const (
	Win  Points = 1
	Draw        = 0.5
	Lose        = 0
)

// Calculate new rating.
// Ra - player A rating.
// Rb - player B rating.
// K - coefficient: if >= 2400 then 10; if < 2400 then 20; if first 40 games then 40.
// Sa - points: 1 for win; 0.5 for draw; 0 for lose.
func Calculate(Ra int, Rb int, K int, Sa Points) int {
	ra := float64(Ra)
	rb := float64(Rb)
	k := float64(K)

	Ea := 1.0 / (1.0 + math.Pow(10, (rb-ra)/400.0))
	ra = ra + k*(float64(Sa)-Ea)
	return int(math.Round(ra))
}
