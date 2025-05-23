package game

// ctor は「AI インスタンスを返す関数」型
type AICtor func() AI

// Registry は文字列 → 生成関数 のマップ
var Registry = map[string]AICtor{}

// RegisterAI は AI 実装側が init() で呼ぶ関数
func RegisterAI(name string, ctor AICtor) {
	Registry[name] = ctor
}
