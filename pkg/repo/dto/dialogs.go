package dto

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
)

type Dialogs struct {
	dto
	PeerID          int64  `gorm:"type:bigint;primary_key;column:PeerID" json:"PeerID"`      // type is required for composite primary key
	PeerType        int32  `gorm:"type:integer;primary_key;column:PeerType" json:"PeerType"` // type is required for composite primary key
	TopMessageID    int64  `gorm:"column:TopMessageID" json:"TopMessageID"`
	AccessHash      int64  `gorm:"column:AccessHash" json:"AccessHash"`
	ReadInboxMaxID  int64  `gorm:"column:ReadInboxMaxID" json:"ReadInboxMaxID"`
	ReadOutboxMaxID int64  `gorm:"column:ReadOutboxMaxID" json:"ReadOutboxMaxID"`
	UnreadCount     int32  `gorm:"column:UnreadCount" json:"UnreadCount"`
	LastUpdate      int64  `gorm:"column:LastUpdate" json:"LastUpdate"`
	NotifyFlags     int32  `gorm:"column:NotifyFlags" json:"NotifyFlags"`
	NotifyMuteUntil int64  `gorm:"column:NotifyMuteUntil" json:"NotifyMuteUntil"`
	NotifySound     string `gorm:"column:NotifySound" json:"NotifySound"`
	MentionedCount  int32  `gorm:"column:MentionedCount" json:"MentionedCount"`
	Pinned          bool   `gorm:"column:IsPinned" json:"IsPinned"`
}

func (Dialogs) TableName() string {
	return "dialogs"
}

func (d *Dialogs) Map(v *msg.Dialog) {
	d.AccessHash = int64(v.AccessHash)
	d.PeerID = v.PeerID
	d.PeerType = v.PeerType
	d.ReadInboxMaxID = v.ReadInboxMaxID
	d.ReadOutboxMaxID = v.ReadOutboxMaxID
	d.TopMessageID = v.TopMessageID
	d.UnreadCount = v.UnreadCount
	d.LastUpdate = time.Now().Unix()
	d.Pinned = v.Pinned

	if v.NotifySettings != nil {
		d.NotifyFlags = v.NotifySettings.Flags
		d.NotifyMuteUntil = v.NotifySettings.MuteUntil
		d.NotifySound = v.NotifySettings.Sound
	}
	d.MentionedCount = v.MentionedCount
}

func (d *Dialogs) MapTo(v *msg.Dialog) {
	v.PeerID = d.PeerID
	v.PeerType = d.PeerType
	v.TopMessageID = d.TopMessageID
	v.ReadInboxMaxID = d.ReadInboxMaxID
	v.ReadOutboxMaxID = d.ReadOutboxMaxID
	v.UnreadCount = d.UnreadCount
	v.AccessHash = uint64(d.AccessHash)
	v.NotifySettings = new(msg.PeerNotifySettings)
	v.NotifySettings.Flags = d.NotifyFlags
	v.NotifySettings.MuteUntil = d.NotifyMuteUntil
	v.NotifySettings.Sound = d.NotifySound
	v.MentionedCount = d.MentionedCount
	v.Pinned = d.Pinned
}
