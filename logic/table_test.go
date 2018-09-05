package logic

import (
	"fmt"
	"sort"
	"testing"
)

func Test_GetCardType(t *testing.T) {
	g := NewGameTable()

	cardData := []int32{0x0a, 0x0b, 0x0c, 0x0d, 0x0e}
	sort.Sort(SortByte(cardData))
	fmt.Printf("%x\n", cardData)
	cardType := g.GetCardType(cardData, 5)
	fmt.Println("牌型", cardType)
}

func Test_CompareCard(t *testing.T) {
	g := NewGameTable()

	firstData := []int32{0x15, 0x05, 0x2e, 0x0e, 0x09}
	nextData := []int32{0x1e, 0xea, 0x25, 0x35, 0x28}
	sort.Sort(SortByte(firstData))
	sort.Sort(SortByte(nextData))
	firstType := g.GetCardType(firstData, 5)
	nextType := g.GetCardType(nextData, 5)
	fmt.Println("牌型1：", firstType, "牌型2：", nextType)
	v := g.CompareCard(firstData, nextData, 5)
	if v == 1 {
		fmt.Println("first赢")
	} else if v == 2 {
		fmt.Println("first输")
	} else {
		fmt.Println("平局")
	}
}

func Test_Combine(t *testing.T) {
	g := NewGameTable()
	firstData := []int32{0x15, 0x05, 0x2e, 0x0e, 0x19, 0x0a, 0x0b}
	g.Players[2].HandPoker = append(g.Players[2].HandPoker, firstData...)
	b := make([]int32, 5)
	g.Combine(g.Players[2].HandPoker, len(g.Players[2].HandPoker), b, 5, 2)
	g.Players[2].PlayerStatus = true
	fmt.Printf("最大牌型%v\n", g.Players[2].CardType)
	fmt.Printf("最大牌型数据%x\n", g.Players[2].RealHandPoker)
	fmt.Printf("原始牌型数据%x\n", g.Players[2].HandPoker) // 原始数据没有被改变
}

func Test_EnsureChairID(t *testing.T) {
	g := NewGameTable()
	g.Players = make([]Player, 6)
	g.ChairNumber = 6
	for i := 0; i != g.ChairNumber; i++ {
		//if i == 5 || i == 0 {
		g.Players[i].PlayerStatus = true
		//}
	}
	fmt.Println(g.EnsureChairID(12))
	fmt.Println(g.EnsureChairID(-1))
}
