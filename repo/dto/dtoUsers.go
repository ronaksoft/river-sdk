package dto

import "git.ronaksoftware.com/ronak/riversdk/msg"

type Users struct {
	dto
	ID         int64  `gorm:"primary_key;column:ID;auto_increment:false" json:"ID"`
	FirstName  string `gorm:"type:TEXT;column:FirstName" json:"FirstName"`
	LastName   string `gorm:"type:TEXT;column:LastName" json:"LastName"`
	AccessHash int64  `gorm:"column:AccessHash" json:"AccessHash"`
	Phone      string `gorm:"type:TEXT;column:Phone" json:"Phone"`
	Username   string `gorm:"type:TEXT;column:Username" json:"Username"`
	ClientID   int64  `gorm:"column:ClientID" json:"ClientID"`
	IsContact  int32  `gorm:"column:IsContact" json:"IsContact"`
	Status     int32  `gorm:"column:Status" json:"Status"`
	Restricted bool   `gorm:"column:Restricted" json:"Restricted"`
}

func (Users) TableName() string {
	return "users"
}

func (u *Users) MapFromUser(t *msg.User) {
	u.ID = t.ID
	u.FirstName = t.FirstName
	u.LastName = t.LastName
	u.Username = t.Username
}

func (u *Users) MapFromContactUser(t *msg.ContactUser) {
	u.ID = t.ID
	u.FirstName = t.FirstName
	u.LastName = t.LastName
	u.AccessHash = int64(t.AccessHash)
	u.Phone = t.Phone
	u.Username = t.Username
	u.ClientID = t.ClientID
	u.IsContact = 1
}

func (u *Users) MapToUser(v *msg.User) {
	v.ID = u.ID
	v.FirstName = u.FirstName
	v.LastName = u.LastName
	v.Username = u.Username
}
func (u *Users) MapToContactUser(v *msg.ContactUser) {
	v.ID = u.ID
	v.FirstName = u.FirstName
	v.LastName = u.LastName
	v.AccessHash = uint64(u.AccessHash)
	v.Phone = u.Phone
	v.Username = u.Username
	v.ClientID = u.ClientID
}

func (u *Users) MapToPhoneContact(v *msg.PhoneContact) {
	v.ClientID = u.ClientID
	v.FirstName = u.FirstName
	v.LastName = u.LastName
	v.Phone = u.Phone
}
