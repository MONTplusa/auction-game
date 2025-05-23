package kimeuti_tarou

import (
	"math/rand"

	"github.com/montplusa/auction-game/game"
)

// 本体
type KimeutiAI struct {
}

func (ai *KimeutiAI) GetName() string { return "決打太郎" }

func (ai *KimeutiAI) SelectAction(gs *game.GameState,
	as *game.AuctionState, j *game.Jewel) [3]int {
	player := as.Turn
	sumValue := 0
	sumMoney := 0
	for i := 0; i < 3; i++ {
		sumValue += as.MaxValue[i]
		sumMoney += gs.Moneys[player][i]
		if gs.Moneys[player][i] < as.MaxValue[i] {
			return [3]int{0, 0, 0}
		}
	}
	if sumValue == sumMoney {
		return [3]int{0, 0, 0}
	}

	numPlayers := len(gs.Scores)
	// phase := gs.Phase
	round := gs.Round
	p := rand.ExpFloat64() / float64((numPlayers)*3-round+1)
	numbid := int(float64(sumMoney) * p)
	sumIncome := 0
	for i := range j.Income {
		sumIncome += j.Income[i]
	}
	if numbid < sumIncome {
		numbid = sumIncome
	}
	if numbid > sumMoney {
		numbid = sumMoney
	}

	if numbid <= sumValue {
		return [3]int{0, 0, 0}
	}

	bid := as.MaxValue
	coins := gs.Moneys[player]

	for i := sumValue; i < numbid; i++ {
		c := rand.Intn(3)
		if bid[c] < coins[c] {
			bid[c]++
		} else {
			i--
		}
	}

	return bid
}

func init() {
	game.RegisterAI("決打太郎", func() game.AI {
		return &KimeutiAI{}
	})
}
