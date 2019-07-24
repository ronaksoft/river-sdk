package domain

// RiverConfigurator high level interface to prevent package conflict when pass RiverConfig to SyncController
type RiverConfigurator interface {
	Save()
	ChangeAuthID(authID int64)
	ChangeAuthKey(authKey []byte)
	ChangeUserID(userID int64)
	ChangeUsername(username string)
	ChangePhone(phone string)
	ChangeFirstName(firstName string)
	ChangeLastName(lastName string)
	ChangeBio(lastName string)
	PickupAuthID() int64
	PickupAuthKey() [256]byte
	PickupUserID() int64
	PickupUsername() string
	PickupPhone() string
	PickupFirstName() string
	PickupLastName() string
	PickupBio() string
}
