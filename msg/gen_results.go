package msg

import (
	"github.com/gobwas/pool/pbytes"
)

func ResultMessagesSendMedia(out *MessageEnvelope, res *MessagesSendMedia) {
	out.Constructor = C_MessagesSendMedia
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultPeer(out *MessageEnvelope, res *Peer) {
	out.Constructor = C_Peer
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInitUserBound(out *MessageEnvelope, res *InitUserBound) {
	out.Constructor = C_InitUserBound
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateUserTyping(out *MessageEnvelope, res *UpdateUserTyping) {
	out.Constructor = C_UpdateUserTyping
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultGroupFull(out *MessageEnvelope, res *GroupFull) {
	out.Constructor = C_GroupFull
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultGroupsAddUser(out *MessageEnvelope, res *GroupsAddUser) {
	out.Constructor = C_GroupsAddUser
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAck(out *MessageEnvelope, res *Ack) {
	out.Constructor = C_Ack
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultContactUser(out *MessageEnvelope, res *ContactUser) {
	out.Constructor = C_ContactUser
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAccountGetNotifySettings(out *MessageEnvelope, res *AccountGetNotifySettings) {
	out.Constructor = C_AccountGetNotifySettings
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateReadHistoryOutbox(out *MessageEnvelope, res *UpdateReadHistoryOutbox) {
	out.Constructor = C_UpdateReadHistoryOutbox
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessageEnvelope(out *MessageEnvelope, res *MessageEnvelope) {
	out.Constructor = C_MessageEnvelope
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateGetDifference(out *MessageEnvelope, res *UpdateGetDifference) {
	out.Constructor = C_UpdateGetDifference
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInitAuthCompleted(out *MessageEnvelope, res *InitAuthCompleted) {
	out.Constructor = C_InitAuthCompleted
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateContainer(out *MessageEnvelope, res *UpdateContainer) {
	out.Constructor = C_UpdateContainer
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateMessagesDeleted(out *MessageEnvelope, res *UpdateMessagesDeleted) {
	out.Constructor = C_UpdateMessagesDeleted
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateGroupAdmins(out *MessageEnvelope, res *UpdateGroupAdmins) {
	out.Constructor = C_UpdateGroupAdmins
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultFile(out *MessageEnvelope, res *File) {
	out.Constructor = C_File
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUser(out *MessageEnvelope, res *User) {
	out.Constructor = C_User
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUsersMany(out *MessageEnvelope, res *UsersMany) {
	out.Constructor = C_UsersMany
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInputMediaUploadedPhoto(out *MessageEnvelope, res *InputMediaUploadedPhoto) {
	out.Constructor = C_InputMediaUploadedPhoto
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInputMediaUploadedDocument(out *MessageEnvelope, res *InputMediaUploadedDocument) {
	out.Constructor = C_InputMediaUploadedDocument
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAccountRegisterDevice(out *MessageEnvelope, res *AccountRegisterDevice) {
	out.Constructor = C_AccountRegisterDevice
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAuthLogout(out *MessageEnvelope, res *AuthLogout) {
	out.Constructor = C_AuthLogout
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUsersGet(out *MessageEnvelope, res *UsersGet) {
	out.Constructor = C_UsersGet
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultRSAPublicKey(out *MessageEnvelope, res *RSAPublicKey) {
	out.Constructor = C_RSAPublicKey
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessagesGetDialog(out *MessageEnvelope, res *MessagesGetDialog) {
	out.Constructor = C_MessagesGetDialog
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultDialog(out *MessageEnvelope, res *Dialog) {
	out.Constructor = C_Dialog
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAuthAuthorization(out *MessageEnvelope, res *AuthAuthorization) {
	out.Constructor = C_AuthAuthorization
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAuthRecall(out *MessageEnvelope, res *AuthRecall) {
	out.Constructor = C_AuthRecall
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultSystemGetPublicKeys(out *MessageEnvelope, res *SystemGetPublicKeys) {
	out.Constructor = C_SystemGetPublicKeys
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessageActionGroupDeleteUser(out *MessageEnvelope, res *MessageActionGroupDeleteUser) {
	out.Constructor = C_MessageActionGroupDeleteUser
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessageActionClearHistory(out *MessageEnvelope, res *MessageActionClearHistory) {
	out.Constructor = C_MessageActionClearHistory
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultGroupsCreate(out *MessageEnvelope, res *GroupsCreate) {
	out.Constructor = C_GroupsCreate
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessagesReadHistory(out *MessageEnvelope, res *MessagesReadHistory) {
	out.Constructor = C_MessagesReadHistory
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultGroupsUpdateAdmin(out *MessageEnvelope, res *GroupsUpdateAdmin) {
	out.Constructor = C_GroupsUpdateAdmin
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultContactsGet(out *MessageEnvelope, res *ContactsGet) {
	out.Constructor = C_ContactsGet
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessagesGetDialogs(out *MessageEnvelope, res *MessagesGetDialogs) {
	out.Constructor = C_MessagesGetDialogs
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateGetState(out *MessageEnvelope, res *UpdateGetState) {
	out.Constructor = C_UpdateGetState
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAccountUpdateUsername(out *MessageEnvelope, res *AccountUpdateUsername) {
	out.Constructor = C_AccountUpdateUsername
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAccountCheckUsername(out *MessageEnvelope, res *AccountCheckUsername) {
	out.Constructor = C_AccountCheckUsername
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateReadHistoryInbox(out *MessageEnvelope, res *UpdateReadHistoryInbox) {
	out.Constructor = C_UpdateReadHistoryInbox
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInputMediaContact(out *MessageEnvelope, res *InputMediaContact) {
	out.Constructor = C_InputMediaContact
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessagesSetTyping(out *MessageEnvelope, res *MessagesSetTyping) {
	out.Constructor = C_MessagesSetTyping
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultGroupsToggleAdmins(out *MessageEnvelope, res *GroupsToggleAdmins) {
	out.Constructor = C_GroupsToggleAdmins
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInitCompleteAuth(out *MessageEnvelope, res *InitCompleteAuth) {
	out.Constructor = C_InitCompleteAuth
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAccountSetPrivacy(out *MessageEnvelope, res *AccountSetPrivacy) {
	out.Constructor = C_AccountSetPrivacy
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateGroupParticipantAdd(out *MessageEnvelope, res *UpdateGroupParticipantAdd) {
	out.Constructor = C_UpdateGroupParticipantAdd
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUserMessage(out *MessageEnvelope, res *UserMessage) {
	out.Constructor = C_UserMessage
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessagesMany(out *MessageEnvelope, res *MessagesMany) {
	out.Constructor = C_MessagesMany
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateDifference(out *MessageEnvelope, res *UpdateDifference) {
	out.Constructor = C_UpdateDifference
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultContactsDelete(out *MessageEnvelope, res *ContactsDelete) {
	out.Constructor = C_ContactsDelete
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultSystemGetDHGroups(out *MessageEnvelope, res *SystemGetDHGroups) {
	out.Constructor = C_SystemGetDHGroups
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateGroupParticipantAdmin(out *MessageEnvelope, res *UpdateGroupParticipantAdmin) {
	out.Constructor = C_UpdateGroupParticipantAdmin
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateMessageEdited(out *MessageEnvelope, res *UpdateMessageEdited) {
	out.Constructor = C_UpdateMessageEdited
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateState(out *MessageEnvelope, res *UpdateState) {
	out.Constructor = C_UpdateState
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAccountGetPrivacy(out *MessageEnvelope, res *AccountGetPrivacy) {
	out.Constructor = C_AccountGetPrivacy
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInitBindUser(out *MessageEnvelope, res *InitBindUser) {
	out.Constructor = C_InitBindUser
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessageActionGroupAddUser(out *MessageEnvelope, res *MessageActionGroupAddUser) {
	out.Constructor = C_MessageActionGroupAddUser
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessageContainer(out *MessageEnvelope, res *MessageContainer) {
	out.Constructor = C_MessageContainer
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessagesClearHistory(out *MessageEnvelope, res *MessagesClearHistory) {
	out.Constructor = C_MessagesClearHistory
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAccountSetNotifySettings(out *MessageEnvelope, res *AccountSetNotifySettings) {
	out.Constructor = C_AccountSetNotifySettings
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateMessageID(out *MessageEnvelope, res *UpdateMessageID) {
	out.Constructor = C_UpdateMessageID
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessagesGet(out *MessageEnvelope, res *MessagesGet) {
	out.Constructor = C_MessagesGet
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultContactsImported(out *MessageEnvelope, res *ContactsImported) {
	out.Constructor = C_ContactsImported
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultClientPendingMessage(out *MessageEnvelope, res *ClientPendingMessage) {
	out.Constructor = C_ClientPendingMessage
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultProtoMessage(out *MessageEnvelope, res *ProtoMessage) {
	out.Constructor = C_ProtoMessage
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInputMediaPhoto(out *MessageEnvelope, res *InputMediaPhoto) {
	out.Constructor = C_InputMediaPhoto
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAuthRegister(out *MessageEnvelope, res *AuthRegister) {
	out.Constructor = C_AuthRegister
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAuthCheckedPhone(out *MessageEnvelope, res *AuthCheckedPhone) {
	out.Constructor = C_AuthCheckedPhone
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessageActionGroupCreated(out *MessageEnvelope, res *MessageActionGroupCreated) {
	out.Constructor = C_MessageActionGroupCreated
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultSystemClientLog(out *MessageEnvelope, res *SystemClientLog) {
	out.Constructor = C_SystemClientLog
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInputMediaDocument(out *MessageEnvelope, res *InputMediaDocument) {
	out.Constructor = C_InputMediaDocument
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInitCompleteAuthInternal(out *MessageEnvelope, res *InitCompleteAuthInternal) {
	out.Constructor = C_InitCompleteAuthInternal
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateEnvelope(out *MessageEnvelope, res *UpdateEnvelope) {
	out.Constructor = C_UpdateEnvelope
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAuthSentCode(out *MessageEnvelope, res *AuthSentCode) {
	out.Constructor = C_AuthSentCode
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessageActionContactRegistered(out *MessageEnvelope, res *MessageActionContactRegistered) {
	out.Constructor = C_MessageActionContactRegistered
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessageActionGroupTitleChanged(out *MessageEnvelope, res *MessageActionGroupTitleChanged) {
	out.Constructor = C_MessageActionGroupTitleChanged
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultFileLocation(out *MessageEnvelope, res *FileLocation) {
	out.Constructor = C_FileLocation
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateGroupParticipantDeleted(out *MessageEnvelope, res *UpdateGroupParticipantDeleted) {
	out.Constructor = C_UpdateGroupParticipantDeleted
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessagesEdit(out *MessageEnvelope, res *MessagesEdit) {
	out.Constructor = C_MessagesEdit
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultGroupsEditTitle(out *MessageEnvelope, res *GroupsEditTitle) {
	out.Constructor = C_GroupsEditTitle
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAuthLogin(out *MessageEnvelope, res *AuthLogin) {
	out.Constructor = C_AuthLogin
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultError(out *MessageEnvelope, res *Error) {
	out.Constructor = C_Error
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessagesForward(out *MessageEnvelope, res *MessagesForward) {
	out.Constructor = C_MessagesForward
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultProtoEncryptedPayload(out *MessageEnvelope, res *ProtoEncryptedPayload) {
	out.Constructor = C_ProtoEncryptedPayload
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultPhoneContact(out *MessageEnvelope, res *PhoneContact) {
	out.Constructor = C_PhoneContact
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAuthResendCode(out *MessageEnvelope, res *AuthResendCode) {
	out.Constructor = C_AuthResendCode
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateUserStatus(out *MessageEnvelope, res *UpdateUserStatus) {
	out.Constructor = C_UpdateUserStatus
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultSystemPublicKeys(out *MessageEnvelope, res *SystemPublicKeys) {
	out.Constructor = C_SystemPublicKeys
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultDHGroup(out *MessageEnvelope, res *DHGroup) {
	out.Constructor = C_DHGroup
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInitDB(out *MessageEnvelope, res *InitDB) {
	out.Constructor = C_InitDB
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultEchoWithDelay(out *MessageEnvelope, res *EchoWithDelay) {
	out.Constructor = C_EchoWithDelay
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultGroup(out *MessageEnvelope, res *Group) {
	out.Constructor = C_Group
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultSystemDHGroups(out *MessageEnvelope, res *SystemDHGroups) {
	out.Constructor = C_SystemDHGroups
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessagesSent(out *MessageEnvelope, res *MessagesSent) {
	out.Constructor = C_MessagesSent
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultGroupsGetFull(out *MessageEnvelope, res *GroupsGetFull) {
	out.Constructor = C_GroupsGetFull
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessagesSend(out *MessageEnvelope, res *MessagesSend) {
	out.Constructor = C_MessagesSend
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultClientUpdateMessagesDeleted(out *MessageEnvelope, res *ClientUpdateMessagesDeleted) {
	out.Constructor = C_ClientUpdateMessagesDeleted
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultGroupsDeleteUser(out *MessageEnvelope, res *GroupsDeleteUser) {
	out.Constructor = C_GroupsDeleteUser
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateNotifySettings(out *MessageEnvelope, res *UpdateNotifySettings) {
	out.Constructor = C_UpdateNotifySettings
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAuthRecalled(out *MessageEnvelope, res *AuthRecalled) {
	out.Constructor = C_AuthRecalled
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessagesDialogs(out *MessageEnvelope, res *MessagesDialogs) {
	out.Constructor = C_MessagesDialogs
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInputPeer(out *MessageEnvelope, res *InputPeer) {
	out.Constructor = C_InputPeer
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessagesGetHistory(out *MessageEnvelope, res *MessagesGetHistory) {
	out.Constructor = C_MessagesGetHistory
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateNewMessage(out *MessageEnvelope, res *UpdateNewMessage) {
	out.Constructor = C_UpdateNewMessage
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultContactsImport(out *MessageEnvelope, res *ContactsImport) {
	out.Constructor = C_ContactsImport
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultPeerNotifySettings(out *MessageEnvelope, res *PeerNotifySettings) {
	out.Constructor = C_PeerNotifySettings
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessageEntity(out *MessageEnvelope, res *MessageEntity) {
	out.Constructor = C_MessageEntity
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultMessagesDelete(out *MessageEnvelope, res *MessagesDelete) {
	out.Constructor = C_MessagesDelete
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAuthDestroyKey(out *MessageEnvelope, res *AuthDestroyKey) {
	out.Constructor = C_AuthDestroyKey
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAccountUpdateProfile(out *MessageEnvelope, res *AccountUpdateProfile) {
	out.Constructor = C_AccountUpdateProfile
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultFileSavePart(out *MessageEnvelope, res *FileSavePart) {
	out.Constructor = C_FileSavePart
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAccountPrivacyRules(out *MessageEnvelope, res *AccountPrivacyRules) {
	out.Constructor = C_AccountPrivacyRules
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultClientUpdatePendingMessageDelivery(out *MessageEnvelope, res *ClientUpdatePendingMessageDelivery) {
	out.Constructor = C_ClientUpdatePendingMessageDelivery
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInputUser(out *MessageEnvelope, res *InputUser) {
	out.Constructor = C_InputUser
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInputFile(out *MessageEnvelope, res *InputFile) {
	out.Constructor = C_InputFile
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultContactsMany(out *MessageEnvelope, res *ContactsMany) {
	out.Constructor = C_ContactsMany
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAccountUnregisterDevice(out *MessageEnvelope, res *AccountUnregisterDevice) {
	out.Constructor = C_AccountUnregisterDevice
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAuthSendCode(out *MessageEnvelope, res *AuthSendCode) {
	out.Constructor = C_AuthSendCode
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultGroupParticipant(out *MessageEnvelope, res *GroupParticipant) {
	out.Constructor = C_GroupParticipant
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInputDocument(out *MessageEnvelope, res *InputDocument) {
	out.Constructor = C_InputDocument
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultBool(out *MessageEnvelope, res *Bool) {
	out.Constructor = C_Bool
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInitResponse(out *MessageEnvelope, res *InitResponse) {
	out.Constructor = C_InitResponse
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAuthCheckPhone(out *MessageEnvelope, res *AuthCheckPhone) {
	out.Constructor = C_AuthCheckPhone
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultInitConnect(out *MessageEnvelope, res *InitConnect) {
	out.Constructor = C_InitConnect
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultFileGet(out *MessageEnvelope, res *FileGet) {
	out.Constructor = C_FileGet
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultAccountChangePhone(out *MessageEnvelope, res *AccountChangePhone) {
	out.Constructor = C_AccountChangePhone
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
func ResultUpdateUsername(out *MessageEnvelope, res *UpdateUsername) {
	out.Constructor = C_UpdateUsername
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}
