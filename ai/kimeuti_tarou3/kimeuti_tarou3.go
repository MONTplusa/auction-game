package kimeuti_tarou3

import (
	"math/rand"

	"github.com/montplusa/auction-game/game"
)

// 本体
type KimeutiAI struct {
}

func (ai *KimeutiAI) GetName() string { return "決打太郎Lv3" }

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

	// numPlayers := len(gs.Scores)
	phase := gs.Phase
	// round := gs.Round
	sumIncome := 0
	for i := range j.Income {
		sumIncome += j.Income[i]
	}
	p := (2 + rand.NormFloat64()) / 12
	numbid := int((float64(j.Point) + float64(5*sumIncome)) * (1 + float64(phase)*0.3) * p)
	if numbid < 1 {
		numbid = 1
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

	for bidloop := sumValue; bidloop < numbid; bidloop++ {
		props := [3]float64{0, 0, 0}
		sumProps := 0.0
		for i := range props {
			opps := 0
			for j := range gs.Moneys {
				if !as.Active[j] {
					continue
				}
				diff := gs.Moneys[j][i] - bid[i]
				if diff < 0 {
					diff = 0
				}
				opps += diff
			}
			prop := 0.0
			if coins[i]-bid[i] > 0 {
				prop = float64(coins[i]-bid[i]) / float64(opps)
			}
			props[i] = prop
			sumProps += props[i]
		}
		r := rand.Float64() * sumProps
		c := 0
		cutSum := 0.0
		for {
			cutSum += props[c]
			if cutSum > r {
				break
			}
			c++
		}
		// c := rand.Intn(3)
		if bid[c] < coins[c] {
			bid[c]++
		} else {
			bidloop--
		}
	}

	return bid
}

func init() {
	game.RegisterAI("決打太郎Lv3", func() game.AI {
		return &KimeutiAI{}
	})
}
