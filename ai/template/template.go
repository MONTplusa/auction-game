package template

import (
	"github.com/montplusa/auction-game/game"
)

type TemplateAI struct{}

// GetName returns the display name of the AI.
func (ai *TemplateAI) GetName() string {
	return "TemplateAI"
}

func (ai *TemplateAI) SelectAction(gs *game.GameState, as *game.AuctionState, jewel *game.Jewel) [3]int {

	return [3]int{0, 0, 0}
}

func init() {
	game.RegisterAI("TemplateAI", func() game.AI { return &TemplateAI{} })
}
