package montplusai

import (
	"math"
	"math/rand"
	"time"

	"github.com/montplusa/auction-game/game"
)

// MontplusAI implements a bidding AI using minimal dominant bids, +1-step, WTP, random bids, and deterministic evaluation.
type MontplusAI struct{}

func (ai *MontplusAI) GetName() string {
	return "MontplusAI Lv1"
}

func (ai *MontplusAI) SelectAction(gs *game.GameState, as *game.AuctionState, jewel *game.Jewel) [3]int {
	// Current player index is as.Turn
	me := as.Turn
	budgets := gs.Moneys[me]
	maxVal := as.MaxValue
	phaseLeft := 10 - gs.Phase

	// 1. Calculate Willingness to Pay (WTP)
	alpha, beta := 1.2, 0.8
	scoreVal := alpha * float64(jewel.Point)
	incomeVal := 0.0
	for _, inc := range jewel.Income {
		if inc > 0 {
			incomeVal = beta * float64(inc) * float64(phaseLeft)
		}
	}
	wtp := int(math.Round(scoreVal + incomeVal))

	// 2. Generate candidate bids (including pass)
	candidates := make([][3]int, 0)
	// pass
	candidates = append(candidates, [3]int{0, 0, 0})

	// +1-step on income color
	incColor := 0
	for idx, inc := range jewel.Income {
		if inc > 0 {
			incColor = idx
			break
		}
	}
	plusOne := [3]int{maxVal[0], maxVal[1], maxVal[2]}
	plusOne[incColor] = maxVal[incColor] + 1
	if plusOne[incColor] <= budgets[incColor] {
		candidates = append(candidates, plusOne)
	}

	// WTP buy (all on income color)
	wtpBid := [3]int{0, 0, 0}
	if wtp > maxVal[incColor] && wtp <= budgets[incColor] {
		wtpBid[incColor] = wtp
		candidates = append(candidates, wtpBid)
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
			rb[c] = lb + rand.Intn(ub-lb+1)
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
	for _, bid := range finalCands {
		// skip non-pass bids that don't exceed current max
		if bid != [3]int{0, 0, 0} {
			exceed := false
			for c := 0; c < 3; c++ {
				if bid[c] > maxVal[c] {
					exceed = true
					break
				}
			}
			if !exceed {
				continue
			}
		}

		var val float64
		if bid == [3]int{0, 0, 0} {
			// pass -> simulate loss
			var lossState *game.GameState
			if as.MaxPlayer >= 0 {
				lossState = simulateLoss(gs, as, jewel, as.MaxPlayer)
			} else {
				lossState = gs.Copy()
			}
			val = evaluateState(lossState, me)
		} else {
			// bid -> simulate win
			winState := simulateWin(gs, as, jewel, me, bid)
			val = evaluateState(winState, me)
		}

		if val > bestVal {
			bestVal = val
			bestBid = bid
		}
	}

	return bestBid
}

func init() {
	game.RegisterAI("MontplusAI Lv1", func() game.AI { return &MontplusAI{} })
}

// generateMinimalDominant enumerates minimal dominant bids
func generateMinimalDominant(gs *game.GameState, as *game.AuctionState) [][3]int {
	me := as.Turn
	maxVal := as.MaxValue

	// thresholds per color
	T := make([][]int, 3)
	for c := 0; c < 3; c++ {
		set := map[int]bool{maxVal[c] + 1: true}
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

// simulateWin applies state changes when me wins
func simulateWin(gs *game.GameState, as *game.AuctionState, jewel *game.Jewel, me int, bid [3]int) *game.GameState {
	next := gs.Copy()
	for c := 0; c < 3; c++ {
		next.Moneys[me][c] -= bid[c]
	}
	next.Scores[me] += jewel.Point
	for c := 0; c < 3; c++ {
		next.Incomes[me][c] += jewel.Income[c]
	}
	return next
}

// simulateLoss applies state changes when opp wins
func simulateLoss(gs *game.GameState, as *game.AuctionState, jewel *game.Jewel, opp int) *game.GameState {
	next := gs.Copy()
	for c := 0; c < 3; c++ {
		next.Moneys[opp][c] -= as.MaxValue[c]
	}
	next.Scores[opp] += jewel.Point
	for c := 0; c < 3; c++ {
		next.Incomes[opp][c] += jewel.Income[c]
	}
	return next
}

// evaluateState scores a GameState for player me with dynamic weighted sums
func evaluateState(gs *game.GameState, me int) float64 {
	// Dynamic weights based on phase (0-10)
	phaseProg := float64(gs.Phase) / 10.0
	roundProg := float64(gs.Round) / 3.0 / float64(len(gs.Scores))
	// Early game: prioritize coin ratios; Late game: prioritize score
	wScore := 0.2 + 1.4*phaseProg
	wCoinNow := 7 * (1.4 - phaseProg) * (1 - 1/(math.Exp((1-roundProg)*5)))
	wCoinNext := 7 * (1.4 - phaseProg) * (1 / (math.Exp((1 - roundProg) * 5)))
	wIncome := 3 * (1.0 - phaseProg)

	// 1. score difference sum
	scMe := float64(gs.Scores[me])
	scoreSum := 0.0
	for j, sc := range gs.Scores {
		if j == me {
			continue
		}
		scoreSum += scMe - float64(sc)
	}

	// 2. current coin ratio sum (log diff)
	coinNowSum := 0.0
	for j, mny := range gs.Moneys {
		if j == me {
			continue
		}
		for c := 0; c < 3; c++ {
			coinNowSum += math.Log(float64(gs.Moneys[me][c])+1) - math.Log(float64(mny[c])+1)
		}
	}

	// 3. next-phase coin ratio sum (after incomes)
	coinNextSum := 0.0
	for j := range gs.Moneys {
		if j == me {
			continue
		}
		for c := 0; c < 3; c++ {
			myNext := float64(gs.Moneys[me][c] + gs.Incomes[me][c])
			oppNext := float64(gs.Moneys[j][c] + gs.Incomes[j][c])
			coinNextSum += math.Log(myNext+1) - math.Log(oppNext+1)
		}
	}

	// 4. income-rate difference sum
	incomeSum := 0.0
	for j, incArr := range gs.Incomes {
		if j == me {
			continue
		}
		for c := 0; c < 3; c++ {
			incomeSum += float64(gs.Incomes[me][c]) - float64(incArr[c])
		}
	}

	return wScore*scoreSum + wCoinNow*coinNowSum + wCoinNext*coinNextSum + wIncome*incomeSum
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
