package msg

import (
	"sync"
)

var (
	PoolMessagesSendMedia = sync.Pool{
		New: func() interface{} {
			m := new(MessagesSendMedia)
			return m
		},
	}
	PoolPeer = sync.Pool{
		New: func() interface{} {
			m := new(Peer)
			return m
		},
	}
	PoolUpdateUserTyping = sync.Pool{
		New: func() interface{} {
			m := new(UpdateUserTyping)
			return m
		},
	}
	PoolGroupFull = sync.Pool{
		New: func() interface{} {
			m := new(GroupFull)
			return m
		},
	}
	PoolUpdateGroupTitleUpdated = sync.Pool{
		New: func() interface{} {
			m := new(UpdateGroupTitleUpdated)
			return m
		},
	}
	PoolGroupsAddUser = sync.Pool{
		New: func() interface{} {
			m := new(GroupsAddUser)
			return m
		},
	}
	PoolAck = sync.Pool{
		New: func() interface{} {
			m := new(Ack)
			return m
		},
	}
	PoolContactUser = sync.Pool{
		New: func() interface{} {
			m := new(ContactUser)
			return m
		},
	}
	PoolAccountGetNotifySettings = sync.Pool{
		New: func() interface{} {
			m := new(AccountGetNotifySettings)
			return m
		},
	}
	PoolUpdateReadHistoryOutbox = sync.Pool{
		New: func() interface{} {
			m := new(UpdateReadHistoryOutbox)
			return m
		},
	}
	PoolMessageEnvelope = sync.Pool{
		New: func() interface{} {
			m := new(MessageEnvelope)
			return m
		},
	}
	PoolUpdateGetDifference = sync.Pool{
		New: func() interface{} {
			m := new(UpdateGetDifference)
			return m
		},
	}
	PoolInitAuthCompleted = sync.Pool{
		New: func() interface{} {
			m := new(InitAuthCompleted)
			return m
		},
	}
	PoolUpdateContainer = sync.Pool{
		New: func() interface{} {
			m := new(UpdateContainer)
			return m
		},
	}
	PoolFile = sync.Pool{
		New: func() interface{} {
			m := new(File)
			return m
		},
	}
	PoolUser = sync.Pool{
		New: func() interface{} {
			m := new(User)
			return m
		},
	}
	PoolUsersMany = sync.Pool{
		New: func() interface{} {
			m := new(UsersMany)
			return m
		},
	}
	PoolAccountRegisterDevice = sync.Pool{
		New: func() interface{} {
			m := new(AccountRegisterDevice)
			return m
		},
	}
	PoolAuthLogout = sync.Pool{
		New: func() interface{} {
			m := new(AuthLogout)
			return m
		},
	}
	PoolUsersGet = sync.Pool{
		New: func() interface{} {
			m := new(UsersGet)
			return m
		},
	}
	PoolRSAPublicKey = sync.Pool{
		New: func() interface{} {
			m := new(RSAPublicKey)
			return m
		},
	}
	PoolMessagesGetDialog = sync.Pool{
		New: func() interface{} {
			m := new(MessagesGetDialog)
			return m
		},
	}
	PoolDialog = sync.Pool{
		New: func() interface{} {
			m := new(Dialog)
			return m
		},
	}
	PoolAuthAuthorization = sync.Pool{
		New: func() interface{} {
			m := new(AuthAuthorization)
			return m
		},
	}
	PoolAuthRecall = sync.Pool{
		New: func() interface{} {
			m := new(AuthRecall)
			return m
		},
	}
	PoolSystemGetPublicKeys = sync.Pool{
		New: func() interface{} {
			m := new(SystemGetPublicKeys)
			return m
		},
	}
	PoolGroupsCreate = sync.Pool{
		New: func() interface{} {
			m := new(GroupsCreate)
			return m
		},
	}
	PoolMessagesReadHistory = sync.Pool{
		New: func() interface{} {
			m := new(MessagesReadHistory)
			return m
		},
	}
	PoolUpdateGroupMemberAdded = sync.Pool{
		New: func() interface{} {
			m := new(UpdateGroupMemberAdded)
			return m
		},
	}
	PoolContactsGet = sync.Pool{
		New: func() interface{} {
			m := new(ContactsGet)
			return m
		},
	}
	PoolMessagesGetDialogs = sync.Pool{
		New: func() interface{} {
			m := new(MessagesGetDialogs)
			return m
		},
	}
	PoolUpdateGetState = sync.Pool{
		New: func() interface{} {
			m := new(UpdateGetState)
			return m
		},
	}
	PoolAccountUpdateUsername = sync.Pool{
		New: func() interface{} {
			m := new(AccountUpdateUsername)
			return m
		},
	}
	PoolAccountCheckUsername = sync.Pool{
		New: func() interface{} {
			m := new(AccountCheckUsername)
			return m
		},
	}
	PoolUpdateReadHistoryInbox = sync.Pool{
		New: func() interface{} {
			m := new(UpdateReadHistoryInbox)
			return m
		},
	}
	PoolMessagesSetTyping = sync.Pool{
		New: func() interface{} {
			m := new(MessagesSetTyping)
			return m
		},
	}
	PoolInitCompleteAuth = sync.Pool{
		New: func() interface{} {
			m := new(InitCompleteAuth)
			return m
		},
	}
	PoolUserMessage = sync.Pool{
		New: func() interface{} {
			m := new(UserMessage)
			return m
		},
	}
	PoolMessagesMany = sync.Pool{
		New: func() interface{} {
			m := new(MessagesMany)
			return m
		},
	}
	PoolUpdateDifference = sync.Pool{
		New: func() interface{} {
			m := new(UpdateDifference)
			return m
		},
	}
	PoolContactsDelete = sync.Pool{
		New: func() interface{} {
			m := new(ContactsDelete)
			return m
		},
	}
	PoolSystemGetDHGroups = sync.Pool{
		New: func() interface{} {
			m := new(SystemGetDHGroups)
			return m
		},
	}
	PoolUpdateMessageEdited = sync.Pool{
		New: func() interface{} {
			m := new(UpdateMessageEdited)
			return m
		},
	}
	PoolUpdateState = sync.Pool{
		New: func() interface{} {
			m := new(UpdateState)
			return m
		},
	}
	PoolMessageContainer = sync.Pool{
		New: func() interface{} {
			m := new(MessageContainer)
			return m
		},
	}
	PoolAccountSetNotifySettings = sync.Pool{
		New: func() interface{} {
			m := new(AccountSetNotifySettings)
			return m
		},
	}
	PoolUpdateMessageID = sync.Pool{
		New: func() interface{} {
			m := new(UpdateMessageID)
			return m
		},
	}
	PoolMessagesGet = sync.Pool{
		New: func() interface{} {
			m := new(MessagesGet)
			return m
		},
	}
	PoolContactsImported = sync.Pool{
		New: func() interface{} {
			m := new(ContactsImported)
			return m
		},
	}
	PoolClientPendingMessage = sync.Pool{
		New: func() interface{} {
			m := new(ClientPendingMessage)
			return m
		},
	}
	PoolProtoMessage = sync.Pool{
		New: func() interface{} {
			m := new(ProtoMessage)
			return m
		},
	}
	PoolAuthRegister = sync.Pool{
		New: func() interface{} {
			m := new(AuthRegister)
			return m
		},
	}
	PoolAuthCheckedPhone = sync.Pool{
		New: func() interface{} {
			m := new(AuthCheckedPhone)
			return m
		},
	}
	PoolSystemClientLog = sync.Pool{
		New: func() interface{} {
			m := new(SystemClientLog)
			return m
		},
	}
	PoolInitCompleteAuthInternal = sync.Pool{
		New: func() interface{} {
			m := new(InitCompleteAuthInternal)
			return m
		},
	}
	PoolUpdateEnvelope = sync.Pool{
		New: func() interface{} {
			m := new(UpdateEnvelope)
			return m
		},
	}
	PoolAuthSentCode = sync.Pool{
		New: func() interface{} {
			m := new(AuthSentCode)
			return m
		},
	}
	PoolFileLocation = sync.Pool{
		New: func() interface{} {
			m := new(FileLocation)
			return m
		},
	}
	PoolMessagesEdit = sync.Pool{
		New: func() interface{} {
			m := new(MessagesEdit)
			return m
		},
	}
	PoolGroupsEditTitle = sync.Pool{
		New: func() interface{} {
			m := new(GroupsEditTitle)
			return m
		},
	}
	PoolAuthLogin = sync.Pool{
		New: func() interface{} {
			m := new(AuthLogin)
			return m
		},
	}
	PoolError = sync.Pool{
		New: func() interface{} {
			m := new(Error)
			return m
		},
	}
	PoolProtoEncryptedPayload = sync.Pool{
		New: func() interface{} {
			m := new(ProtoEncryptedPayload)
			return m
		},
	}
	PoolPhoneContact = sync.Pool{
		New: func() interface{} {
			m := new(PhoneContact)
			return m
		},
	}
	PoolUpdateUserStatus = sync.Pool{
		New: func() interface{} {
			m := new(UpdateUserStatus)
			return m
		},
	}
	PoolSystemPublicKeys = sync.Pool{
		New: func() interface{} {
			m := new(SystemPublicKeys)
			return m
		},
	}
	PoolDHGroup = sync.Pool{
		New: func() interface{} {
			m := new(DHGroup)
			return m
		},
	}
	PoolInitDB = sync.Pool{
		New: func() interface{} {
			m := new(InitDB)
			return m
		},
	}
	PoolEchoWithDelay = sync.Pool{
		New: func() interface{} {
			m := new(EchoWithDelay)
			return m
		},
	}
	PoolGroup = sync.Pool{
		New: func() interface{} {
			m := new(Group)
			return m
		},
	}
	PoolSystemDHGroups = sync.Pool{
		New: func() interface{} {
			m := new(SystemDHGroups)
			return m
		},
	}
	PoolMessagesSent = sync.Pool{
		New: func() interface{} {
			m := new(MessagesSent)
			return m
		},
	}
	PoolGroupsGetFull = sync.Pool{
		New: func() interface{} {
			m := new(GroupsGetFull)
			return m
		},
	}
	PoolMessagesSend = sync.Pool{
		New: func() interface{} {
			m := new(MessagesSend)
			return m
		},
	}
	PoolGroupsDeleteUser = sync.Pool{
		New: func() interface{} {
			m := new(GroupsDeleteUser)
			return m
		},
	}
	PoolUpdateNotifySettings = sync.Pool{
		New: func() interface{} {
			m := new(UpdateNotifySettings)
			return m
		},
	}
	PoolAuthRecalled = sync.Pool{
		New: func() interface{} {
			m := new(AuthRecalled)
			return m
		},
	}
	PoolMessagesDialogs = sync.Pool{
		New: func() interface{} {
			m := new(MessagesDialogs)
			return m
		},
	}
	PoolInputPeer = sync.Pool{
		New: func() interface{} {
			m := new(InputPeer)
			return m
		},
	}
	PoolMessagesGetHistory = sync.Pool{
		New: func() interface{} {
			m := new(MessagesGetHistory)
			return m
		},
	}
	PoolUpdateNewMessage = sync.Pool{
		New: func() interface{} {
			m := new(UpdateNewMessage)
			return m
		},
	}
	PoolContactsImport = sync.Pool{
		New: func() interface{} {
			m := new(ContactsImport)
			return m
		},
	}
	PoolPeerNotifySettings = sync.Pool{
		New: func() interface{} {
			m := new(PeerNotifySettings)
			return m
		},
	}
	PoolAccountUpdateProfile = sync.Pool{
		New: func() interface{} {
			m := new(AccountUpdateProfile)
			return m
		},
	}
	PoolFileSavePart = sync.Pool{
		New: func() interface{} {
			m := new(FileSavePart)
			return m
		},
	}
	PoolUpdateGroupMemberDeleted = sync.Pool{
		New: func() interface{} {
			m := new(UpdateGroupMemberDeleted)
			return m
		},
	}
	PoolInputUser = sync.Pool{
		New: func() interface{} {
			m := new(InputUser)
			return m
		},
	}
	PoolClientPendingMessageDelivery = sync.Pool{
		New: func() interface{} {
			m := new(ClientPendingMessageDelivery)
			return m
		},
	}
	PoolInputFile = sync.Pool{
		New: func() interface{} {
			m := new(InputFile)
			return m
		},
	}
	PoolContactsMany = sync.Pool{
		New: func() interface{} {
			m := new(ContactsMany)
			return m
		},
	}
	PoolAccountUnregisterDevice = sync.Pool{
		New: func() interface{} {
			m := new(AccountUnregisterDevice)
			return m
		},
	}
	PoolAuthSendCode = sync.Pool{
		New: func() interface{} {
			m := new(AuthSendCode)
			return m
		},
	}
	PoolGroupParticipant = sync.Pool{
		New: func() interface{} {
			m := new(GroupParticipant)
			return m
		},
	}
	PoolBool = sync.Pool{
		New: func() interface{} {
			m := new(Bool)
			return m
		},
	}
	PoolInitResponse = sync.Pool{
		New: func() interface{} {
			m := new(InitResponse)
			return m
		},
	}
	PoolAuthCheckPhone = sync.Pool{
		New: func() interface{} {
			m := new(AuthCheckPhone)
			return m
		},
	}
	PoolInitConnect = sync.Pool{
		New: func() interface{} {
			m := new(InitConnect)
			return m
		},
	}
	PoolFileGet = sync.Pool{
		New: func() interface{} {
			m := new(FileGet)
			return m
		},
	}
	PoolUpdateUsername = sync.Pool{
		New: func() interface{} {
			m := new(UpdateUsername)
			return m
		},
	}
)
