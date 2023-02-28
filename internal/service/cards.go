package service

import "math/rand"

const (
	TotalCardsCnt uint8 = 24
	TotalPlayers  uint8 = 4
	HandCardCount uint8 = 6
)

type Card struct {
	Head uint8
	Tail uint8
}

var InvalidCard = Card{
	Head: 0,
	Tail: 0,
}

var JinPing = Card{
	Head: 6,
	Tail: 4,
}

var HongBa = Card{
	Head: 4,
	Tail: 4,
}

var AllCards = initAllCards()

func initAllCards() []Card {
	arr := []Card{
		{
			Head: 6,
			Tail: 6,
		},
		{
			Head: 6,
			Tail: 5,
		},
		{
			Head: 5,
			Tail: 5,
		},
		{
			Head: 6,
			Tail: 4,
		},
		{
			Head: 4,
			Tail: 4,
		},
		{
			Head: 3,
			Tail: 3,
		},
		{
			Head: 6,
			Tail: 1,
		},
		{
			Head: 5,
			Tail: 1,
		},
		{
			Head: 3,
			Tail: 1,
		},
		{
			Head: 2,
			Tail: 2,
		},
		{
			Head: 1,
			Tail: 1,
		},
		{
			Head: 6,
			Tail: 6,
		},
		{
			Head: 6,
			Tail: 5,
		},
		{
			Head: 5,
			Tail: 5,
		},
		{
			Head: 6,
			Tail: 4,
		},
		{
			Head: 4,
			Tail: 4,
		},
		{
			Head: 3,
			Tail: 3,
		},
		{
			Head: 6,
			Tail: 1,
		},
		{
			Head: 5,
			Tail: 1,
		},
		{
			Head: 3,
			Tail: 1,
		},
		{
			Head: 2,
			Tail: 2,
		},
		{
			Head: 1,
			Tail: 1,
		},
		{
			Head: 6,
			Tail: 3,
		},
		{
			Head: 6,
			Tail: 2,
		},
	}
	return arr
}

func Shuffle() []uint8 {
	res := make([]uint8, TotalCardsCnt)
	for i := range res {
		res[i] = uint8(i)
	}
	rand.Shuffle(int(TotalCardsCnt), func(i, j int) {
		res[i], res[j] = res[j], res[i]
	})
	return res
}

func (c Card) GetCount() uint8 {
	return c.Tail + c.Head
}

// false 横牌， true 竖牌
func (c Card) isStanding() bool {
	return c.Tail != c.Head
}
