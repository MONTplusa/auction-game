package game

// NewGameState initializes and returns a GameState for N players.
// Each player starts with 10 coins of each color, zero score, and zero income.
func NewGameState(N int) *GameState {
	scores := make([]int, N)
	incomes := make([][3]int, N)
	moneys := make([][3]int, N)
	for i := 0; i < N; i++ {
		moneys[i] = [3]int{10, 10, 10}
	}
	return &GameState{
		Phase:   1,
		Round:   1,
		Scores:  scores,
		Incomes: incomes,
		Moneys:  moneys,
	}
}

// ApplyPhaseIncome adds each player's current income to their moneys.
// Should be called at the start of each phase (including after Phase++).
func (g *GameState) ApplyPhaseIncome() {
	for i := range g.Moneys {
		for c := 0; c < 3; c++ {
			g.Moneys[i][c] += g.Incomes[i][c]
		}
	}
}

// AdvanceRound progresses the game to the next round.
// When rounds in the current phase exceed 3*N, it rolls over to the next phase,
// applies phase income, and resets the round counter.
func (g *GameState) AdvanceRound() {
	N := len(g.Scores)
	g.Round++
	if g.Round > 3*N {
		// Move to next phase
		g.Phase++
		// Give income from owned jewels
		g.ApplyPhaseIncome()
		// Reset round to 1
		g.Round = 1
	}
}
