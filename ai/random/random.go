package random

import (
	"math/rand"

	"github.com/montplusa/auction-game/game"
)

// RandomAI is a simple AI that randomly decides to bid or pass.
type RandomAI struct{}

// GetName returns the display name of the AI.
func (ai *RandomAI) GetName() string {
	return "RandomAI"
}

// SelectAction returns either a pass ([0,0,0]) or a random valid bid.
// It randomly selects one coin color to increase above the current max.
func (ai *RandomAI) SelectAction(gs *game.GameState, as *game.AuctionState, jewel *game.Jewel) [3]int {
	player := as.Turn
	maxVal := as.MaxValue
	money := gs.Moneys[player]

	// 50% chance to pass
	if rand.Float64() < 0.5 {
		return [3]int{0, 0, 0}
	}

	bid := [3]int{}
	for c := 0; c < 3; c++ {
		// If insufficient funds, pass
		if money[c] < maxVal[c] {
			return [3]int{0, 0, 0}
		}

		// Generate a bid: match maxVal on all, exceed on chosen color
		// Choose an amount between maxVal[c]+1 and money[c]
		rangeMax := money[c] - maxVal[c]
		amount := rand.Intn(rangeMax+1) + maxVal[c]
		bid[c] = amount
	}
	return bid
}

func init() {
	game.RegisterAI("RandomAI", func() game.AI { return &RandomAI{} })
}
