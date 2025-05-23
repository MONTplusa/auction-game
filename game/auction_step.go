package game

// StepAuction executes exactly one action in the current auction.
// Returns true if the auction is completed with a winner, false otherwise.
func (g *GameState) StepAuction(as *AuctionState, jewel *Jewel, ais []AI) bool {
	N := len(g.Scores)
	// Initialize internal state on first call
	if as.Active == nil {
		as.Active = make([]bool, N)
		for i := range as.Active {
			as.Active[i] = true
		}
		as.activeCount = N
		as.consecutivePasses = 0
		as.MaxPlayer = -1
		as.MaxValue = [3]int{0, 0, 0}
	}

	player := as.Turn
	if as.Active[player] {
		// Player makes a bid
		bidVal := ais[player].SelectAction(g, as, jewel)
		// Hook for UI
		if ActionHook != nil {
			ActionHook(as, jewel, player, bidVal)
		}
		// Validate
		if isValidBid(bidVal, as.MaxValue) && hasEnoughMoney(g.Moneys[player], bidVal) {
			as.MaxValue = bidVal
			as.MaxPlayer = player
			as.consecutivePasses = 0
		} else {
			as.Active[player] = false
			as.activeCount--
			as.consecutivePasses++
		}
	}
	// Advance turn
	as.Turn = (as.Turn + 1) % N

	// Check completion condition
	if as.activeCount > 1 {
		return false
	}
	if as.activeCount == 1 && as.MaxPlayer < 0 {
		// No bids yet
		return false
	}
	// Finalize auction: award jewel
	if as.MaxPlayer >= 0 {
		for c := 0; c < 3; c++ {
			g.Moneys[as.MaxPlayer][c] -= as.MaxValue[c]
		}
		g.Scores[as.MaxPlayer] += jewel.Point
		for c := 0; c < 3; c++ {
			g.Incomes[as.MaxPlayer][c] += jewel.Income[c]
		}
	}
	return true
}
