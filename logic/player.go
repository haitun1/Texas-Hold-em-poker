package logic

import "reflect"

// Player 玩家
type Player struct {
	HandPoker     []int32 // 手牌
	RealHandPoker []int32 // 真实手牌（5张）
	HandCount     int     // 手牌数量
	PlayerStatus  bool    // 状态
	PlayerScore   int64   // 分数
	Fold          bool    // 弃牌（盖牌）
	Bet           bool    // 押注
	Call          bool    // 跟注
	Check         bool    // 让牌
	Raise         bool    // 加注
	Re_raise      bool    // 再加注(可能不需要)
	All_In        bool    // 全押（全押的情况：底分<跟注， 手动全押：底分=跟注）
	CardType      int16   // 玩家当前最大牌型
	BetScore      int     // 下注分数（本轮）
	MainPot       int64   // 主池（边池计算可以使用底池-主池）
}

// Init 玩家信息初始化
func (p *Player) Init() {
	p.HandPoker = make([]int32, 7)
	p.RealHandPoker = make([]int32, 5)
	p.HandCount = 0
	p.BetScore = 0
	p.MainPot = 0
	p.Fold = false
	p.Bet = false
	p.Call = false
	p.Check = false
	p.Raise = false
	p.Re_raise = false
	p.All_In = false
	p.CardType = 0
}

// Reset 玩家信息重置
func (p *Player) Reset() {
	p.HandPoker = make([]int32, 7)
	p.RealHandPoker = make([]int32, 5)
	p.BetScore = 0
	p.MainPot = 0
	p.HandCount = 0
	p.BetScore = 0
	p.Fold = false
	p.Bet = false
	p.Call = false
	p.Check = false
	p.Raise = false
	p.Re_raise = false
	p.All_In = false
	p.CardType = 0
}

// HoleHand 发牌
func (p *Player) HoleHand(hand []int32) {
	copy(p.HandPoker[0:2], hand[0:2])
	p.HandCount = 2
	p.PlayerStatus = true
}

// FlopHand 翻牌
func (p *Player) FlopHand(hand []int32) {
	copy(p.HandPoker[2:5], hand[0:3])
	p.HandCount = 5
}

// TurnHand 转牌
func (p *Player) TurnHand(hand []int32) {
	copy(p.HandPoker[5:6], hand[3:4])
	p.HandCount = 6
}

// RiverHand 河牌
func (p *Player) RiverHand(hand []int32) {
	copy(p.HandPoker[6:7], hand[4:5])
	p.HandCount = 7
}

// zeroSlice 重置切片
func zeroSlice(v interface{}, refType reflect.Kind) {
	if reflect.TypeOf(v).Kind() != reflect.Slice {
		panic("zeroSlice wrong type")
	}
	slice := reflect.ValueOf(v)

	for i := 0; i < slice.Len(); i++ {
		value := slice.Index(i)
		if refType >= reflect.Uint && refType <= reflect.Uintptr {
			value.SetUint(0)
		} else if refType >= reflect.Int && refType <= reflect.Int64 {
			value.SetInt(0)
		} else if refType == reflect.Bool {
			value.SetBool(false)
		}
	}
}
