package riversdk

import messageHole "git.ronaksoftware.com/ronak/riversdk/pkg/message_hole"

/*
   Creation Time: 2019 - Jul - 09
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

type MessageHole struct {
	lists map[string]*messageHole.HoleManager
}

func (mh *MessageHole) Init() {
	mh.lists = make(map[string]*messageHole.HoleManager)
}

func (mh *MessageHole) getManager(key string) *messageHole.HoleManager {
	hm, ok := mh.lists[key]
	if !ok {
		hm = &messageHole.HoleManager{}
		mh.lists[key] = hm
	}
	return hm
}

func (mh *MessageHole) InsertHole(key string, min, max int64) {
	m := mh.getManager(key)
	m.InsertBar(messageHole.Bar{Min: min, Max: max, Type: messageHole.Hole})
}

func (mh *MessageHole) InsertFill(key string, min, max int64) {
	m := mh.getManager(key)
	m.InsertBar(messageHole.Bar{Min: min, Max: max, Type: messageHole.Filled})
}

func (mh *MessageHole) IsInHole(key string, msgID int64) bool {
	m := mh.getManager(key)
	return m.IsPointHole(msgID)
}

func (mh *MessageHole) IsRangeFilled(key string, min, max int64) bool {
	m := mh.getManager(key)
	return m.IsRangeFilled(min, max)
}

func (mh *MessageHole) SetLowerFilled(key string) {
	m := mh.getManager(key)
	m.SetLowerFilled()
}

func (mh *MessageHole) SetUpperFilled(key string, maxID int64) {
	m := mh.getManager(key)
	m.SetUpperFilled(maxID)
}

func (mh *MessageHole) String(key string) string {
	m := mh.getManager(key)
	return m.String()
}

func (mh *MessageHole) Release(key string) {
	delete(mh.lists, key)
}
