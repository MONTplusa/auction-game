//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"sort"
	"syscall/js"

	_ "github.com/montplusa/auction-game/ai/all"
	"github.com/montplusa/auction-game/game"
	"github.com/montplusa/auction-game/generator"
)

// Global state
var (
	gs               *game.GameState
	ais              []game.AI
	types            []string
	N                int
	currentJewel     *game.Jewel
	auctionState     *game.AuctionState
	states           []map[string]interface{}
	idx              int
	waitingHuman     bool
	previewJewel     bool
	nextIsPhaseStart bool
)

func initGame(this js.Value, args []js.Value) interface{} {
	N = args[0].Int()
	typesJSON := args[1].String()
	json.Unmarshal([]byte(typesJSON), &types)

	// New game state
	gs = game.NewGameState(N)
	gs.ApplyPhaseIncome()

	// Setup AIs/human channel
	game.HumanBidCh = make(chan [3]int, 1)
	ais = make([]game.AI, N)
	for i := 0; i < N; i++ {
		t := types[i]
		if t == "Human" {
			ais[i] = &game.HumanAI{Index: i}
			continue
		}
		if ctor, ok := game.Registry[t]; ok {
			ais[i] = ctor()
		}
	}

	// First auction
	currentJewel = generator.GenerateJewel()
	auctionState = game.NewAuctionState(0, N)

	// Reset snapshots
	states = nil
	idx = 0

	// Hook for in-auction actions
	game.ActionHook = func(as *game.AuctionState, j *game.Jewel, p int, bid [3]int) {
		recordSnapshot(j, as, false)
	}

	// Record initial phase-start snapshot (Phase 1 start)
	waitingHuman = false
	previewJewel = true
	nextIsPhaseStart = true // Set true to record Phase1 start
	recordSnapshot(currentJewel, auctionState, false)

	return nil
}

func nextStep(this js.Value, args []js.Value) interface{} {
	// Preview snapshot for round/phase start
	if previewJewel {
		recordSnapshot(currentJewel, auctionState, nextIsPhaseStart)
		// reset preview flags
		previewJewel = false
		nextIsPhaseStart = false
		return nil
	}

	// Game over guard
	if auctionState == nil {
		recordSnapshot(currentJewel, nil, false)
		return nil
	}

	// Handle human turn if active
	t := auctionState.Turn
	if types[t] == "Human" && auctionState.Active[t] {
		if !waitingHuman {
			waitingHuman = true
			recordSnapshot(currentJewel, auctionState, false)
			return nil
		}
		waitingHuman = false
	}

	// Execute one action
	finished := gs.StepAuction(auctionState, currentJewel, ais)
	recordSnapshot(currentJewel, auctionState, false)

	if finished {
		// Prepare next phaseflag before applying income
		gs.Round++
		if gs.Round > 3*N {
			// Phase rolls over
			nextIsPhaseStart = true
			gs.Phase++
			if gs.Phase > 10 {
				auctionState = nil
				return nil
			}
			gs.ApplyPhaseIncome()
			gs.Round = 1
		}
		// Start next auction, schedule preview
		currentJewel = generator.GenerateJewel()
		auctionState = game.NewAuctionState((gs.Round-1)%N, N)
		waitingHuman = false
		previewJewel = true
		// Initial preview for new round/phase will use nextIsPhaseStart
	}

	return nil
}

func submitBid(this js.Value, args []js.Value) interface{} {
	// Only accept on correct human turn
	if auctionState == nil {
		return nil
	}
	t := auctionState.Turn
	if types[t] != "Human" || !waitingHuman || !auctionState.Active[t] {
		return nil
	}
	r := args[0].Int()
	g := args[1].Int()
	b := args[2].Int()
	game.HumanBidCh <- [3]int{r, g, b}
	nextStep(js.Value{}, nil)
	return nil
}

func recordSnapshot(j *game.Jewel, as *game.AuctionState, isPhaseStart bool) {
	jInfo := map[string]interface{}{
		"Point":  currentJewel.Point,
		"Income": []int{currentJewel.Income[0], currentJewel.Income[1], currentJewel.Income[2]},
	}

	ranks := calculateRanks()
	players := make([]interface{}, N)
	for i := 0; i < N; i++ {
		m := gs.Moneys[i]
		inc := gs.Incomes[i]
		var bidSlice []int
		if as != nil {
			mv := as.MaxValue
			bidSlice = []int{mv[0], mv[1], mv[2]}
		}
		players[i] = map[string]interface{}{
			"Index":      i,
			"Name":       ais[i].GetName(),
			"Rank":       ranks[i],
			"Score":      gs.Scores[i],
			"Moneys":     []int{m[0], m[1], m[2]},
			"Income":     []int{inc[0], inc[1], inc[2]},
			"CurrentBid": bidSlice,
			"HasPassed":  as != nil && !as.Active[i],
		}
	}

	var auction interface{}
	if as != nil {
		aInfo := map[string]interface{}{"MaxPlayer": -1, "MaxValue": []int{0, 0, 0}}
		if as.MaxPlayer >= 0 {
			aInfo = map[string]interface{}{"MaxPlayer": as.MaxPlayer, "MaxValue": []int{as.MaxValue[0], as.MaxValue[1], as.MaxValue[2]}}
		}
		auction = aInfo
	} else {
		auction = nil
	}
	turn := -1
	if as != nil {
		turn = as.Turn
	}
	state := map[string]interface{}{
		"Phase":        gs.Phase,
		"Round":        gs.Round,
		"Turn":         turn,
		"PhaseStart":   isPhaseStart,
		"Jewel":        jInfo,
		"Auction":      auction,
		"Players":      players,
		"WaitingHuman": waitingHuman,
		"GameOver":     as == nil,
	}
	states = append(states, state)
	idx = len(states) - 1
}

func calculateRanks() []int {
	type pair struct{ idx, score, moneySum int }
	arr := make([]pair, N)
	for i := range arr {
		sum := gs.Moneys[i][0] + gs.Moneys[i][1] + gs.Moneys[i][2]
		arr[i] = pair{i, gs.Scores[i], sum}
	}
	sort.Slice(arr, func(i, j int) bool {
		if arr[i].score != arr[j].score {
			return arr[i].score > arr[j].score
		}
		return arr[i].moneySum > arr[j].moneySum
	})
	ranks := make([]int, N)
	rank := 1
	ranks[arr[0].idx] = rank
	for i := 1; i < len(arr); i++ {
		if arr[i].score != arr[i-1].score || arr[i].moneySum != arr[i-1].moneySum {
			rank = i + 1
		}
		ranks[arr[i].idx] = rank
	}
	return ranks
}

func getCurrentState(this js.Value, args []js.Value) interface{} {
	data, _ := json.Marshal(states[idx])
	return js.Global().Get("JSON").Call("parse", string(data))
}

func getAllStates(this js.Value, args []js.Value) interface{} {
	data, _ := json.Marshal(states)
	return js.Global().Get("JSON").Call("parse", string(data))
}

func main() {
	js.Global().Set("initGame", js.FuncOf(initGame))
	js.Global().Set("nextStep", js.FuncOf(nextStep))
	js.Global().Set("submitBid", js.FuncOf(submitBid))
	js.Global().Set("getCurrentState", js.FuncOf(getCurrentState))
	js.Global().Set("getAllStates", js.FuncOf(getAllStates))
	js.Global().Set("getAvailableAIs", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		keys := make([]interface{}, 0, len(game.Registry))
		for k := range game.Registry {
			keys = append(keys, k)
		}
		return js.ValueOf(keys)
	}))
	select {}
}
