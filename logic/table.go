package logic

import (
	"math/rand"
	"sort"
	"time"
)

type SortByte []int32

func (a SortByte) Len() int             { return len(a) }
func (a SortByte) Swap(i, j int)        { a[i], a[j] = a[j], a[i] }
func (a SortByte) Less(i, j int) bool { // 从大到小排序
	if (a[i] & 0xF) == (a[j] & 0xF) {
		return a[i] > a[j]
	} else {
		return (a[i] & 0xF) > (a[j] & 0xF)
	}
}

// Table 桌子
type Table struct {
	GameRules    int64    // 游戏配置
	RoomType     int32    // 游戏房间
	PokerNumber  int      // 扑克数量
	CardsCount   [13]int  // 扑克剩余数目
	ChairNumber  int      // 椅子数量
	PlayerNumber int      // 玩家数量
	CardData     []int32  // 牌库
	Players      []Player // 游戏玩家
	//	EndPoints          int64          // 游戏底分
	TableCurrentStatus GameStatusType // 游戏状态
	Banker             int            // 庄家（D）
	CommonCard         []int32        // 公共牌
	Pont               int64          // 底池（总下注）
	LastHandleBet      int            // 上个玩家操作分数（弃牌， 全押不改变分数,一轮游戏结束清零）
	FlopSettle         bool           // 剩下一个人结算
}

// Reset 重置
func (t *Table) Reset() {
	for i := 0; i != t.ChairNumber; i++ {
		if t.Players[i].PlayerStatus == true {
			t.Players[i].Reset()
		}
	}
	t.FlopSettle = false
	t.TableCurrentStatus = StatusFree
	t.PokerNumber = 52
	t.CommonCard = make([]int32, 5)
	t.Pont = 0
	t.Banker = t.EnsureChairID(t.Banker + 1)
	t.LastHandleBet = 0
}

// Init 初始化
func (t *Table) Init() {
	//t.Players = make([]Player, 6)         // 暂时这样写（方便测试）
	for i := 0; i != t.ChairNumber; i++ { // 首次初始化全部椅子
		//	if t.Players[i].PlayerStatus == true {
		t.Players[i].Init()
		//	}
	}
	t.TableCurrentStatus = StatusFree
	t.CommonCard = make([]int32, 5)
	t.CardData = make([]int32, 52)
	t.FlopSettle = false
	t.Pont = 0
	//	t.Banker = ？ // 首轮庄家确定
	t.LastHandleBet = 0
}

// EnsureChairID 确定椅子号 (可能存在问题)
func (t *Table) EnsureChairID(chairID int) int {
	if chairID < 0 {
		i := chairID + t.ChairNumber // 可以不用for循环一定为正
		for i < 0 {
			i += t.ChairNumber
		}
		for ; i != chairID; i-- {
			if i < 0 {
				i += t.ChairNumber
			}
			if t.Players[i].PlayerStatus == true {
				chairID = i
				break
			}
		}
	} else if chairID >= t.ChairNumber {
		i := chairID - t.ChairNumber
		for i >= t.ChairNumber {
			i -= t.ChairNumber
		}
		for ; i != chairID; i++ {
			if i > t.ChairNumber {
				i -= t.ChairNumber
			}
			if t.Players[i].PlayerStatus == true {
				chairID = i
				break
			}
		}
	}
	return chairID
}

// SetGameRule 获取配置
func (t *Table) SetGameRule(gameRule int64) {
	t.GameRules = gameRule
}

// RandomShuffle 随机牌库（src）
func (t *Table) RandomShuffle(src []int32) []int32 {
	dest := make([]int32, len(src))
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	perm := r.Perm(len(src))
	for i, v := range perm {
		dest[v] = src[i]
	}
	for i := 0; i != 13; i++ {
		t.CardsCount[i] = 4
	}
	return dest
}

// GetCardValue 返回牌值
func (t *Table) GetCardValue(PokerData int32) int32 {
	return PokerData & 0x0f
}

// GetCardColor 返回花色
func (t *Table) GetCardColor(PokerData int32) int32 {
	return PokerData & 0xf0
}

// GetCardType 获取牌值对应牌型
func (t *Table) GetCardType(cardData []int32, cardCount int16) int16 {
	/*cardData := make([]int32, len(handData)) //
	copy(cardData[0:], handData[0:])
	sort.Sort(SortByte(cardData[0:]))*/
	sameColor := true
	lineCard := true
	firstColor := t.GetCardColor(cardData[0])
	firstValue := t.GetCardValue(cardData[0])

	//牌形分析
	for i := int16(1); i < cardCount; i++ {
		/*
			// 如果有王的话
			if firstColor == 0x40 || t.GetCardColor(cardData[i]) == 0x40 {
				sameColor = false
				lineCard = false
				break
			}
		*/
		//数据分析
		if t.GetCardColor(cardData[i]) != firstColor {
			sameColor = false
		}
		if t.GetCardValue(cardData[i]) != (firstValue - int32(i)) {
			lineCard = false
			if (t.GetCardValue(cardData[i]) == firstValue-int32(i)-8) && (t.GetCardValue(cardData[i]) > 0x9 || t.GetCardValue(cardData[i]) == 0x1) {
				lineCard = true
			}
		}

		//结束判断
		if sameColor == false && lineCard == false {
			break
		}
	}

	// 皇家同花顺
	if sameColor == true && t.GetCardValue(cardData[4]) == int32(10) {
		return Royal_Straight_Flush
	}
	//同花顺类型
	if sameColor == true && lineCard == true {
		return Straight_Flush
	}
	// 四条
	//	thereOfKind := false
	TwoPair := false
	OnePair := false
	//	FullHouse := false
	ThreeOfKing := false
	for i := int16(0); i < cardCount; i++ {
		//变量定义
		sameCount := int16(1)
		sameCardData := []int32{cardData[i], 0, 0, 0}
		logicValue := t.GetCardValue(cardData[i])

		//获取同牌
		for j := i + 1; j < cardCount; j++ {
			//逻辑对比
			if t.GetCardValue(cardData[j]) != logicValue {
				break
			}

			//设置扑克
			sameCardData[sameCount] = cardData[j]
			sameCount++
		}
		if sameCount == 4 { // 四条
			return Four_of_king
		}
		if sameCount == 3 {
			ThreeOfKing = true
		}
		if sameCount == 2 {
			if OnePair == true {
				TwoPair = true
			} else {
				OnePair = true
			}
		}
		//设置递增
		i += sameCount - 1
	}

	//葫芦 三带二
	if ThreeOfKing == true && OnePair == true {
		return Full_house
	}
	//同花类型
	if sameColor == true && lineCard == false {
		return Flush
	}

	// 顺子
	if sameColor == false && lineCard == true {
		return Straight
	}
	// 三条
	if ThreeOfKing == true {
		return Three_Of_King
	}
	if TwoPair == true {
		return Two_pair
	}
	if OnePair == true {
		return One_pair
	}
	if len(cardData) == 5 {
		return High_card
	}
	return CtZero
}

// CompareCard 比较牌型(同为皇家同花顺不在这里考虑)存在平局，用bool不适合，以后尝试优化
func (t *Table) CompareCard(firstData []int32, nextData []int32, cardCount int16) int {
	if len(firstData) != 5 || len(nextData) != 5 {
		return Tied
	}
	cardData := make([]int32, len(firstData))
	copy(cardData[0:], firstData[0:])
	sort.Sort(SortByte(cardData[0:]))
	cardData1 := make([]int32, len(nextData))
	copy(cardData1[0:], nextData[0:])
	sort.Sort(SortByte(cardData1[0:]))
	//获取点数
	nextType := t.GetCardType(cardData1, cardCount)
	firstType := t.GetCardType(cardData, cardCount)

	//点数判断
	if firstType != nextType {
		if firstType > nextType {
			return Win
		} else {
			return Lose
		}
	} else if firstType == Royal_Straight_Flush {
		return Tied
	} else if firstType == Straight_Flush {
		if t.GetCardValue(cardData[0]) > t.GetCardValue(cardData1[0]) {
			return Win
		} else if t.GetCardValue(cardData[0]) < t.GetCardValue(cardData1[0]) {
			return Lose
		} else {
			return Tied
		}
	} else if firstType == Four_of_king {
		if t.GetCardValue(cardData[2]) > t.GetCardValue(cardData1[2]) {
			return Win
		} else if t.GetCardValue(cardData[2]) < t.GetCardValue(cardData1[2]) {
			return Lose
		} else {
			return Tied
		}
	} else if firstType == Full_house {
		if t.GetCardValue(cardData[2]) > t.GetCardValue(cardData1[2]) {
			return Win
		} else if t.GetCardValue(cardData[2]) < t.GetCardValue(cardData1[2]) {
			return Lose
		} else {
			return Tied
		}
	} else if firstType == Flush {
		for i := 0; i != 5; i++ {
			if t.GetCardValue(cardData[i]) > t.GetCardValue(cardData1[i]) {
				return Win
			} else if t.GetCardValue(cardData[i]) < t.GetCardValue(cardData1[i]) {
				return Lose
			} else {
				if i == 4 {
					return Tied
				}
			}
		}
	} else if firstType == Straight {
		if t.GetCardValue(cardData[0]) > t.GetCardValue(cardData1[0]) {
			return Win
		} else if t.GetCardValue(cardData[0]) < t.GetCardValue(cardData1[0]) {
			return Lose
		} else {
			return Tied
		}
	} else if firstType == Three_Of_King {
		if t.GetCardValue(cardData[2]) > t.GetCardValue(cardData1[2]) {
			return Win
		} else if t.GetCardValue(cardData[2]) < t.GetCardValue(cardData1[2]) {
			return Lose
		} else {
			return Tied
		}
	} else if firstType == Two_pair { // 判断存在问题（可能）
		i := 0
		m := 2 // 确定唯一单数的位置
		n := 2
		if t.GetCardValue(cardData[i]) != t.GetCardValue(cardData[i+1]) { // 不用考虑越界问题，因为牌型判断过
			i++
			m = 0
		}
		j := 0
		if t.GetCardValue(cardData1[j]) != t.GetCardValue(cardData1[j+1]) {
			j++
			n = 0
		}
		if t.GetCardValue(cardData[i]) > t.GetCardValue(cardData1[j]) {
			return Win
		} else if t.GetCardValue(cardData[j]) < t.GetCardValue(cardData1[j]) {
			return Lose
		} else {
			if m == 0 {
				i += 2 // 因为是两对 +2后肯定是对子
			} else {
				i = m + 1
				if t.GetCardValue(cardData[i]) != t.GetCardValue(cardData[i+1]) {
					i++
					m = 4
				}
			}
			if n == 0 {
				j += 2
			} else {
				j = m + 1
				if t.GetCardValue(cardData[j]) != t.GetCardValue(cardData[j+1]) {
					j++
					n = 4
				}
			}
			if t.GetCardValue(cardData[i]) > t.GetCardValue(cardData1[j]) {
				return Win
			} else if t.GetCardValue(cardData[j]) < t.GetCardValue(cardData1[j]) {
				return Lose
			} else {
				if t.GetCardValue(cardData[m]) > t.GetCardValue(cardData1[n]) {
					return Win
				} else if t.GetCardValue(cardData[m]) < t.GetCardValue(cardData1[n]) {
					return Lose
				} else {
					return Tied
				}
			}
		}
	} else if firstType == One_pair {
		i := 0
		j := 0
		for ; i != 5; i++ {
			if t.GetCardValue(cardData[i]) == t.GetCardValue(cardData[i+1]) {
				break
			}
		}
		for ; j != 5; j++ {
			if t.GetCardValue(cardData1[j]) == t.GetCardValue(cardData1[j+1]) {
				break
			}
		}
		// 懒得用循环，以后看情况优化(可能存在问题)
		if t.GetCardValue(cardData[i]) > t.GetCardValue(cardData1[j]) {
			return Win
		} else if t.GetCardValue(cardData[i]) < t.GetCardValue(cardData1[j]) {
			return Lose
		} else {
			n := 0
			m := 0
			for n != 5 || m != 5 {
				if i == n {
					n += 2
					continue
				}
				if j == m {
					m += 2
					continue
				}
				if t.GetCardValue(cardData[n]) > t.GetCardValue(cardData1[m]) {
					return Win
				} else if t.GetCardValue(cardData[n]) < t.GetCardValue(cardData1[m]) {
					return Lose
				} else {
					n++
					m++
				}

			}
			return Tied
		}
	} else if firstType == High_card {
		for i := 0; i != 5; i++ {
			if t.GetCardValue(cardData[i]) > t.GetCardValue(cardData1[i]) {
				return Win
			} else if t.GetCardValue(cardData[i]) < t.GetCardValue(cardData1[i]) {
				return Lose
			} else {
				if i == 4 {
					return Tied
				}
			}
		}
	}
	return Tied
}

// Combine 排列组合选择最优
func (t *Table) Combine(card []int32, n int, realHandCard []int32, count int, chairID int) {
	if n < count {
		return
	}
	c := make([]int32, n)
	copy(c[0:], card[0:])
	for i := 0; i != n; i++ {
		if n == count {
			copy(realHandCard[0:], c[0:])
			sort.Sort(SortByte(realHandCard))
			if t.Players[chairID].CardType == 0 { // 在外面判断椅子号存在
				t.Players[chairID].RealHandPoker = append(t.Players[chairID].RealHandPoker, realHandCard...)
				t.Players[chairID].CardType = t.GetCardType(t.Players[chairID].RealHandPoker, int16(count))
			} else {
				if t.CompareCard(realHandCard, t.Players[chairID].RealHandPoker, int16(count)) == 1 {
					t.Players[chairID].RealHandPoker = t.Players[chairID].RealHandPoker[:0] // 置空
					t.Players[chairID].RealHandPoker = append(t.Players[chairID].RealHandPoker, realHandCard...)
					t.Players[chairID].CardType = t.GetCardType(t.Players[chairID].RealHandPoker, int16(count))
				}
			}
			return
		}
		if n > count {
			if i+1 < len(c) {
				c = append(c[:i], c[i+1:]...)
			} else {
				copy(c[0:], c[0:(i-1)])
			}
			t.Combine(c, n-1, realHandCard, count, chairID)
			c = make([]int32, n)
			copy(c[0:], card[0:]) // 重置c的状态
		}
	}
}

// FlopHandNotifyBefore 翻牌前下注通知
func (t *Table) FlopHandNotifyBefore(chairID int) {
	if t.LastHandleBet == 0 { // 允许让牌
		// 让牌标志 = true
	}
	// 获取玩家底分 用于发送客户端
}

// TurnHandNotifyBefore 转牌前下注通知
func (t *Table) TurnHandNotifyBefore(chairID int) {
	if t.LastHandleBet == 0 { // 允许让牌
		// 让牌标志 = true
	}
}

// RiverHandNotifyBefore 河牌前下注通知
func (t *Table) RiverHandNotifyBefore(chairID int) {
	if t.LastHandleBet == 0 { // 允许让牌
		// 让牌标志 = true
	}
}

// RiverHandNotify 河牌后下注通知
func (t *Table) RiverHandNotify(chairID int) {
	if t.LastHandleBet == 0 { // 允许让牌
		// 让牌标志 = true
	}
}

// FlopHandBetBefore 翻牌前下注 在table内判断， 在table_frame_sink内验证判断并用于发送消息
func (t *Table) FlopHandBetBefore(chairID int) {
	for i := 0; i != t.ChairNumber; i++ { // 判断是否只剩自己
		if t.Players[i].PlayerStatus == true {
			if i == chairID {
				continue
			}

			if t.Players[i].Fold == false { // 有人没弃牌，退出
				break
			}

		}
		if i == t.ChairNumber-1 {
			t.FlopSettle = true // 单人结算
			return
		}
	}

}

// TurnHandBetBefore 转牌前下注
func (t *Table) TurnHandBetBefore(chairID int) {
	for i := 0; i != t.ChairNumber; i++ { // 判断是否只剩自己
		if t.Players[i].PlayerStatus == true {
			if i == chairID {
				continue
			}

			if t.Players[i].Fold == false { // 有人没弃牌，退出
				break
			}

		}
		if i == t.ChairNumber-1 {
			t.FlopSettle = true // 单人结算
			return
		}
	}
}

// RiverHandBetBefore 河牌前下注
func (t *Table) RiverHandBetBefore(chairID int) {
	for i := 0; i != t.ChairNumber; i++ { // 判断是否只剩自己
		if t.Players[i].PlayerStatus == true {
			if i == chairID {
				continue
			}

			if t.Players[i].Fold == false { // 有人没弃牌，退出
				break
			}

		}
		if i == t.ChairNumber-1 {
			t.FlopSettle = true // 单人结算
			return
		}
	}
}

// RiverHandBet 河牌后下注
func (t *Table) RiverHandBet(chairID int) {
	for i := 0; i != t.ChairNumber; i++ { // 判断是否只剩自己
		if t.Players[i].PlayerStatus == true {
			if i == chairID {
				continue
			}

			if t.Players[i].Fold == false { // 有人没弃牌，退出
				break
			}

		}
		if i == t.ChairNumber-1 {
			t.FlopSettle = true // 单人结算
			return
		}
	}
}

// NewGameTable 测试使用
func NewGameTable() *Table {
	t := &Table{}
	t.Init()
	return t
}
