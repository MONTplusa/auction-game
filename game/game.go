package game

// GameState represents the overall state of the auction game across phases and rounds.
type GameState struct {
	Phase   int      // 現在のフェーズ (1～10)
	Round   int      // フェーズ内の現在ラウンド (1～3N)
	Scores  []int    // 各プレイヤーの累計得点 (長さ N)
	Incomes [][3]int // 各プレイヤーがフェーズ開始時に得るコイン収入 (長さ N, 各要素は [赤,緑,青])
	Moneys  [][3]int // 各プレイヤーの現在所持コイン (長さ N, 各要素は [赤,緑,青])
}

// AuctionState holds the state for a single auction round.
type AuctionState struct {
	MaxPlayer         int    // 暫定最高入札者のプレイヤー番号（未入札なら -1）
	MaxValue          [3]int // 暫定最高入札額 ([赤,緑,青], 未入札なら {0,0,0})
	Turn              int    // 現在手番のプレイヤー番号 (0～N-1)
	Active            []bool // 有効な入札者一覧
	activeCount       int    // 残り有効入札者数
	consecutivePasses int    // 連続パス数
}

// NewAuctionState returns a freshly initialized AuctionState for a new auction.
func NewAuctionState(startTurn, numPlayers int) *AuctionState {
	as := &AuctionState{
		Turn:   startTurn,
		Active: make([]bool, numPlayers),
		// MaxPlayer/MaxValue はゼロ値のまま（MaxPlayer=-1がデフォルトになるように初期化）
	}
	for i := range as.Active {
		as.Active[i] = true
	}
	as.activeCount = numPlayers
	as.consecutivePasses = 0
	as.MaxPlayer = -1
	as.MaxValue = [3]int{0, 0, 0}
	return as
}

// Jewel describes the auction item.
type Jewel struct {
	Point  int    // 入手した際に得られる得点 (1～10)
	Income [3]int // 各フェーズごとに得られるコイン収入 ([赤,緑,青])
}

// AI defines the bid strategy interface.
type AI interface {
	// GetName returns the AI's display name.
	GetName() string

	// SelectAction is called on each auction turn:
	//   gameState   -- 全体のゲーム進行状態
	//   auctionState-- 現在オークションの状態
	//   jewel       -- 今回のオークション対象宝石
	// 戻り値は提示額 [赤,緑,青]、{0,0,0} は降りるを意味する。
	SelectAction(gameState *GameState, auctionState *AuctionState, jewel *Jewel) [3]int
}
