package dto

import msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"

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
	StatusTime int64  `gorm:"column:StatusTime" json:"StatusTime"`
	Restricted bool   `gorm:"column:Restricted" json:"Restricted"`
	Bio        string `gorm:"type:TEXT;column:Bio" json:"Bio"`
	Photo      []byte `gorm:"type:blob;column:Photo" json:"Photo"`
}

func (Users) TableName() string {
	return "users"
}

func (u *Users) MapFromUser(t *msg.User) {
	u.ID = t.ID
	u.FirstName = t.FirstName
	u.LastName = t.LastName
	u.Username = t.Username
	u.Status = int32(t.Status)
	u.Restricted = t.Restricted
	u.AccessHash = int64(t.AccessHash)
	u.Bio = t.Bio
	if t.Photo != nil {
		if t.Photo.PhotoID != 0 {
			u.Photo, _ = t.Photo.Marshal()
		}
	}
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
	if t.Photo != nil {
		if t.Photo.PhotoID != 0 {
			u.Photo, _ = t.Photo.Marshal()
		}
	}
}

func (u *Users) MapToUser(v *msg.User) {
	v.ID = u.ID
	v.FirstName = u.FirstName
	v.LastName = u.LastName
	v.Username = u.Username
	v.Status = msg.UserStatus(u.Status)
	v.Restricted = u.Restricted
	v.AccessHash = uint64(u.AccessHash)
	v.Bio = u.Bio

	if v.Photo == nil {
		v.Photo = new(msg.UserPhoto)
	}
	err := v.Photo.Unmarshal(u.Photo)
	if err != nil {
		v.Photo = nil
	}
}
func (u *Users) MapToContactUser(v *msg.ContactUser) {
	v.ID = u.ID
	v.FirstName = u.FirstName
	v.LastName = u.LastName
	v.AccessHash = uint64(u.AccessHash)
	v.Phone = u.Phone
	v.Username = u.Username
	v.ClientID = u.ClientID
	if v.Photo == nil {
		v.Photo = new(msg.UserPhoto)
	}
	err := v.Photo.Unmarshal(u.Photo)
	if err != nil {
		v.Photo = nil
	}
}

func (u *Users) MapToPhoneContact(v *msg.PhoneContact) {
	v.ClientID = u.ClientID
	v.FirstName = u.FirstName
	v.LastName = u.LastName
	v.Phone = u.Phone
}
