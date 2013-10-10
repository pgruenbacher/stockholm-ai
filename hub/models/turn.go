package models

import (
	"appengine/datastore"
	"common"
	"fmt"
	"sort"
	"time"
)

const (
	TurnKind = "Turn"
)

func countTurnsKeyForParent(k interface{}) string {
	return fmt.Sprintf("Turns{Count,Parent:%v}", k)
}

func turnsKeyForParent(k interface{}) string {
	return fmt.Sprintf("Turns{Parent:%v}", k)
}

func latestTurnKeyForParent(k interface{}) string {
	return fmt.Sprintf("Turns{Latest,Parent:%v}", k)
}

type Turns []Turn

func (self Turns) Len() int {
	return len(self)
}

func (self Turns) Less(i, j int) bool {
	return self[i].Ordinal < self[j].Ordinal
}

func (self Turns) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self Turns) process(c common.Context) Turns {
	for index, _ := range self {
		(&self[index]).process(c)
	}
	return self
}

type Turn struct {
	Id              *datastore.Key
	Ordinal         int
	SerializedState []byte `json:"-"`
	State           State  `datastore:"-"`
	CreatedAt       time.Time
}

func (self *Turn) Next(c common.Context, orders Orders) *Turn {
	cpy := *self
	cpy.Id = nil
	cpy.Ordinal += 1
	(&cpy.State).Next(c, orders)
	return &cpy
}

func (self *Turn) process(c common.Context) *Turn {
	if len(self.SerializedState) > 0 {
		common.MustUnmarshalJSON(self.SerializedState, &self.State)
	}
	return self
}

func findTurnsByParent(c common.Context, parent *datastore.Key) (result Turns) {
	ids, err := datastore.NewQuery(TurnKind).Ancestor(parent).GetAll(c, &result)
	common.AssertOkError(err)
	for index, id := range ids {
		result[index].Id = id
	}
	if result == nil {
		result = Turns{}
	}
	return
}

func GetTurnsByParent(c common.Context, parent *datastore.Key) (result Turns) {
	common.Memoize(c, turnsKeyForParent(parent), &result, func() interface{} {
		return findTurnsByParent(c, parent)
	})
	sort.Sort(result)
	return result.process(c)
}

func findLatestTurnByParent(c common.Context, parent *datastore.Key) *Turn {
	var turns Turns
	ids, err := datastore.NewQuery(TurnKind).Ancestor(parent).Order("-CreatedAt").Limit(1).GetAll(c, &turns)
	common.AssertOkError(err)
	for index, id := range ids {
		turns[index].Id = id
	}
	if len(turns) == 0 {
		return nil
	}
	return &turns[0]
}

func GetLatestTurnByParent(c common.Context, parent *datastore.Key) *Turn {
	var result Turn
	if common.Memoize(c, latestTurnKeyForParent(parent), &result, func() interface{} {
		return findLatestTurnByParent(c, parent)
	}) {
		return (&result).process(c)
	}
	return nil
}

func countTurnsByParent(c common.Context, parent *datastore.Key) (result int) {
	var err error
	result, err = datastore.NewQuery(TurnKind).Ancestor(parent).Count(c)
	common.AssertOkError(err)
	return
}

func CountTurnsByParent(c common.Context, parent *datastore.Key) (result int) {
	common.Memoize(c, countTurnsKeyForParent(parent), &result, func() interface{} {
		return countTurnsByParent(c, parent)
	})
	return
}

func (self *Turn) Save(c common.Context, parent *datastore.Key) *Turn {
	self.SerializedState = common.MustMarshalJSON(self.State)
	var err error
	if self.Id == nil {
		self.CreatedAt = time.Now()
		count := CountTurnsByParent(c, parent)
		self.Id, err = datastore.Put(c, datastore.NewKey(c, TurnKind, "", 0, parent), self)
		common.AssertOkError(err)
		common.MemPut(c, latestTurnKeyForParent(parent), self)
		common.MemPut(c, countTurnsKeyForParent(parent), count+1)
	} else {
		_, err = datastore.Put(c, self.Id, self)
		common.AssertOkError(err)
	}
	common.MemDel(c, turnsKeyForParent(parent))
	return self
}