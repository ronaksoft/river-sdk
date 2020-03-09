// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: chat.api.updates.proto

package msg

import (
	fmt "fmt"
	pbytes "github.com/gobwas/pool/pbytes"
	proto "github.com/gogo/protobuf/proto"
	math "math"
	sync "sync"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

const C_UpdateGetState int64 = 1437250230

type poolUpdateGetState struct {
	pool sync.Pool
}

func (p *poolUpdateGetState) Get() *UpdateGetState {
	x, ok := p.pool.Get().(*UpdateGetState)
	if !ok {
		return &UpdateGetState{}
	}
	return x
}

func (p *poolUpdateGetState) Put(x *UpdateGetState) {
	p.pool.Put(x)
}

var PoolUpdateGetState = poolUpdateGetState{}

func ResultUpdateGetState(out *MessageEnvelope, res *UpdateGetState) {
	out.Constructor = C_UpdateGetState
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateGetDifference int64 = 556775761

type poolUpdateGetDifference struct {
	pool sync.Pool
}

func (p *poolUpdateGetDifference) Get() *UpdateGetDifference {
	x, ok := p.pool.Get().(*UpdateGetDifference)
	if !ok {
		return &UpdateGetDifference{}
	}
	return x
}

func (p *poolUpdateGetDifference) Put(x *UpdateGetDifference) {
	p.pool.Put(x)
}

var PoolUpdateGetDifference = poolUpdateGetDifference{}

func ResultUpdateGetDifference(out *MessageEnvelope, res *UpdateGetDifference) {
	out.Constructor = C_UpdateGetDifference
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateDifference int64 = 1742546619

type poolUpdateDifference struct {
	pool sync.Pool
}

func (p *poolUpdateDifference) Get() *UpdateDifference {
	x, ok := p.pool.Get().(*UpdateDifference)
	if !ok {
		return &UpdateDifference{}
	}
	x.Updates = x.Updates[:0]
	x.Users = x.Users[:0]
	x.Groups = x.Groups[:0]
	x.CurrentUpdateID = 0
	return x
}

func (p *poolUpdateDifference) Put(x *UpdateDifference) {
	p.pool.Put(x)
}

var PoolUpdateDifference = poolUpdateDifference{}

func ResultUpdateDifference(out *MessageEnvelope, res *UpdateDifference) {
	out.Constructor = C_UpdateDifference
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateTooLong int64 = 1531755547

type poolUpdateTooLong struct {
	pool sync.Pool
}

func (p *poolUpdateTooLong) Get() *UpdateTooLong {
	x, ok := p.pool.Get().(*UpdateTooLong)
	if !ok {
		return &UpdateTooLong{}
	}
	return x
}

func (p *poolUpdateTooLong) Put(x *UpdateTooLong) {
	p.pool.Put(x)
}

var PoolUpdateTooLong = poolUpdateTooLong{}

func ResultUpdateTooLong(out *MessageEnvelope, res *UpdateTooLong) {
	out.Constructor = C_UpdateTooLong
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateState int64 = 1837585836

type poolUpdateState struct {
	pool sync.Pool
}

func (p *poolUpdateState) Get() *UpdateState {
	x, ok := p.pool.Get().(*UpdateState)
	if !ok {
		return &UpdateState{}
	}
	return x
}

func (p *poolUpdateState) Put(x *UpdateState) {
	p.pool.Put(x)
}

var PoolUpdateState = poolUpdateState{}

func ResultUpdateState(out *MessageEnvelope, res *UpdateState) {
	out.Constructor = C_UpdateState
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateMessageID int64 = 2139063022

type poolUpdateMessageID struct {
	pool sync.Pool
}

func (p *poolUpdateMessageID) Get() *UpdateMessageID {
	x, ok := p.pool.Get().(*UpdateMessageID)
	if !ok {
		return &UpdateMessageID{}
	}
	return x
}

func (p *poolUpdateMessageID) Put(x *UpdateMessageID) {
	p.pool.Put(x)
}

var PoolUpdateMessageID = poolUpdateMessageID{}

func ResultUpdateMessageID(out *MessageEnvelope, res *UpdateMessageID) {
	out.Constructor = C_UpdateMessageID
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateNewMessage int64 = 3426925183

type poolUpdateNewMessage struct {
	pool sync.Pool
}

func (p *poolUpdateNewMessage) Get() *UpdateNewMessage {
	x, ok := p.pool.Get().(*UpdateNewMessage)
	if !ok {
		return &UpdateNewMessage{}
	}
	x.AccessHash = 0
	return x
}

func (p *poolUpdateNewMessage) Put(x *UpdateNewMessage) {
	p.pool.Put(x)
}

var PoolUpdateNewMessage = poolUpdateNewMessage{}

func ResultUpdateNewMessage(out *MessageEnvelope, res *UpdateNewMessage) {
	out.Constructor = C_UpdateNewMessage
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateMessageEdited int64 = 1825079988

type poolUpdateMessageEdited struct {
	pool sync.Pool
}

func (p *poolUpdateMessageEdited) Get() *UpdateMessageEdited {
	x, ok := p.pool.Get().(*UpdateMessageEdited)
	if !ok {
		return &UpdateMessageEdited{}
	}
	return x
}

func (p *poolUpdateMessageEdited) Put(x *UpdateMessageEdited) {
	p.pool.Put(x)
}

var PoolUpdateMessageEdited = poolUpdateMessageEdited{}

func ResultUpdateMessageEdited(out *MessageEnvelope, res *UpdateMessageEdited) {
	out.Constructor = C_UpdateMessageEdited
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateMessagesDeleted int64 = 670568714

type poolUpdateMessagesDeleted struct {
	pool sync.Pool
}

func (p *poolUpdateMessagesDeleted) Get() *UpdateMessagesDeleted {
	x, ok := p.pool.Get().(*UpdateMessagesDeleted)
	if !ok {
		return &UpdateMessagesDeleted{}
	}
	x.MessageIDs = x.MessageIDs[:0]
	x.Peer = nil
	return x
}

func (p *poolUpdateMessagesDeleted) Put(x *UpdateMessagesDeleted) {
	p.pool.Put(x)
}

var PoolUpdateMessagesDeleted = poolUpdateMessagesDeleted{}

func ResultUpdateMessagesDeleted(out *MessageEnvelope, res *UpdateMessagesDeleted) {
	out.Constructor = C_UpdateMessagesDeleted
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateReadHistoryInbox int64 = 1529128378

type poolUpdateReadHistoryInbox struct {
	pool sync.Pool
}

func (p *poolUpdateReadHistoryInbox) Get() *UpdateReadHistoryInbox {
	x, ok := p.pool.Get().(*UpdateReadHistoryInbox)
	if !ok {
		return &UpdateReadHistoryInbox{}
	}
	return x
}

func (p *poolUpdateReadHistoryInbox) Put(x *UpdateReadHistoryInbox) {
	p.pool.Put(x)
}

var PoolUpdateReadHistoryInbox = poolUpdateReadHistoryInbox{}

func ResultUpdateReadHistoryInbox(out *MessageEnvelope, res *UpdateReadHistoryInbox) {
	out.Constructor = C_UpdateReadHistoryInbox
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateReadHistoryOutbox int64 = 510866108

type poolUpdateReadHistoryOutbox struct {
	pool sync.Pool
}

func (p *poolUpdateReadHistoryOutbox) Get() *UpdateReadHistoryOutbox {
	x, ok := p.pool.Get().(*UpdateReadHistoryOutbox)
	if !ok {
		return &UpdateReadHistoryOutbox{}
	}
	return x
}

func (p *poolUpdateReadHistoryOutbox) Put(x *UpdateReadHistoryOutbox) {
	p.pool.Put(x)
}

var PoolUpdateReadHistoryOutbox = poolUpdateReadHistoryOutbox{}

func ResultUpdateReadHistoryOutbox(out *MessageEnvelope, res *UpdateReadHistoryOutbox) {
	out.Constructor = C_UpdateReadHistoryOutbox
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateUserTyping int64 = 178254060

type poolUpdateUserTyping struct {
	pool sync.Pool
}

func (p *poolUpdateUserTyping) Get() *UpdateUserTyping {
	x, ok := p.pool.Get().(*UpdateUserTyping)
	if !ok {
		return &UpdateUserTyping{}
	}
	return x
}

func (p *poolUpdateUserTyping) Put(x *UpdateUserTyping) {
	p.pool.Put(x)
}

var PoolUpdateUserTyping = poolUpdateUserTyping{}

func ResultUpdateUserTyping(out *MessageEnvelope, res *UpdateUserTyping) {
	out.Constructor = C_UpdateUserTyping
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateUserStatus int64 = 2696747995

type poolUpdateUserStatus struct {
	pool sync.Pool
}

func (p *poolUpdateUserStatus) Get() *UpdateUserStatus {
	x, ok := p.pool.Get().(*UpdateUserStatus)
	if !ok {
		return &UpdateUserStatus{}
	}
	return x
}

func (p *poolUpdateUserStatus) Put(x *UpdateUserStatus) {
	p.pool.Put(x)
}

var PoolUpdateUserStatus = poolUpdateUserStatus{}

func ResultUpdateUserStatus(out *MessageEnvelope, res *UpdateUserStatus) {
	out.Constructor = C_UpdateUserStatus
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateUsername int64 = 4290110589

type poolUpdateUsername struct {
	pool sync.Pool
}

func (p *poolUpdateUsername) Get() *UpdateUsername {
	x, ok := p.pool.Get().(*UpdateUsername)
	if !ok {
		return &UpdateUsername{}
	}
	x.Phone = ""
	return x
}

func (p *poolUpdateUsername) Put(x *UpdateUsername) {
	p.pool.Put(x)
}

var PoolUpdateUsername = poolUpdateUsername{}

func ResultUpdateUsername(out *MessageEnvelope, res *UpdateUsername) {
	out.Constructor = C_UpdateUsername
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateUserPhoto int64 = 302028082

type poolUpdateUserPhoto struct {
	pool sync.Pool
}

func (p *poolUpdateUserPhoto) Get() *UpdateUserPhoto {
	x, ok := p.pool.Get().(*UpdateUserPhoto)
	if !ok {
		return &UpdateUserPhoto{}
	}
	x.Photo = nil
	x.PhotoID = 0
	x.DeletedPhotoIDs = x.DeletedPhotoIDs[:0]
	return x
}

func (p *poolUpdateUserPhoto) Put(x *UpdateUserPhoto) {
	p.pool.Put(x)
}

var PoolUpdateUserPhoto = poolUpdateUserPhoto{}

func ResultUpdateUserPhoto(out *MessageEnvelope, res *UpdateUserPhoto) {
	out.Constructor = C_UpdateUserPhoto
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateNotifySettings int64 = 3187524885

type poolUpdateNotifySettings struct {
	pool sync.Pool
}

func (p *poolUpdateNotifySettings) Get() *UpdateNotifySettings {
	x, ok := p.pool.Get().(*UpdateNotifySettings)
	if !ok {
		return &UpdateNotifySettings{}
	}
	return x
}

func (p *poolUpdateNotifySettings) Put(x *UpdateNotifySettings) {
	p.pool.Put(x)
}

var PoolUpdateNotifySettings = poolUpdateNotifySettings{}

func ResultUpdateNotifySettings(out *MessageEnvelope, res *UpdateNotifySettings) {
	out.Constructor = C_UpdateNotifySettings
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateGroupParticipantAdd int64 = 1623827837

type poolUpdateGroupParticipantAdd struct {
	pool sync.Pool
}

func (p *poolUpdateGroupParticipantAdd) Get() *UpdateGroupParticipantAdd {
	x, ok := p.pool.Get().(*UpdateGroupParticipantAdd)
	if !ok {
		return &UpdateGroupParticipantAdd{}
	}
	return x
}

func (p *poolUpdateGroupParticipantAdd) Put(x *UpdateGroupParticipantAdd) {
	p.pool.Put(x)
}

var PoolUpdateGroupParticipantAdd = poolUpdateGroupParticipantAdd{}

func ResultUpdateGroupParticipantAdd(out *MessageEnvelope, res *UpdateGroupParticipantAdd) {
	out.Constructor = C_UpdateGroupParticipantAdd
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateGroupParticipantDeleted int64 = 2489941844

type poolUpdateGroupParticipantDeleted struct {
	pool sync.Pool
}

func (p *poolUpdateGroupParticipantDeleted) Get() *UpdateGroupParticipantDeleted {
	x, ok := p.pool.Get().(*UpdateGroupParticipantDeleted)
	if !ok {
		return &UpdateGroupParticipantDeleted{}
	}
	return x
}

func (p *poolUpdateGroupParticipantDeleted) Put(x *UpdateGroupParticipantDeleted) {
	p.pool.Put(x)
}

var PoolUpdateGroupParticipantDeleted = poolUpdateGroupParticipantDeleted{}

func ResultUpdateGroupParticipantDeleted(out *MessageEnvelope, res *UpdateGroupParticipantDeleted) {
	out.Constructor = C_UpdateGroupParticipantDeleted
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateGroupParticipantAdmin int64 = 1813022164

type poolUpdateGroupParticipantAdmin struct {
	pool sync.Pool
}

func (p *poolUpdateGroupParticipantAdmin) Get() *UpdateGroupParticipantAdmin {
	x, ok := p.pool.Get().(*UpdateGroupParticipantAdmin)
	if !ok {
		return &UpdateGroupParticipantAdmin{}
	}
	return x
}

func (p *poolUpdateGroupParticipantAdmin) Put(x *UpdateGroupParticipantAdmin) {
	p.pool.Put(x)
}

var PoolUpdateGroupParticipantAdmin = poolUpdateGroupParticipantAdmin{}

func ResultUpdateGroupParticipantAdmin(out *MessageEnvelope, res *UpdateGroupParticipantAdmin) {
	out.Constructor = C_UpdateGroupParticipantAdmin
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateGroupAdmins int64 = 694155405

type poolUpdateGroupAdmins struct {
	pool sync.Pool
}

func (p *poolUpdateGroupAdmins) Get() *UpdateGroupAdmins {
	x, ok := p.pool.Get().(*UpdateGroupAdmins)
	if !ok {
		return &UpdateGroupAdmins{}
	}
	return x
}

func (p *poolUpdateGroupAdmins) Put(x *UpdateGroupAdmins) {
	p.pool.Put(x)
}

var PoolUpdateGroupAdmins = poolUpdateGroupAdmins{}

func ResultUpdateGroupAdmins(out *MessageEnvelope, res *UpdateGroupAdmins) {
	out.Constructor = C_UpdateGroupAdmins
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateGroupPhoto int64 = 367193154

type poolUpdateGroupPhoto struct {
	pool sync.Pool
}

func (p *poolUpdateGroupPhoto) Get() *UpdateGroupPhoto {
	x, ok := p.pool.Get().(*UpdateGroupPhoto)
	if !ok {
		return &UpdateGroupPhoto{}
	}
	x.Photo = nil
	x.PhotoID = 0
	return x
}

func (p *poolUpdateGroupPhoto) Put(x *UpdateGroupPhoto) {
	p.pool.Put(x)
}

var PoolUpdateGroupPhoto = poolUpdateGroupPhoto{}

func ResultUpdateGroupPhoto(out *MessageEnvelope, res *UpdateGroupPhoto) {
	out.Constructor = C_UpdateGroupPhoto
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateReadMessagesContents int64 = 2991403048

type poolUpdateReadMessagesContents struct {
	pool sync.Pool
}

func (p *poolUpdateReadMessagesContents) Get() *UpdateReadMessagesContents {
	x, ok := p.pool.Get().(*UpdateReadMessagesContents)
	if !ok {
		return &UpdateReadMessagesContents{}
	}
	x.MessageIDs = x.MessageIDs[:0]
	return x
}

func (p *poolUpdateReadMessagesContents) Put(x *UpdateReadMessagesContents) {
	p.pool.Put(x)
}

var PoolUpdateReadMessagesContents = poolUpdateReadMessagesContents{}

func ResultUpdateReadMessagesContents(out *MessageEnvelope, res *UpdateReadMessagesContents) {
	out.Constructor = C_UpdateReadMessagesContents
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateAuthorizationReset int64 = 2359297647

type poolUpdateAuthorizationReset struct {
	pool sync.Pool
}

func (p *poolUpdateAuthorizationReset) Get() *UpdateAuthorizationReset {
	x, ok := p.pool.Get().(*UpdateAuthorizationReset)
	if !ok {
		return &UpdateAuthorizationReset{}
	}
	return x
}

func (p *poolUpdateAuthorizationReset) Put(x *UpdateAuthorizationReset) {
	p.pool.Put(x)
}

var PoolUpdateAuthorizationReset = poolUpdateAuthorizationReset{}

func ResultUpdateAuthorizationReset(out *MessageEnvelope, res *UpdateAuthorizationReset) {
	out.Constructor = C_UpdateAuthorizationReset
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateDraftMessage int64 = 3453026195

type poolUpdateDraftMessage struct {
	pool sync.Pool
}

func (p *poolUpdateDraftMessage) Get() *UpdateDraftMessage {
	x, ok := p.pool.Get().(*UpdateDraftMessage)
	if !ok {
		return &UpdateDraftMessage{}
	}
	x.UpdateID = 0
	return x
}

func (p *poolUpdateDraftMessage) Put(x *UpdateDraftMessage) {
	p.pool.Put(x)
}

var PoolUpdateDraftMessage = poolUpdateDraftMessage{}

func ResultUpdateDraftMessage(out *MessageEnvelope, res *UpdateDraftMessage) {
	out.Constructor = C_UpdateDraftMessage
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateDraftMessageCleared int64 = 2011635602

type poolUpdateDraftMessageCleared struct {
	pool sync.Pool
}

func (p *poolUpdateDraftMessageCleared) Get() *UpdateDraftMessageCleared {
	x, ok := p.pool.Get().(*UpdateDraftMessageCleared)
	if !ok {
		return &UpdateDraftMessageCleared{}
	}
	x.UpdateID = 0
	return x
}

func (p *poolUpdateDraftMessageCleared) Put(x *UpdateDraftMessageCleared) {
	p.pool.Put(x)
}

var PoolUpdateDraftMessageCleared = poolUpdateDraftMessageCleared{}

func ResultUpdateDraftMessageCleared(out *MessageEnvelope, res *UpdateDraftMessageCleared) {
	out.Constructor = C_UpdateDraftMessageCleared
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateDialogPinned int64 = 231538299

type poolUpdateDialogPinned struct {
	pool sync.Pool
}

func (p *poolUpdateDialogPinned) Get() *UpdateDialogPinned {
	x, ok := p.pool.Get().(*UpdateDialogPinned)
	if !ok {
		return &UpdateDialogPinned{}
	}
	return x
}

func (p *poolUpdateDialogPinned) Put(x *UpdateDialogPinned) {
	p.pool.Put(x)
}

var PoolUpdateDialogPinned = poolUpdateDialogPinned{}

func ResultUpdateDialogPinned(out *MessageEnvelope, res *UpdateDialogPinned) {
	out.Constructor = C_UpdateDialogPinned
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateDialogPinnedReorder int64 = 1567423539

type poolUpdateDialogPinnedReorder struct {
	pool sync.Pool
}

func (p *poolUpdateDialogPinnedReorder) Get() *UpdateDialogPinnedReorder {
	x, ok := p.pool.Get().(*UpdateDialogPinnedReorder)
	if !ok {
		return &UpdateDialogPinnedReorder{}
	}
	x.Peer = x.Peer[:0]
	return x
}

func (p *poolUpdateDialogPinnedReorder) Put(x *UpdateDialogPinnedReorder) {
	p.pool.Put(x)
}

var PoolUpdateDialogPinnedReorder = poolUpdateDialogPinnedReorder{}

func ResultUpdateDialogPinnedReorder(out *MessageEnvelope, res *UpdateDialogPinnedReorder) {
	out.Constructor = C_UpdateDialogPinnedReorder
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateAccountPrivacy int64 = 629173761

type poolUpdateAccountPrivacy struct {
	pool sync.Pool
}

func (p *poolUpdateAccountPrivacy) Get() *UpdateAccountPrivacy {
	x, ok := p.pool.Get().(*UpdateAccountPrivacy)
	if !ok {
		return &UpdateAccountPrivacy{}
	}
	x.ChatInvite = x.ChatInvite[:0]
	x.LastSeen = x.LastSeen[:0]
	x.PhoneNumber = x.PhoneNumber[:0]
	x.ProfilePhoto = x.ProfilePhoto[:0]
	x.ForwardedMessage = x.ForwardedMessage[:0]
	x.Call = x.Call[:0]
	return x
}

func (p *poolUpdateAccountPrivacy) Put(x *UpdateAccountPrivacy) {
	p.pool.Put(x)
}

var PoolUpdateAccountPrivacy = poolUpdateAccountPrivacy{}

func ResultUpdateAccountPrivacy(out *MessageEnvelope, res *UpdateAccountPrivacy) {
	out.Constructor = C_UpdateAccountPrivacy
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateLabelItemsAdded int64 = 2216022057

type poolUpdateLabelItemsAdded struct {
	pool sync.Pool
}

func (p *poolUpdateLabelItemsAdded) Get() *UpdateLabelItemsAdded {
	x, ok := p.pool.Get().(*UpdateLabelItemsAdded)
	if !ok {
		return &UpdateLabelItemsAdded{}
	}
	x.MessageIDs = x.MessageIDs[:0]
	x.LabelIDs = x.LabelIDs[:0]
	x.Labels = x.Labels[:0]
	return x
}

func (p *poolUpdateLabelItemsAdded) Put(x *UpdateLabelItemsAdded) {
	p.pool.Put(x)
}

var PoolUpdateLabelItemsAdded = poolUpdateLabelItemsAdded{}

func ResultUpdateLabelItemsAdded(out *MessageEnvelope, res *UpdateLabelItemsAdded) {
	out.Constructor = C_UpdateLabelItemsAdded
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateLabelItemsRemoved int64 = 830226827

type poolUpdateLabelItemsRemoved struct {
	pool sync.Pool
}

func (p *poolUpdateLabelItemsRemoved) Get() *UpdateLabelItemsRemoved {
	x, ok := p.pool.Get().(*UpdateLabelItemsRemoved)
	if !ok {
		return &UpdateLabelItemsRemoved{}
	}
	x.MessageIDs = x.MessageIDs[:0]
	x.LabelIDs = x.LabelIDs[:0]
	x.Labels = x.Labels[:0]
	return x
}

func (p *poolUpdateLabelItemsRemoved) Put(x *UpdateLabelItemsRemoved) {
	p.pool.Put(x)
}

var PoolUpdateLabelItemsRemoved = poolUpdateLabelItemsRemoved{}

func ResultUpdateLabelItemsRemoved(out *MessageEnvelope, res *UpdateLabelItemsRemoved) {
	out.Constructor = C_UpdateLabelItemsRemoved
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateLabelSet int64 = 2353687359

type poolUpdateLabelSet struct {
	pool sync.Pool
}

func (p *poolUpdateLabelSet) Get() *UpdateLabelSet {
	x, ok := p.pool.Get().(*UpdateLabelSet)
	if !ok {
		return &UpdateLabelSet{}
	}
	x.Labels = x.Labels[:0]
	return x
}

func (p *poolUpdateLabelSet) Put(x *UpdateLabelSet) {
	p.pool.Put(x)
}

var PoolUpdateLabelSet = poolUpdateLabelSet{}

func ResultUpdateLabelSet(out *MessageEnvelope, res *UpdateLabelSet) {
	out.Constructor = C_UpdateLabelSet
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateLabelDeleted int64 = 3702192307

type poolUpdateLabelDeleted struct {
	pool sync.Pool
}

func (p *poolUpdateLabelDeleted) Get() *UpdateLabelDeleted {
	x, ok := p.pool.Get().(*UpdateLabelDeleted)
	if !ok {
		return &UpdateLabelDeleted{}
	}
	x.LabelIDs = x.LabelIDs[:0]
	return x
}

func (p *poolUpdateLabelDeleted) Put(x *UpdateLabelDeleted) {
	p.pool.Put(x)
}

var PoolUpdateLabelDeleted = poolUpdateLabelDeleted{}

func ResultUpdateLabelDeleted(out *MessageEnvelope, res *UpdateLabelDeleted) {
	out.Constructor = C_UpdateLabelDeleted
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateUserBlocked int64 = 3750625773

type poolUpdateUserBlocked struct {
	pool sync.Pool
}

func (p *poolUpdateUserBlocked) Get() *UpdateUserBlocked {
	x, ok := p.pool.Get().(*UpdateUserBlocked)
	if !ok {
		return &UpdateUserBlocked{}
	}
	return x
}

func (p *poolUpdateUserBlocked) Put(x *UpdateUserBlocked) {
	p.pool.Put(x)
}

var PoolUpdateUserBlocked = poolUpdateUserBlocked{}

func ResultUpdateUserBlocked(out *MessageEnvelope, res *UpdateUserBlocked) {
	out.Constructor = C_UpdateUserBlocked
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateMessagePoll int64 = 383248674

type poolUpdateMessagePoll struct {
	pool sync.Pool
}

func (p *poolUpdateMessagePoll) Get() *UpdateMessagePoll {
	x, ok := p.pool.Get().(*UpdateMessagePoll)
	if !ok {
		return &UpdateMessagePoll{}
	}
	x.Poll = nil
	return x
}

func (p *poolUpdateMessagePoll) Put(x *UpdateMessagePoll) {
	p.pool.Put(x)
}

var PoolUpdateMessagePoll = poolUpdateMessagePoll{}

func ResultUpdateMessagePoll(out *MessageEnvelope, res *UpdateMessagePoll) {
	out.Constructor = C_UpdateMessagePoll
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateBotCallbackQuery int64 = 3408999713

type poolUpdateBotCallbackQuery struct {
	pool sync.Pool
}

func (p *poolUpdateBotCallbackQuery) Get() *UpdateBotCallbackQuery {
	x, ok := p.pool.Get().(*UpdateBotCallbackQuery)
	if !ok {
		return &UpdateBotCallbackQuery{}
	}
	x.MessageID = 0
	x.Data = nil
	return x
}

func (p *poolUpdateBotCallbackQuery) Put(x *UpdateBotCallbackQuery) {
	p.pool.Put(x)
}

var PoolUpdateBotCallbackQuery = poolUpdateBotCallbackQuery{}

func ResultUpdateBotCallbackQuery(out *MessageEnvelope, res *UpdateBotCallbackQuery) {
	out.Constructor = C_UpdateBotCallbackQuery
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UpdateBotInlineCallbackQuery int64 = 426698513

type poolUpdateBotInlineCallbackQuery struct {
	pool sync.Pool
}

func (p *poolUpdateBotInlineCallbackQuery) Get() *UpdateBotInlineCallbackQuery {
	x, ok := p.pool.Get().(*UpdateBotInlineCallbackQuery)
	if !ok {
		return &UpdateBotInlineCallbackQuery{}
	}
	x.MessageID = 0
	x.Data = nil
	return x
}

func (p *poolUpdateBotInlineCallbackQuery) Put(x *UpdateBotInlineCallbackQuery) {
	p.pool.Put(x)
}

var PoolUpdateBotInlineCallbackQuery = poolUpdateBotInlineCallbackQuery{}

func ResultUpdateBotInlineCallbackQuery(out *MessageEnvelope, res *UpdateBotInlineCallbackQuery) {
	out.Constructor = C_UpdateBotInlineCallbackQuery
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

func init() {
	ConstructorNames[1437250230] = "UpdateGetState"
	ConstructorNames[556775761] = "UpdateGetDifference"
	ConstructorNames[1742546619] = "UpdateDifference"
	ConstructorNames[1531755547] = "UpdateTooLong"
	ConstructorNames[1837585836] = "UpdateState"
	ConstructorNames[2139063022] = "UpdateMessageID"
	ConstructorNames[3426925183] = "UpdateNewMessage"
	ConstructorNames[1825079988] = "UpdateMessageEdited"
	ConstructorNames[670568714] = "UpdateMessagesDeleted"
	ConstructorNames[1529128378] = "UpdateReadHistoryInbox"
	ConstructorNames[510866108] = "UpdateReadHistoryOutbox"
	ConstructorNames[178254060] = "UpdateUserTyping"
	ConstructorNames[2696747995] = "UpdateUserStatus"
	ConstructorNames[4290110589] = "UpdateUsername"
	ConstructorNames[302028082] = "UpdateUserPhoto"
	ConstructorNames[3187524885] = "UpdateNotifySettings"
	ConstructorNames[1623827837] = "UpdateGroupParticipantAdd"
	ConstructorNames[2489941844] = "UpdateGroupParticipantDeleted"
	ConstructorNames[1813022164] = "UpdateGroupParticipantAdmin"
	ConstructorNames[694155405] = "UpdateGroupAdmins"
	ConstructorNames[367193154] = "UpdateGroupPhoto"
	ConstructorNames[2991403048] = "UpdateReadMessagesContents"
	ConstructorNames[2359297647] = "UpdateAuthorizationReset"
	ConstructorNames[3453026195] = "UpdateDraftMessage"
	ConstructorNames[2011635602] = "UpdateDraftMessageCleared"
	ConstructorNames[231538299] = "UpdateDialogPinned"
	ConstructorNames[1567423539] = "UpdateDialogPinnedReorder"
	ConstructorNames[629173761] = "UpdateAccountPrivacy"
	ConstructorNames[2216022057] = "UpdateLabelItemsAdded"
	ConstructorNames[830226827] = "UpdateLabelItemsRemoved"
	ConstructorNames[2353687359] = "UpdateLabelSet"
	ConstructorNames[3702192307] = "UpdateLabelDeleted"
	ConstructorNames[3750625773] = "UpdateUserBlocked"
	ConstructorNames[383248674] = "UpdateMessagePoll"
	ConstructorNames[3408999713] = "UpdateBotCallbackQuery"
	ConstructorNames[426698513] = "UpdateBotInlineCallbackQuery"
}
