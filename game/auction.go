package game

// ActionHook is called after each player's SelectAction, for snapshotting or other callbacks.
var ActionHook func(auctionState *AuctionState, jewel *Jewel, player int, bid [3]int)

// isValidBid returns true if bid >= maxVal on all colors and strictly > on at least one.
func isValidBid(bid, maxVal [3]int) bool {
	greater := false
	for i := 0; i < 3; i++ {
		if bid[i] < maxVal[i] {
			return false
		}
		if bid[i] > maxVal[i] {
			greater = true
		}
	}
	return greater
}

// hasEnoughMoney returns true if player has enough coins for the bid.
func hasEnoughMoney(money, bid [3]int) bool {
	for i := 0; i < 3; i++ {
		if money[i] < bid[i] {
			return false
		}
	}
	return true
}
