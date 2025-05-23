package generator

import (
	"math/rand"
	"time"

	"github.com/montplusa/auction-game/game"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateJewel はランダムに Jewel を生成します。
// - Point: 1～10 の等確率
// - Income: 3 種類のうち 1 種を選び、そのコイン収入を 0～5 でランダムに設定。他の 2 種は 0。
func GenerateJewel() *game.Jewel {
	// 得点を 1～10 の範囲で生成
	point := rand.Intn(10) + 1

	// 収入配列を初期化し、ランダムに 1 種類を設定
	income := [3]int{0, 0, 0}
	coinType := rand.Intn(3)        // 0:赤, 1:緑, 2:青
	income[coinType] = rand.Intn(6) // 0～5

	return &game.Jewel{
		Point:  point,
		Income: income,
	}
}
