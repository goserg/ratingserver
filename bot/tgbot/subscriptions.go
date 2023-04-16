package tgbot

import (
	botmodel "ratingserver/bot/model"

	mapset "github.com/deckarep/golang-set/v2"
)

type subscriptions struct {
	m map[botmodel.EventType]mapset.Set[int]
}

func newSubs() subscriptions {
	m := make(map[botmodel.EventType]mapset.Set[int])
	return subscriptions{
		m: m,
	}
}

func (s *subscriptions) Add(t botmodel.EventType, userID int) {
	if s.m[t] == nil {
		s.m[t] = mapset.NewSet[int]()
	}
	s.m[t].Add(userID)
}

func (s *subscriptions) Remove(t botmodel.EventType, userID int) {
	if s.m[t] == nil {
		return
	}
	s.m[t].Remove(userID)
}

func (s *subscriptions) GetUserIDs(t botmodel.EventType) []int {
	if s.m[t] == nil {
		return nil
	}
	return s.m[t].ToSlice()
}
