package game

// HumanBidCh is used by HumanAI to receive bids from an external source (e.g., UI).
// バッファサイズを 1 にして、UI からの送信でデッドロックしないようにする
var HumanBidCh = make(chan [3]int, 1)

// HumanAI blocks waiting for a bid on HumanBidCh.
type HumanAI struct{ Index int }

// GetName returns the display name for human player.
func (h *HumanAI) GetName() string { return "Human" }

// SelectAction waits for the UI to submit a bid via HumanBidCh.
// It then returns that bid (or {0,0,0} for a pass).
func (h *HumanAI) SelectAction(gs *GameState, as *AuctionState, jewel *Jewel) [3]int {
	bid := <-HumanBidCh
	return bid
}
