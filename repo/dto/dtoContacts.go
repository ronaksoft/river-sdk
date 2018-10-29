package dto

import "git.ronaksoftware.com/ronak/riversdk/msg"

type Contacts struct {
	dto
	ID         int64  `gorm:"primary_key;column:ID;auto_increment:false" json:"ID"`
	FirstName  string `gorm:"type:TEXT;column:FirstName" json:"FirstName"`
	LastName   string `gorm:"type:TEXT;column:LastName" json:"LastName"`
	AccessHash int64  `gorm:"column:AccessHash" json:"AccessHash"`
	Phone      string `gorm:"type:TEXT;column:Phone" json:"Phone"`
	Username   string `gorm:"type:TEXT;column:Username" json:"Username"`
	ClientID   int64  `gorm:"column:ClientID" json:"ClientID"`
}

func (Contacts) TableName() string {
	return "users"
}

func (u *Contacts) MapFromContactUser(t *msg.ContactUser) {
	u.ID = t.ID
	u.FirstName = t.FirstName
	u.LastName = t.LastName
	u.AccessHash = int64(t.AccessHash)
	u.Phone = t.Phone
	u.Username = t.Username
	u.ClientID = t.ClientID
}

func (u *Contacts) MapToUser(v *msg.User) {
	v.ID = u.ID
	v.FirstName = u.FirstName
	v.LastName = u.LastName
}
func (u *Contacts) MapToContactUser(v *msg.ContactUser) {
	v.ID = u.ID
	v.FirstName = u.FirstName
	v.LastName = u.LastName
	v.AccessHash = uint64(u.AccessHash)
	v.Phone = u.Phone
	v.Username = u.Username
	v.ClientID = u.ClientID
}

func (u *Contacts) MapToPhoneContact(v *msg.PhoneContact) {
	v.ClientID = u.ClientID
	v.FirstName = u.FirstName
	v.LastName = u.LastName
	v.Phone = u.Phone
}
