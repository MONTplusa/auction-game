package montplusa

import (
	"math"
	"math/rand"
	"time"

	"github.com/montplusa/auction-game/game"
)

const DEPTH = 11

// Montplusa implements a bidding AI using minimal dominant bids, +1-step, WTP, random bids, and deterministic evaluation.
type Montplusa struct{}

func (ai *Montplusa) GetName() string {
	return "Montplusa"
}

func (ai *Montplusa) SelectAction(gs *game.GameState, as *game.AuctionState, jewel *game.Jewel) [3]int {
	// Current player index is as.Turn
	me := as.Turn
	budgets := gs.Moneys[me]
	maxVal := as.MaxValue

	// 2. Generate candidate bids (including pass)
	candidates := make([][3]int, 0)
	// pass
	candidates = append(candidates, [3]int{0, 0, 0})

	// +1-step

	for c := 0; c < 3; c++ {
		plusOne := [3]int{maxVal[0], maxVal[1], maxVal[2]}
		plusOne[c] = maxVal[c] + 1
		if plusOne[c] <= budgets[c] {
			candidates = append(candidates, plusOne)
		}

	}

	// minimal dominant bids
	doms := generateMinimalDominant(gs, as)
	for _, b := range doms {
		valid := true
		for c := 0; c < 3; c++ {
			if b[c] > budgets[c] {
				valid = false
				break
			}
		}
		if valid {
			candidates = append(candidates, b)
		}
	}

	// random bids (up to 100 candidates in [maxVal..budgets])
	rand.Seed(time.Now().UnixNano())
	randBids := make([][3]int, 0)
	tries := 0
	for len(randBids) < 100 && tries < 1000 {
		tries++
		var rb [3]int
		ok := true
		for c := 0; c < 3; c++ {
			lb, ub := maxVal[c], budgets[c]
			if ub < lb {
				ok = false
				break
			}
			r := rand.Float64()
			if r < 0.1 {
				rb[c] = lb
			} else {
				rb[c] = lb + int(float64((ub-lb))-math.Sqrt(rand.Float64()*float64((ub-lb+1)*(ub-lb+1))))
			}
		}
		if ok {
			randBids = append(randBids, rb)
		}
	}
	for _, rb := range randBids {
		candidates = append(candidates, rb)
	}

	// dedupe
	seen := make(map[[3]int]bool)
	finalCands := make([][3]int, 0)
	for _, b := range candidates {
		if !seen[b] {
			seen[b] = true
			finalCands = append(finalCands, b)
		}
	}

	// 3. Evaluate candidates deterministically: pass->loss, bid->win
	bestVal := math.Inf(-1)
	bestBid := [3]int{0, 0, 0}
	// fmt.Println(gs.Phase, gs.Round)
	for _, bid := range finalCands {
		// skip non-pass bids that don't exceed current max
		if bid != [3]int{0, 0, 0} {
			exceed := false
			greater := true
			for c := 0; c < 3; c++ {
				if bid[c] > maxVal[c] {
					exceed = true
				}
				if bid[c] < maxVal[c] {
					greater = false
					break
				}
			}
			if !exceed || !greater {
				continue
			}
		}

		var val float64
		if bid == [3]int{0, 0, 0} {
			// pass -> simulate loss
			var lossState *game.GameState
			lossState = gs.Copy()
			minv := evaluateState(lossState, me)
			for i := range gs.Scores {
				if i == me {
					continue
				}
				if !as.Active[i] {
					continue
				}
				greater := false
				ok := true
				for c := 0; c < 3; c++ {
					if as.MaxValue[c] < gs.Moneys[i][c] {
						greater = true
					}
					if as.MaxValue[c] > gs.Moneys[i][c] {
						ok = false
						break
					}
				}
				if !ok {
					continue
				}
				if !greater && i != as.MaxPlayer {
					continue
				}
				lossState = simulate(gs, jewel, i, as.MaxValue)
				v := evaluateState(lossState, me)
				if v < minv {
					minv = v
				}
			}
			val = minv
		} else {
			// bid -> simulate win
			newState := simulate(gs, jewel, me, bid)
			minv := evaluateState(newState, me)
			for i := range gs.Scores {
				if i == me {
					continue
				}
				if !as.Active[i] {
					continue
				}
				greater := false
				ok := true
				for c := 0; c < 3; c++ {
					if bid[c] < gs.Moneys[i][c] {
						greater = true
					}
					if bid[c] > gs.Moneys[i][c] {
						ok = false
						break
					}
				}
				if !ok {
					continue
				}
				if !greater {
					continue
				}
				newState = simulate(gs, jewel, i, bid)
				v := evaluateState(newState, me)
				if v < minv {
					minv = v
				}
			}
			val = minv
		}

		if val > bestVal {
			bestVal = val
			bestBid = bid
		}
		// fmt.Println(bid, val)
	}

	return bestBid
}

func init() {
	game.RegisterAI("Montplusa", func() game.AI { return &Montplusa{} })
}

// generateMinimalDominant enumerates minimal dominant bids
func generateMinimalDominant(gs *game.GameState, as *game.AuctionState) [][3]int {
	me := as.Turn
	maxVal := as.MaxValue

	// thresholds per color
	T := make([][]int, 3)
	for c := 0; c < 3; c++ {
		set := map[int]bool{maxVal[c] + 1: true, maxVal[c]: true}
		for j, active := range as.Active {
			if !active || j == me {
				continue
			}
			set[gs.Moneys[j][c]+1] = true
		}
		for v := range set {
			T[c] = append(T[c], v)
		}
		sortInts(T[c])
	}

	// enumerate minimal
	cands := make([][3]int, 0)
	for _, r := range T[0] {
		for _, g := range T[1] {
			for _, b := range T[2] {
				bid := [3]int{r, g, b}
				same := true
				for i := 0; i < 3; i++ {
					if bid[i] < maxVal[i] {
						bid[i] = maxVal[i]
					}
					if bid[i] != maxVal[i] {
						same = false
						break
					}
				}
				if same {
					continue
				}
				if isDominantBid(bid, gs, as) {
					cands = append(cands, bid)
				}
			}
		}
	}
	return filterMinimal(cands)
}

// isDominantBid checks if bid exceeds each active opponent in at least one color
func isDominantBid(bid [3]int, gs *game.GameState, as *game.AuctionState) bool {
	me := as.Turn
	for j, active := range as.Active {
		if !active || j == me {
			continue
		}
		exceeds := false
		for c := 0; c < 3; c++ {
			if bid[c] > gs.Moneys[j][c] {
				exceeds = true
				break
			}
		}
		if !exceeds {
			return false
		}
	}
	return true
}

// filterMinimal removes bids dominated by another candidate
func filterMinimal(cands [][3]int) [][3]int {
	res := make([][3]int, 0)
	for i, a := range cands {
		keep := true
		for j, b := range cands {
			if i == j {
				continue
			}
			allLeq, anyLess := true, false
			for c := 0; c < 3; c++ {
				if b[c] > a[c] {
					allLeq = false
					break
				}
				if b[c] < a[c] {
					anyLess = true
				}
			}
			if allLeq && anyLess {
				keep = false
				break
			}
		}
		if keep {
			res = append(res, a)
		}
	}
	return res
}

// simulate applies state changes
func simulate(gs *game.GameState, jewel *game.Jewel, winner int, bid [3]int) *game.GameState {
	next := gs.Copy()
	for c := 0; c < 3; c++ {
		next.Moneys[winner][c] -= bid[c]
	}
	next.Scores[winner] += jewel.Point
	for c := 0; c < 3; c++ {
		next.Incomes[winner][c] += jewel.Income[c]
	}
	return next
}

// evaluateState scores a GameState for player me with dynamic weighted sums
func evaluateState(gs *game.GameState, me int) float64 {
	numPlayers := len(gs.Scores)
	value := 0.0
	for p := 0; p < numPlayers; p++ {
		v := evaluateVersus(gs, me, p)
		value += v
	}
	return value
}

func evaluateVersus(gs *game.GameState, me int, opp int) float64 {
	value := 0.0
	numPlayers := len(gs.Scores)
	weights := [DEPTH]float64{1.0, 0.7, 0.4, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.0}
	roundProp := float64(gs.Round) / float64(3*numPlayers)
	weights2 := [DEPTH]float64{}
	weights2[0] = (1 - roundProp) * weights[0]

	for i := 1; i < DEPTH; i++ {
		weights2[i] = weights[i-1]*roundProp + weights[i]*(1-roundProp)
	}

	for i := 0; i < DEPTH; i++ {
		if gs.Phase+i > 10 {
			break
		}
		v := 0.0
		for c := 0; c < 3; c++ {
			myMoney := float64(gs.Moneys[me][c])*math.Pow(0.5, math.Max(float64(i)-roundProp-0.5, 0)) + float64(i*gs.Incomes[me][c]+i*(i-1)*2)
			oppMoney := float64(gs.Moneys[opp][c])*math.Pow(0.5, math.Max(float64(i)-roundProp-0.5, 0)) + float64(i*gs.Incomes[opp][c]+i*(i-1)*2)
			/*for a := 0; a < numPlayers; a++ {
				if a == me {
					continue
				}
				m := float64(gs.Moneys[a][c])*math.Pow(0.7, math.Max(float64(i)-roundProp, 0)) + float64(i*gs.Incomes[a][c]+i*(i+1)*1)
				if m > oppMoney {
					oppMoney = m
				}
			}*/
			diff := math.Log(float64(myMoney)+1) - math.Log(float64(oppMoney)+1)
			v += diff
			if diff > 2 {
				v -= (diff - 2) * 0.8
			}
		}
		value += v * weights2[i] * 6

	}
	value += float64(gs.Scores[me]-gs.Scores[opp]) * weights[0] * float64(2+gs.Phase) / 12

	return value
}

// sortInts sorts small integer slices using insertion sort
func sortInts(a []int) {
	for i := 1; i < len(a); i++ {
		key := a[i]
		j := i - 1
		for j >= 0 && a[j] > key {
			a[j+1] = a[j]
			j--
		}
		a[j+1] = key
	}
}
