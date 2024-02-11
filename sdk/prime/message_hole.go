package riversdk

import (
    "github.com/ronaksoft/river-sdk/internal/hole"
)

/*
   Creation Time: 2019 - Jul - 09
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

type MessageHole struct {
    lists map[string]*hole.Detector
}

func (mh *MessageHole) Init() {
    mh.lists = make(map[string]*hole.Detector)
}

func (mh *MessageHole) detector(key string) *hole.Detector {
    hm, ok := mh.lists[key]
    if !ok {
        hm = &hole.Detector{}
        mh.lists[key] = hm
    }
    return hm
}

func (mh *MessageHole) InsertHole(key string, min, max int64) {
    m := mh.detector(key)
    m.InsertBar(hole.Bar{Min: min, Max: max, Type: hole.Hole})
}

func (mh *MessageHole) InsertFill(key string, min, max int64) {
    m := mh.detector(key)
    m.InsertBar(hole.Bar{Min: min, Max: max, Type: hole.Filled})
}

func (mh *MessageHole) IsInHole(key string, msgID int64) bool {
    m := mh.detector(key)
    return m.IsPointHole(msgID)
}

func (mh *MessageHole) IsRangeFilled(key string, min, max int64) bool {
    m := mh.detector(key)
    return m.IsRangeFilled(min, max)
}

func (mh *MessageHole) SetLowerFilled(key string) {
    m := mh.detector(key)
    m.SetLowerFilled()
}

func (mh *MessageHole) SetUpperFilled(key string, maxID int64) {
    m := mh.detector(key)
    m.SetUpperFilled(maxID)
}

func (mh *MessageHole) String(key string) string {
    m := mh.detector(key)
    return m.String()
}

func (mh *MessageHole) Release(key string) {
    delete(mh.lists, key)
}
