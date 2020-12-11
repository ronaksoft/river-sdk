package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_UpdateGetState int64 = 1632290153

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

const C_UpdateGetDifference int64 = 3467307691

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
	x.From = 0
	x.Limit = 0
	p.pool.Put(x)
}

var PoolUpdateGetDifference = poolUpdateGetDifference{}

const C_UpdateDifference int64 = 2937859860

type poolUpdateDifference struct {
	pool sync.Pool
}

func (p *poolUpdateDifference) Get() *UpdateDifference {
	x, ok := p.pool.Get().(*UpdateDifference)
	if !ok {
		return &UpdateDifference{}
	}
	return x
}

func (p *poolUpdateDifference) Put(x *UpdateDifference) {
	x.More = false
	x.MaxUpdateID = 0
	x.MinUpdateID = 0
	x.Updates = x.Updates[:0]
	x.Users = x.Users[:0]
	x.Groups = x.Groups[:0]
	x.CurrentUpdateID = 0
	p.pool.Put(x)
}

var PoolUpdateDifference = poolUpdateDifference{}

const C_UpdateTooLong int64 = 3588398962

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

const C_UpdateState int64 = 2070778422

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
	x.UpdateID = 0
	p.pool.Put(x)
}

var PoolUpdateState = poolUpdateState{}

const C_UpdateMessageID int64 = 1764208092

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
	x.UCount = 0
	x.MessageID = 0
	x.RandomID = 0
	p.pool.Put(x)
}

var PoolUpdateMessageID = poolUpdateMessageID{}

const C_UpdateNewMessage int64 = 75740112

type poolUpdateNewMessage struct {
	pool sync.Pool
}

func (p *poolUpdateNewMessage) Get() *UpdateNewMessage {
	x, ok := p.pool.Get().(*UpdateNewMessage)
	if !ok {
		return &UpdateNewMessage{}
	}
	return x
}

func (p *poolUpdateNewMessage) Put(x *UpdateNewMessage) {
	x.UCount = 0
	x.UpdateID = 0
	if x.Message != nil {
		PoolUserMessage.Put(x.Message)
		x.Message = nil
	}
	if x.Sender != nil {
		PoolUser.Put(x.Sender)
		x.Sender = nil
	}
	x.AccessHash = 0
	x.SenderRefID = 0
	p.pool.Put(x)
}

var PoolUpdateNewMessage = poolUpdateNewMessage{}

const C_UpdateMessageEdited int64 = 2202915150

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
	x.UCount = 0
	x.UpdateID = 0
	if x.Message != nil {
		PoolUserMessage.Put(x.Message)
		x.Message = nil
	}
	p.pool.Put(x)
}

var PoolUpdateMessageEdited = poolUpdateMessageEdited{}

const C_UpdateMessagesDeleted int64 = 1003091958

type poolUpdateMessagesDeleted struct {
	pool sync.Pool
}

func (p *poolUpdateMessagesDeleted) Get() *UpdateMessagesDeleted {
	x, ok := p.pool.Get().(*UpdateMessagesDeleted)
	if !ok {
		return &UpdateMessagesDeleted{}
	}
	return x
}

func (p *poolUpdateMessagesDeleted) Put(x *UpdateMessagesDeleted) {
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	x.MessageIDs = x.MessageIDs[:0]
	if x.Peer != nil {
		PoolPeer.Put(x.Peer)
		x.Peer = nil
	}
	p.pool.Put(x)
}

var PoolUpdateMessagesDeleted = poolUpdateMessagesDeleted{}

const C_UpdateReadHistoryInbox int64 = 4013107819

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
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	if x.Peer != nil {
		PoolPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MaxID = 0
	p.pool.Put(x)
}

var PoolUpdateReadHistoryInbox = poolUpdateReadHistoryInbox{}

const C_UpdateReadHistoryOutbox int64 = 4011050865

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
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	if x.Peer != nil {
		PoolPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MaxID = 0
	x.UserID = 0
	p.pool.Put(x)
}

var PoolUpdateReadHistoryOutbox = poolUpdateReadHistoryOutbox{}

const C_UpdateMessagePinned int64 = 2761988296

type poolUpdateMessagePinned struct {
	pool sync.Pool
}

func (p *poolUpdateMessagePinned) Get() *UpdateMessagePinned {
	x, ok := p.pool.Get().(*UpdateMessagePinned)
	if !ok {
		return &UpdateMessagePinned{}
	}
	return x
}

func (p *poolUpdateMessagePinned) Put(x *UpdateMessagePinned) {
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	if x.Peer != nil {
		PoolPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.UserID = 0
	x.MsgID = 0
	x.Version = 0
	p.pool.Put(x)
}

var PoolUpdateMessagePinned = poolUpdateMessagePinned{}

const C_UpdateUserTyping int64 = 3261004099

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
	x.UCount = 0
	x.TeamID = 0
	x.UserID = 0
	x.Action = 0
	x.PeerID = 0
	x.PeerType = 0
	p.pool.Put(x)
}

var PoolUpdateUserTyping = poolUpdateUserTyping{}

const C_UpdateUserStatus int64 = 1752961652

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
	x.UCount = 0
	x.UserID = 0
	x.Status = 0
	p.pool.Put(x)
}

var PoolUpdateUserStatus = poolUpdateUserStatus{}

const C_UpdateUsername int64 = 3411383202

type poolUpdateUsername struct {
	pool sync.Pool
}

func (p *poolUpdateUsername) Get() *UpdateUsername {
	x, ok := p.pool.Get().(*UpdateUsername)
	if !ok {
		return &UpdateUsername{}
	}
	return x
}

func (p *poolUpdateUsername) Put(x *UpdateUsername) {
	x.UCount = 0
	x.UpdateID = 0
	x.UserID = 0
	x.Username = ""
	x.FirstName = ""
	x.LastName = ""
	x.Bio = ""
	x.Phone = ""
	p.pool.Put(x)
}

var PoolUpdateUsername = poolUpdateUsername{}

const C_UpdateUserPhoto int64 = 72923648

type poolUpdateUserPhoto struct {
	pool sync.Pool
}

func (p *poolUpdateUserPhoto) Get() *UpdateUserPhoto {
	x, ok := p.pool.Get().(*UpdateUserPhoto)
	if !ok {
		return &UpdateUserPhoto{}
	}
	return x
}

func (p *poolUpdateUserPhoto) Put(x *UpdateUserPhoto) {
	x.UCount = 0
	x.UpdateID = 0
	x.UserID = 0
	if x.Photo != nil {
		PoolUserPhoto.Put(x.Photo)
		x.Photo = nil
	}
	x.PhotoID = 0
	x.DeletedPhotoIDs = x.DeletedPhotoIDs[:0]
	p.pool.Put(x)
}

var PoolUpdateUserPhoto = poolUpdateUserPhoto{}

const C_UpdateNotifySettings int64 = 3766115140

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
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	x.UserID = 0
	if x.NotifyPeer != nil {
		PoolPeer.Put(x.NotifyPeer)
		x.NotifyPeer = nil
	}
	if x.Settings != nil {
		PoolPeerNotifySettings.Put(x.Settings)
		x.Settings = nil
	}
	p.pool.Put(x)
}

var PoolUpdateNotifySettings = poolUpdateNotifySettings{}

const C_UpdateGroupParticipantAdd int64 = 3544906637

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
	x.UCount = 0
	x.UpdateID = 0
	x.GroupID = 0
	x.UserID = 0
	x.InviterID = 0
	x.Date = 0
	p.pool.Put(x)
}

var PoolUpdateGroupParticipantAdd = poolUpdateGroupParticipantAdd{}

const C_UpdateGroupParticipantDeleted int64 = 921388023

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
	x.UCount = 0
	x.UpdateID = 0
	x.GroupID = 0
	x.UserID = 0
	p.pool.Put(x)
}

var PoolUpdateGroupParticipantDeleted = poolUpdateGroupParticipantDeleted{}

const C_UpdateGroupParticipantAdmin int64 = 4102007577

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
	x.UCount = 0
	x.UpdateID = 0
	x.GroupID = 0
	x.UserID = 0
	x.IsAdmin = false
	p.pool.Put(x)
}

var PoolUpdateGroupParticipantAdmin = poolUpdateGroupParticipantAdmin{}

const C_UpdateGroupAdmins int64 = 1878951933

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
	x.UCount = 0
	x.UpdateID = 0
	x.GroupID = 0
	x.AdminEnabled = false
	p.pool.Put(x)
}

var PoolUpdateGroupAdmins = poolUpdateGroupAdmins{}

const C_UpdateGroupPhoto int64 = 3710117357

type poolUpdateGroupPhoto struct {
	pool sync.Pool
}

func (p *poolUpdateGroupPhoto) Get() *UpdateGroupPhoto {
	x, ok := p.pool.Get().(*UpdateGroupPhoto)
	if !ok {
		return &UpdateGroupPhoto{}
	}
	return x
}

func (p *poolUpdateGroupPhoto) Put(x *UpdateGroupPhoto) {
	x.UCount = 0
	x.UpdateID = 0
	x.GroupID = 0
	if x.Photo != nil {
		PoolGroupPhoto.Put(x.Photo)
		x.Photo = nil
	}
	x.PhotoID = 0
	p.pool.Put(x)
}

var PoolUpdateGroupPhoto = poolUpdateGroupPhoto{}

const C_UpdateReadMessagesContents int64 = 256065898

type poolUpdateReadMessagesContents struct {
	pool sync.Pool
}

func (p *poolUpdateReadMessagesContents) Get() *UpdateReadMessagesContents {
	x, ok := p.pool.Get().(*UpdateReadMessagesContents)
	if !ok {
		return &UpdateReadMessagesContents{}
	}
	return x
}

func (p *poolUpdateReadMessagesContents) Put(x *UpdateReadMessagesContents) {
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	x.MessageIDs = x.MessageIDs[:0]
	if x.Peer != nil {
		PoolPeer.Put(x.Peer)
		x.Peer = nil
	}
	p.pool.Put(x)
}

var PoolUpdateReadMessagesContents = poolUpdateReadMessagesContents{}

const C_UpdateAuthorizationReset int64 = 1770313879

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
	x.UCount = 0
	x.UpdateID = 0
	p.pool.Put(x)
}

var PoolUpdateAuthorizationReset = poolUpdateAuthorizationReset{}

const C_UpdateDraftMessage int64 = 2643668944

type poolUpdateDraftMessage struct {
	pool sync.Pool
}

func (p *poolUpdateDraftMessage) Get() *UpdateDraftMessage {
	x, ok := p.pool.Get().(*UpdateDraftMessage)
	if !ok {
		return &UpdateDraftMessage{}
	}
	return x
}

func (p *poolUpdateDraftMessage) Put(x *UpdateDraftMessage) {
	x.UCount = 0
	x.UpdateID = 0
	if x.Message != nil {
		PoolDraftMessage.Put(x.Message)
		x.Message = nil
	}
	p.pool.Put(x)
}

var PoolUpdateDraftMessage = poolUpdateDraftMessage{}

const C_UpdateDraftMessageCleared int64 = 3294904674

type poolUpdateDraftMessageCleared struct {
	pool sync.Pool
}

func (p *poolUpdateDraftMessageCleared) Get() *UpdateDraftMessageCleared {
	x, ok := p.pool.Get().(*UpdateDraftMessageCleared)
	if !ok {
		return &UpdateDraftMessageCleared{}
	}
	return x
}

func (p *poolUpdateDraftMessageCleared) Put(x *UpdateDraftMessageCleared) {
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	if x.Peer != nil {
		PoolPeer.Put(x.Peer)
		x.Peer = nil
	}
	p.pool.Put(x)
}

var PoolUpdateDraftMessageCleared = poolUpdateDraftMessageCleared{}

const C_UpdateDialogPinned int64 = 1569664568

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
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	if x.Peer != nil {
		PoolPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.Pinned = false
	p.pool.Put(x)
}

var PoolUpdateDialogPinned = poolUpdateDialogPinned{}

const C_UpdateDialogPinnedReorder int64 = 4008682179

type poolUpdateDialogPinnedReorder struct {
	pool sync.Pool
}

func (p *poolUpdateDialogPinnedReorder) Get() *UpdateDialogPinnedReorder {
	x, ok := p.pool.Get().(*UpdateDialogPinnedReorder)
	if !ok {
		return &UpdateDialogPinnedReorder{}
	}
	return x
}

func (p *poolUpdateDialogPinnedReorder) Put(x *UpdateDialogPinnedReorder) {
	x.UCount = 0
	x.UpdateID = 0
	x.Peer = x.Peer[:0]
	p.pool.Put(x)
}

var PoolUpdateDialogPinnedReorder = poolUpdateDialogPinnedReorder{}

const C_UpdateAccountPrivacy int64 = 2013786192

type poolUpdateAccountPrivacy struct {
	pool sync.Pool
}

func (p *poolUpdateAccountPrivacy) Get() *UpdateAccountPrivacy {
	x, ok := p.pool.Get().(*UpdateAccountPrivacy)
	if !ok {
		return &UpdateAccountPrivacy{}
	}
	return x
}

func (p *poolUpdateAccountPrivacy) Put(x *UpdateAccountPrivacy) {
	x.UCount = 0
	x.UpdateID = 0
	x.ChatInvite = x.ChatInvite[:0]
	x.LastSeen = x.LastSeen[:0]
	x.PhoneNumber = x.PhoneNumber[:0]
	x.ProfilePhoto = x.ProfilePhoto[:0]
	x.ForwardedMessage = x.ForwardedMessage[:0]
	x.Call = x.Call[:0]
	p.pool.Put(x)
}

var PoolUpdateAccountPrivacy = poolUpdateAccountPrivacy{}

const C_UpdateLabelItemsAdded int64 = 2552510165

type poolUpdateLabelItemsAdded struct {
	pool sync.Pool
}

func (p *poolUpdateLabelItemsAdded) Get() *UpdateLabelItemsAdded {
	x, ok := p.pool.Get().(*UpdateLabelItemsAdded)
	if !ok {
		return &UpdateLabelItemsAdded{}
	}
	return x
}

func (p *poolUpdateLabelItemsAdded) Put(x *UpdateLabelItemsAdded) {
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	if x.Peer != nil {
		PoolPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MessageIDs = x.MessageIDs[:0]
	x.LabelIDs = x.LabelIDs[:0]
	x.Labels = x.Labels[:0]
	p.pool.Put(x)
}

var PoolUpdateLabelItemsAdded = poolUpdateLabelItemsAdded{}

const C_UpdateLabelItemsRemoved int64 = 3223106630

type poolUpdateLabelItemsRemoved struct {
	pool sync.Pool
}

func (p *poolUpdateLabelItemsRemoved) Get() *UpdateLabelItemsRemoved {
	x, ok := p.pool.Get().(*UpdateLabelItemsRemoved)
	if !ok {
		return &UpdateLabelItemsRemoved{}
	}
	return x
}

func (p *poolUpdateLabelItemsRemoved) Put(x *UpdateLabelItemsRemoved) {
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	if x.Peer != nil {
		PoolPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MessageIDs = x.MessageIDs[:0]
	x.LabelIDs = x.LabelIDs[:0]
	x.Labels = x.Labels[:0]
	p.pool.Put(x)
}

var PoolUpdateLabelItemsRemoved = poolUpdateLabelItemsRemoved{}

const C_UpdateLabelSet int64 = 3098156256

type poolUpdateLabelSet struct {
	pool sync.Pool
}

func (p *poolUpdateLabelSet) Get() *UpdateLabelSet {
	x, ok := p.pool.Get().(*UpdateLabelSet)
	if !ok {
		return &UpdateLabelSet{}
	}
	return x
}

func (p *poolUpdateLabelSet) Put(x *UpdateLabelSet) {
	x.UCount = 0
	x.UpdateID = 0
	x.Labels = x.Labels[:0]
	p.pool.Put(x)
}

var PoolUpdateLabelSet = poolUpdateLabelSet{}

const C_UpdateLabelDeleted int64 = 2364090608

type poolUpdateLabelDeleted struct {
	pool sync.Pool
}

func (p *poolUpdateLabelDeleted) Get() *UpdateLabelDeleted {
	x, ok := p.pool.Get().(*UpdateLabelDeleted)
	if !ok {
		return &UpdateLabelDeleted{}
	}
	return x
}

func (p *poolUpdateLabelDeleted) Put(x *UpdateLabelDeleted) {
	x.UCount = 0
	x.UpdateID = 0
	x.LabelIDs = x.LabelIDs[:0]
	p.pool.Put(x)
}

var PoolUpdateLabelDeleted = poolUpdateLabelDeleted{}

const C_UpdateUserBlocked int64 = 2570026653

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
	x.UCount = 0
	x.UpdateID = 0
	x.UserID = 0
	x.Blocked = false
	p.pool.Put(x)
}

var PoolUpdateUserBlocked = poolUpdateUserBlocked{}

const C_UpdateMessagePoll int64 = 1349949010

type poolUpdateMessagePoll struct {
	pool sync.Pool
}

func (p *poolUpdateMessagePoll) Get() *UpdateMessagePoll {
	x, ok := p.pool.Get().(*UpdateMessagePoll)
	if !ok {
		return &UpdateMessagePoll{}
	}
	return x
}

func (p *poolUpdateMessagePoll) Put(x *UpdateMessagePoll) {
	x.UCount = 0
	x.UpdateID = 0
	x.PollID = 0
	if x.Poll != nil {
		PoolMediaPoll.Put(x.Poll)
		x.Poll = nil
	}
	if x.Results != nil {
		PoolPollResults.Put(x.Results)
		x.Results = nil
	}
	p.pool.Put(x)
}

var PoolUpdateMessagePoll = poolUpdateMessagePoll{}

const C_UpdateBotCallbackQuery int64 = 2133244656

type poolUpdateBotCallbackQuery struct {
	pool sync.Pool
}

func (p *poolUpdateBotCallbackQuery) Get() *UpdateBotCallbackQuery {
	x, ok := p.pool.Get().(*UpdateBotCallbackQuery)
	if !ok {
		return &UpdateBotCallbackQuery{}
	}
	return x
}

func (p *poolUpdateBotCallbackQuery) Put(x *UpdateBotCallbackQuery) {
	x.UCount = 0
	x.UpdateID = 0
	x.QueryID = 0
	x.UserID = 0
	if x.Peer != nil {
		PoolPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MessageID = 0
	x.Data = x.Data[:0]
	p.pool.Put(x)
}

var PoolUpdateBotCallbackQuery = poolUpdateBotCallbackQuery{}

const C_UpdateBotInlineQuery int64 = 2949144765

type poolUpdateBotInlineQuery struct {
	pool sync.Pool
}

func (p *poolUpdateBotInlineQuery) Get() *UpdateBotInlineQuery {
	x, ok := p.pool.Get().(*UpdateBotInlineQuery)
	if !ok {
		return &UpdateBotInlineQuery{}
	}
	return x
}

func (p *poolUpdateBotInlineQuery) Put(x *UpdateBotInlineQuery) {
	x.UCount = 0
	x.UpdateID = 0
	x.QueryID = 0
	x.UserID = 0
	if x.Peer != nil {
		PoolPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.Query = ""
	x.Offset = ""
	if x.Geo != nil {
		PoolGeoLocation.Put(x.Geo)
		x.Geo = nil
	}
	p.pool.Put(x)
}

var PoolUpdateBotInlineQuery = poolUpdateBotInlineQuery{}

const C_UpdateBotInlineSend int64 = 1813939863

type poolUpdateBotInlineSend struct {
	pool sync.Pool
}

func (p *poolUpdateBotInlineSend) Get() *UpdateBotInlineSend {
	x, ok := p.pool.Get().(*UpdateBotInlineSend)
	if !ok {
		return &UpdateBotInlineSend{}
	}
	return x
}

func (p *poolUpdateBotInlineSend) Put(x *UpdateBotInlineSend) {
	x.UCount = 0
	x.UpdateID = 0
	x.UserID = 0
	x.Query = ""
	x.ResultID = ""
	if x.Geo != nil {
		PoolGeoLocation.Put(x.Geo)
		x.Geo = nil
	}
	p.pool.Put(x)
}

var PoolUpdateBotInlineSend = poolUpdateBotInlineSend{}

const C_UpdateTeamCreated int64 = 3792600641

type poolUpdateTeamCreated struct {
	pool sync.Pool
}

func (p *poolUpdateTeamCreated) Get() *UpdateTeamCreated {
	x, ok := p.pool.Get().(*UpdateTeamCreated)
	if !ok {
		return &UpdateTeamCreated{}
	}
	return x
}

func (p *poolUpdateTeamCreated) Put(x *UpdateTeamCreated) {
	x.UCount = 0
	x.UpdateID = 0
	if x.Team != nil {
		PoolTeam.Put(x.Team)
		x.Team = nil
	}
	p.pool.Put(x)
}

var PoolUpdateTeamCreated = poolUpdateTeamCreated{}

const C_UpdateTeamMemberAdded int64 = 1371743118

type poolUpdateTeamMemberAdded struct {
	pool sync.Pool
}

func (p *poolUpdateTeamMemberAdded) Get() *UpdateTeamMemberAdded {
	x, ok := p.pool.Get().(*UpdateTeamMemberAdded)
	if !ok {
		return &UpdateTeamMemberAdded{}
	}
	return x
}

func (p *poolUpdateTeamMemberAdded) Put(x *UpdateTeamMemberAdded) {
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	if x.User != nil {
		PoolUser.Put(x.User)
		x.User = nil
	}
	if x.Contact != nil {
		PoolContactUser.Put(x.Contact)
		x.Contact = nil
	}
	x.AdderID = 0
	x.Hash = 0
	p.pool.Put(x)
}

var PoolUpdateTeamMemberAdded = poolUpdateTeamMemberAdded{}

const C_UpdateTeamMemberRemoved int64 = 4102954453

type poolUpdateTeamMemberRemoved struct {
	pool sync.Pool
}

func (p *poolUpdateTeamMemberRemoved) Get() *UpdateTeamMemberRemoved {
	x, ok := p.pool.Get().(*UpdateTeamMemberRemoved)
	if !ok {
		return &UpdateTeamMemberRemoved{}
	}
	return x
}

func (p *poolUpdateTeamMemberRemoved) Put(x *UpdateTeamMemberRemoved) {
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	x.UserID = 0
	x.RemoverID = 0
	x.Hash = 0
	p.pool.Put(x)
}

var PoolUpdateTeamMemberRemoved = poolUpdateTeamMemberRemoved{}

const C_UpdateTeamMemberStatus int64 = 1178766456

type poolUpdateTeamMemberStatus struct {
	pool sync.Pool
}

func (p *poolUpdateTeamMemberStatus) Get() *UpdateTeamMemberStatus {
	x, ok := p.pool.Get().(*UpdateTeamMemberStatus)
	if !ok {
		return &UpdateTeamMemberStatus{}
	}
	return x
}

func (p *poolUpdateTeamMemberStatus) Put(x *UpdateTeamMemberStatus) {
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	x.ChangerID = 0
	x.Admin = false
	p.pool.Put(x)
}

var PoolUpdateTeamMemberStatus = poolUpdateTeamMemberStatus{}

const C_UpdateTeamPhoto int64 = 2939683506

type poolUpdateTeamPhoto struct {
	pool sync.Pool
}

func (p *poolUpdateTeamPhoto) Get() *UpdateTeamPhoto {
	x, ok := p.pool.Get().(*UpdateTeamPhoto)
	if !ok {
		return &UpdateTeamPhoto{}
	}
	return x
}

func (p *poolUpdateTeamPhoto) Put(x *UpdateTeamPhoto) {
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	if x.Photo != nil {
		PoolTeamPhoto.Put(x.Photo)
		x.Photo = nil
	}
	p.pool.Put(x)
}

var PoolUpdateTeamPhoto = poolUpdateTeamPhoto{}

const C_UpdateTeam int64 = 931901879

type poolUpdateTeam struct {
	pool sync.Pool
}

func (p *poolUpdateTeam) Get() *UpdateTeam {
	x, ok := p.pool.Get().(*UpdateTeam)
	if !ok {
		return &UpdateTeam{}
	}
	return x
}

func (p *poolUpdateTeam) Put(x *UpdateTeam) {
	x.UCount = 0
	x.UpdateID = 0
	x.TeamID = 0
	x.Name = ""
	p.pool.Put(x)
}

var PoolUpdateTeam = poolUpdateTeam{}

const C_UpdateCommunityMessage int64 = 2394032357

type poolUpdateCommunityMessage struct {
	pool sync.Pool
}

func (p *poolUpdateCommunityMessage) Get() *UpdateCommunityMessage {
	x, ok := p.pool.Get().(*UpdateCommunityMessage)
	if !ok {
		return &UpdateCommunityMessage{}
	}
	return x
}

func (p *poolUpdateCommunityMessage) Put(x *UpdateCommunityMessage) {
	x.TeamID = 0
	x.SenderID = 0
	x.ReceiverID = 0
	x.Body = ""
	x.CreatedOn = 0
	x.GlobalMsgID = 0
	x.Entities = x.Entities[:0]
	x.SenderMsgID = 0
	p.pool.Put(x)
}

var PoolUpdateCommunityMessage = poolUpdateCommunityMessage{}

const C_UpdateCommunityReadOutbox int64 = 3478641786

type poolUpdateCommunityReadOutbox struct {
	pool sync.Pool
}

func (p *poolUpdateCommunityReadOutbox) Get() *UpdateCommunityReadOutbox {
	x, ok := p.pool.Get().(*UpdateCommunityReadOutbox)
	if !ok {
		return &UpdateCommunityReadOutbox{}
	}
	return x
}

func (p *poolUpdateCommunityReadOutbox) Put(x *UpdateCommunityReadOutbox) {
	x.TeamID = 0
	x.SenderID = 0
	x.ReceiverID = 0
	x.SenderMsgID = 0
	p.pool.Put(x)
}

var PoolUpdateCommunityReadOutbox = poolUpdateCommunityReadOutbox{}

const C_UpdateCommunityTyping int64 = 114872457

type poolUpdateCommunityTyping struct {
	pool sync.Pool
}

func (p *poolUpdateCommunityTyping) Get() *UpdateCommunityTyping {
	x, ok := p.pool.Get().(*UpdateCommunityTyping)
	if !ok {
		return &UpdateCommunityTyping{}
	}
	return x
}

func (p *poolUpdateCommunityTyping) Put(x *UpdateCommunityTyping) {
	x.TeamID = 0
	x.SenderID = 0
	x.ReceiverID = 0
	x.Action = 0
	p.pool.Put(x)
}

var PoolUpdateCommunityTyping = poolUpdateCommunityTyping{}

const C_UpdateReaction int64 = 2547814946

type poolUpdateReaction struct {
	pool sync.Pool
}

func (p *poolUpdateReaction) Get() *UpdateReaction {
	x, ok := p.pool.Get().(*UpdateReaction)
	if !ok {
		return &UpdateReaction{}
	}
	return x
}

func (p *poolUpdateReaction) Put(x *UpdateReaction) {
	x.UCount = 0
	x.UpdateID = 0
	x.MessageID = 0
	x.Counter = x.Counter[:0]
	x.TeamID = 0
	if x.Peer != nil {
		PoolPeer.Put(x.Peer)
		x.Peer = nil
	}
	if x.Sender != nil {
		PoolUser.Put(x.Sender)
		x.Sender = nil
	}
	x.YourReactions = x.YourReactions[:0]
	x.Reaction = ""
	p.pool.Put(x)
}

var PoolUpdateReaction = poolUpdateReaction{}

const C_UpdateCalendarEventAdded int64 = 4108732989

type poolUpdateCalendarEventAdded struct {
	pool sync.Pool
}

func (p *poolUpdateCalendarEventAdded) Get() *UpdateCalendarEventAdded {
	x, ok := p.pool.Get().(*UpdateCalendarEventAdded)
	if !ok {
		return &UpdateCalendarEventAdded{}
	}
	return x
}

func (p *poolUpdateCalendarEventAdded) Put(x *UpdateCalendarEventAdded) {
	x.UCount = 0
	x.UpdateID = 0
	if x.Event != nil {
		PoolCalendarEvent.Put(x.Event)
		x.Event = nil
	}
	p.pool.Put(x)
}

var PoolUpdateCalendarEventAdded = poolUpdateCalendarEventAdded{}

const C_UpdateCalendarEventRemoved int64 = 252222583

type poolUpdateCalendarEventRemoved struct {
	pool sync.Pool
}

func (p *poolUpdateCalendarEventRemoved) Get() *UpdateCalendarEventRemoved {
	x, ok := p.pool.Get().(*UpdateCalendarEventRemoved)
	if !ok {
		return &UpdateCalendarEventRemoved{}
	}
	return x
}

func (p *poolUpdateCalendarEventRemoved) Put(x *UpdateCalendarEventRemoved) {
	x.UCount = 0
	x.UpdateID = 0
	x.EventID = 0
	p.pool.Put(x)
}

var PoolUpdateCalendarEventRemoved = poolUpdateCalendarEventRemoved{}

const C_UpdateCalendarEventEdited int64 = 2907013722

type poolUpdateCalendarEventEdited struct {
	pool sync.Pool
}

func (p *poolUpdateCalendarEventEdited) Get() *UpdateCalendarEventEdited {
	x, ok := p.pool.Get().(*UpdateCalendarEventEdited)
	if !ok {
		return &UpdateCalendarEventEdited{}
	}
	return x
}

func (p *poolUpdateCalendarEventEdited) Put(x *UpdateCalendarEventEdited) {
	x.UCount = 0
	x.UpdateID = 0
	if x.Event != nil {
		PoolCalendarEvent.Put(x.Event)
		x.Event = nil
	}
	p.pool.Put(x)
}

var PoolUpdateCalendarEventEdited = poolUpdateCalendarEventEdited{}

const C_UpdateRedirect int64 = 4026993662

type poolUpdateRedirect struct {
	pool sync.Pool
}

func (p *poolUpdateRedirect) Get() *UpdateRedirect {
	x, ok := p.pool.Get().(*UpdateRedirect)
	if !ok {
		return &UpdateRedirect{}
	}
	return x
}

func (p *poolUpdateRedirect) Put(x *UpdateRedirect) {
	x.UCount = 0
	x.UpdateID = 0
	x.Redirects = x.Redirects[:0]
	x.Empty = false
	p.pool.Put(x)
}

var PoolUpdateRedirect = poolUpdateRedirect{}

const C_ClientRedirect int64 = 3851114106

type poolClientRedirect struct {
	pool sync.Pool
}

func (p *poolClientRedirect) Get() *ClientRedirect {
	x, ok := p.pool.Get().(*ClientRedirect)
	if !ok {
		return &ClientRedirect{}
	}
	return x
}

func (p *poolClientRedirect) Put(x *ClientRedirect) {
	x.HostPort = ""
	x.Permanent = false
	x.Target = 0
	x.Alternatives = x.Alternatives[:0]
	p.pool.Put(x)
}

var PoolClientRedirect = poolClientRedirect{}

const C_UpdatePhoneCall int64 = 2953098428

type poolUpdatePhoneCall struct {
	pool sync.Pool
}

func (p *poolUpdatePhoneCall) Get() *UpdatePhoneCall {
	x, ok := p.pool.Get().(*UpdatePhoneCall)
	if !ok {
		return &UpdatePhoneCall{}
	}
	return x
}

func (p *poolUpdatePhoneCall) Put(x *UpdatePhoneCall) {
	x.UCount = 0
	x.TeamID = 0
	x.PeerID = 0
	x.PeerType = 0
	x.CallID = 0
	x.UserID = 0
	x.AccessHash = 0
	x.Action = 0
	x.ActionData = x.ActionData[:0]
	p.pool.Put(x)
}

var PoolUpdatePhoneCall = poolUpdatePhoneCall{}

func init() {
	registry.RegisterConstructor(1632290153, "msg.UpdateGetState")
	registry.RegisterConstructor(3467307691, "msg.UpdateGetDifference")
	registry.RegisterConstructor(2937859860, "msg.UpdateDifference")
	registry.RegisterConstructor(3588398962, "msg.UpdateTooLong")
	registry.RegisterConstructor(2070778422, "msg.UpdateState")
	registry.RegisterConstructor(1764208092, "msg.UpdateMessageID")
	registry.RegisterConstructor(75740112, "msg.UpdateNewMessage")
	registry.RegisterConstructor(2202915150, "msg.UpdateMessageEdited")
	registry.RegisterConstructor(1003091958, "msg.UpdateMessagesDeleted")
	registry.RegisterConstructor(4013107819, "msg.UpdateReadHistoryInbox")
	registry.RegisterConstructor(4011050865, "msg.UpdateReadHistoryOutbox")
	registry.RegisterConstructor(2761988296, "msg.UpdateMessagePinned")
	registry.RegisterConstructor(3261004099, "msg.UpdateUserTyping")
	registry.RegisterConstructor(1752961652, "msg.UpdateUserStatus")
	registry.RegisterConstructor(3411383202, "msg.UpdateUsername")
	registry.RegisterConstructor(72923648, "msg.UpdateUserPhoto")
	registry.RegisterConstructor(3766115140, "msg.UpdateNotifySettings")
	registry.RegisterConstructor(3544906637, "msg.UpdateGroupParticipantAdd")
	registry.RegisterConstructor(921388023, "msg.UpdateGroupParticipantDeleted")
	registry.RegisterConstructor(4102007577, "msg.UpdateGroupParticipantAdmin")
	registry.RegisterConstructor(1878951933, "msg.UpdateGroupAdmins")
	registry.RegisterConstructor(3710117357, "msg.UpdateGroupPhoto")
	registry.RegisterConstructor(256065898, "msg.UpdateReadMessagesContents")
	registry.RegisterConstructor(1770313879, "msg.UpdateAuthorizationReset")
	registry.RegisterConstructor(2643668944, "msg.UpdateDraftMessage")
	registry.RegisterConstructor(3294904674, "msg.UpdateDraftMessageCleared")
	registry.RegisterConstructor(1569664568, "msg.UpdateDialogPinned")
	registry.RegisterConstructor(4008682179, "msg.UpdateDialogPinnedReorder")
	registry.RegisterConstructor(2013786192, "msg.UpdateAccountPrivacy")
	registry.RegisterConstructor(2552510165, "msg.UpdateLabelItemsAdded")
	registry.RegisterConstructor(3223106630, "msg.UpdateLabelItemsRemoved")
	registry.RegisterConstructor(3098156256, "msg.UpdateLabelSet")
	registry.RegisterConstructor(2364090608, "msg.UpdateLabelDeleted")
	registry.RegisterConstructor(2570026653, "msg.UpdateUserBlocked")
	registry.RegisterConstructor(1349949010, "msg.UpdateMessagePoll")
	registry.RegisterConstructor(2133244656, "msg.UpdateBotCallbackQuery")
	registry.RegisterConstructor(2949144765, "msg.UpdateBotInlineQuery")
	registry.RegisterConstructor(1813939863, "msg.UpdateBotInlineSend")
	registry.RegisterConstructor(3792600641, "msg.UpdateTeamCreated")
	registry.RegisterConstructor(1371743118, "msg.UpdateTeamMemberAdded")
	registry.RegisterConstructor(4102954453, "msg.UpdateTeamMemberRemoved")
	registry.RegisterConstructor(1178766456, "msg.UpdateTeamMemberStatus")
	registry.RegisterConstructor(2939683506, "msg.UpdateTeamPhoto")
	registry.RegisterConstructor(931901879, "msg.UpdateTeam")
	registry.RegisterConstructor(2394032357, "msg.UpdateCommunityMessage")
	registry.RegisterConstructor(3478641786, "msg.UpdateCommunityReadOutbox")
	registry.RegisterConstructor(114872457, "msg.UpdateCommunityTyping")
	registry.RegisterConstructor(2547814946, "msg.UpdateReaction")
	registry.RegisterConstructor(4108732989, "msg.UpdateCalendarEventAdded")
	registry.RegisterConstructor(252222583, "msg.UpdateCalendarEventRemoved")
	registry.RegisterConstructor(2907013722, "msg.UpdateCalendarEventEdited")
	registry.RegisterConstructor(4026993662, "msg.UpdateRedirect")
	registry.RegisterConstructor(3851114106, "msg.ClientRedirect")
	registry.RegisterConstructor(2953098428, "msg.UpdatePhoneCall")
}

func (x *UpdateGetState) DeepCopy(z *UpdateGetState) {
}

func (x *UpdateGetDifference) DeepCopy(z *UpdateGetDifference) {
	z.From = x.From
	z.Limit = x.Limit
}

func (x *UpdateDifference) DeepCopy(z *UpdateDifference) {
	z.More = x.More
	z.MaxUpdateID = x.MaxUpdateID
	z.MinUpdateID = x.MinUpdateID
	for idx := range x.Updates {
		if x.Updates[idx] != nil {
			xx := PoolUpdateEnvelope.Get()
			x.Updates[idx].DeepCopy(xx)
			z.Updates = append(z.Updates, xx)
		}
	}
	for idx := range x.Users {
		if x.Users[idx] != nil {
			xx := PoolUser.Get()
			x.Users[idx].DeepCopy(xx)
			z.Users = append(z.Users, xx)
		}
	}
	for idx := range x.Groups {
		if x.Groups[idx] != nil {
			xx := PoolGroup.Get()
			x.Groups[idx].DeepCopy(xx)
			z.Groups = append(z.Groups, xx)
		}
	}
	z.CurrentUpdateID = x.CurrentUpdateID
}

func (x *UpdateTooLong) DeepCopy(z *UpdateTooLong) {
}

func (x *UpdateState) DeepCopy(z *UpdateState) {
	z.UpdateID = x.UpdateID
}

func (x *UpdateMessageID) DeepCopy(z *UpdateMessageID) {
	z.UCount = x.UCount
	z.MessageID = x.MessageID
	z.RandomID = x.RandomID
}

func (x *UpdateNewMessage) DeepCopy(z *UpdateNewMessage) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	if x.Message != nil {
		z.Message = PoolUserMessage.Get()
		x.Message.DeepCopy(z.Message)
	}
	if x.Sender != nil {
		z.Sender = PoolUser.Get()
		x.Sender.DeepCopy(z.Sender)
	}
	z.AccessHash = x.AccessHash
	z.SenderRefID = x.SenderRefID
}

func (x *UpdateMessageEdited) DeepCopy(z *UpdateMessageEdited) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	if x.Message != nil {
		z.Message = PoolUserMessage.Get()
		x.Message.DeepCopy(z.Message)
	}
}

func (x *UpdateMessagesDeleted) DeepCopy(z *UpdateMessagesDeleted) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	z.MessageIDs = append(z.MessageIDs[:0], x.MessageIDs...)
	if x.Peer != nil {
		z.Peer = PoolPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
}

func (x *UpdateReadHistoryInbox) DeepCopy(z *UpdateReadHistoryInbox) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	if x.Peer != nil {
		z.Peer = PoolPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MaxID = x.MaxID
}

func (x *UpdateReadHistoryOutbox) DeepCopy(z *UpdateReadHistoryOutbox) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	if x.Peer != nil {
		z.Peer = PoolPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MaxID = x.MaxID
	z.UserID = x.UserID
}

func (x *UpdateMessagePinned) DeepCopy(z *UpdateMessagePinned) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	if x.Peer != nil {
		z.Peer = PoolPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.UserID = x.UserID
	z.MsgID = x.MsgID
	z.Version = x.Version
}

func (x *UpdateUserTyping) DeepCopy(z *UpdateUserTyping) {
	z.UCount = x.UCount
	z.TeamID = x.TeamID
	z.UserID = x.UserID
	z.Action = x.Action
	z.PeerID = x.PeerID
	z.PeerType = x.PeerType
}

func (x *UpdateUserStatus) DeepCopy(z *UpdateUserStatus) {
	z.UCount = x.UCount
	z.UserID = x.UserID
	z.Status = x.Status
}

func (x *UpdateUsername) DeepCopy(z *UpdateUsername) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.UserID = x.UserID
	z.Username = x.Username
	z.FirstName = x.FirstName
	z.LastName = x.LastName
	z.Bio = x.Bio
	z.Phone = x.Phone
}

func (x *UpdateUserPhoto) DeepCopy(z *UpdateUserPhoto) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.UserID = x.UserID
	if x.Photo != nil {
		z.Photo = PoolUserPhoto.Get()
		x.Photo.DeepCopy(z.Photo)
	}
	z.PhotoID = x.PhotoID
	z.DeletedPhotoIDs = append(z.DeletedPhotoIDs[:0], x.DeletedPhotoIDs...)
}

func (x *UpdateNotifySettings) DeepCopy(z *UpdateNotifySettings) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	z.UserID = x.UserID
	if x.NotifyPeer != nil {
		z.NotifyPeer = PoolPeer.Get()
		x.NotifyPeer.DeepCopy(z.NotifyPeer)
	}
	if x.Settings != nil {
		z.Settings = PoolPeerNotifySettings.Get()
		x.Settings.DeepCopy(z.Settings)
	}
}

func (x *UpdateGroupParticipantAdd) DeepCopy(z *UpdateGroupParticipantAdd) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.GroupID = x.GroupID
	z.UserID = x.UserID
	z.InviterID = x.InviterID
	z.Date = x.Date
}

func (x *UpdateGroupParticipantDeleted) DeepCopy(z *UpdateGroupParticipantDeleted) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.GroupID = x.GroupID
	z.UserID = x.UserID
}

func (x *UpdateGroupParticipantAdmin) DeepCopy(z *UpdateGroupParticipantAdmin) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.GroupID = x.GroupID
	z.UserID = x.UserID
	z.IsAdmin = x.IsAdmin
}

func (x *UpdateGroupAdmins) DeepCopy(z *UpdateGroupAdmins) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.GroupID = x.GroupID
	z.AdminEnabled = x.AdminEnabled
}

func (x *UpdateGroupPhoto) DeepCopy(z *UpdateGroupPhoto) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.GroupID = x.GroupID
	if x.Photo != nil {
		z.Photo = PoolGroupPhoto.Get()
		x.Photo.DeepCopy(z.Photo)
	}
	z.PhotoID = x.PhotoID
}

func (x *UpdateReadMessagesContents) DeepCopy(z *UpdateReadMessagesContents) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	z.MessageIDs = append(z.MessageIDs[:0], x.MessageIDs...)
	if x.Peer != nil {
		z.Peer = PoolPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
}

func (x *UpdateAuthorizationReset) DeepCopy(z *UpdateAuthorizationReset) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
}

func (x *UpdateDraftMessage) DeepCopy(z *UpdateDraftMessage) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	if x.Message != nil {
		z.Message = PoolDraftMessage.Get()
		x.Message.DeepCopy(z.Message)
	}
}

func (x *UpdateDraftMessageCleared) DeepCopy(z *UpdateDraftMessageCleared) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	if x.Peer != nil {
		z.Peer = PoolPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
}

func (x *UpdateDialogPinned) DeepCopy(z *UpdateDialogPinned) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	if x.Peer != nil {
		z.Peer = PoolPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.Pinned = x.Pinned
}

func (x *UpdateDialogPinnedReorder) DeepCopy(z *UpdateDialogPinnedReorder) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	for idx := range x.Peer {
		if x.Peer[idx] != nil {
			xx := PoolPeer.Get()
			x.Peer[idx].DeepCopy(xx)
			z.Peer = append(z.Peer, xx)
		}
	}
}

func (x *UpdateAccountPrivacy) DeepCopy(z *UpdateAccountPrivacy) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	for idx := range x.ChatInvite {
		if x.ChatInvite[idx] != nil {
			xx := PoolPrivacyRule.Get()
			x.ChatInvite[idx].DeepCopy(xx)
			z.ChatInvite = append(z.ChatInvite, xx)
		}
	}
	for idx := range x.LastSeen {
		if x.LastSeen[idx] != nil {
			xx := PoolPrivacyRule.Get()
			x.LastSeen[idx].DeepCopy(xx)
			z.LastSeen = append(z.LastSeen, xx)
		}
	}
	for idx := range x.PhoneNumber {
		if x.PhoneNumber[idx] != nil {
			xx := PoolPrivacyRule.Get()
			x.PhoneNumber[idx].DeepCopy(xx)
			z.PhoneNumber = append(z.PhoneNumber, xx)
		}
	}
	for idx := range x.ProfilePhoto {
		if x.ProfilePhoto[idx] != nil {
			xx := PoolPrivacyRule.Get()
			x.ProfilePhoto[idx].DeepCopy(xx)
			z.ProfilePhoto = append(z.ProfilePhoto, xx)
		}
	}
	for idx := range x.ForwardedMessage {
		if x.ForwardedMessage[idx] != nil {
			xx := PoolPrivacyRule.Get()
			x.ForwardedMessage[idx].DeepCopy(xx)
			z.ForwardedMessage = append(z.ForwardedMessage, xx)
		}
	}
	for idx := range x.Call {
		if x.Call[idx] != nil {
			xx := PoolPrivacyRule.Get()
			x.Call[idx].DeepCopy(xx)
			z.Call = append(z.Call, xx)
		}
	}
}

func (x *UpdateLabelItemsAdded) DeepCopy(z *UpdateLabelItemsAdded) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	if x.Peer != nil {
		z.Peer = PoolPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MessageIDs = append(z.MessageIDs[:0], x.MessageIDs...)
	z.LabelIDs = append(z.LabelIDs[:0], x.LabelIDs...)
	for idx := range x.Labels {
		if x.Labels[idx] != nil {
			xx := PoolLabel.Get()
			x.Labels[idx].DeepCopy(xx)
			z.Labels = append(z.Labels, xx)
		}
	}
}

func (x *UpdateLabelItemsRemoved) DeepCopy(z *UpdateLabelItemsRemoved) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	if x.Peer != nil {
		z.Peer = PoolPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MessageIDs = append(z.MessageIDs[:0], x.MessageIDs...)
	z.LabelIDs = append(z.LabelIDs[:0], x.LabelIDs...)
	for idx := range x.Labels {
		if x.Labels[idx] != nil {
			xx := PoolLabel.Get()
			x.Labels[idx].DeepCopy(xx)
			z.Labels = append(z.Labels, xx)
		}
	}
}

func (x *UpdateLabelSet) DeepCopy(z *UpdateLabelSet) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	for idx := range x.Labels {
		if x.Labels[idx] != nil {
			xx := PoolLabel.Get()
			x.Labels[idx].DeepCopy(xx)
			z.Labels = append(z.Labels, xx)
		}
	}
}

func (x *UpdateLabelDeleted) DeepCopy(z *UpdateLabelDeleted) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.LabelIDs = append(z.LabelIDs[:0], x.LabelIDs...)
}

func (x *UpdateUserBlocked) DeepCopy(z *UpdateUserBlocked) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.UserID = x.UserID
	z.Blocked = x.Blocked
}

func (x *UpdateMessagePoll) DeepCopy(z *UpdateMessagePoll) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.PollID = x.PollID
	if x.Poll != nil {
		z.Poll = PoolMediaPoll.Get()
		x.Poll.DeepCopy(z.Poll)
	}
	if x.Results != nil {
		z.Results = PoolPollResults.Get()
		x.Results.DeepCopy(z.Results)
	}
}

func (x *UpdateBotCallbackQuery) DeepCopy(z *UpdateBotCallbackQuery) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.QueryID = x.QueryID
	z.UserID = x.UserID
	if x.Peer != nil {
		z.Peer = PoolPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MessageID = x.MessageID
	z.Data = append(z.Data[:0], x.Data...)
}

func (x *UpdateBotInlineQuery) DeepCopy(z *UpdateBotInlineQuery) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.QueryID = x.QueryID
	z.UserID = x.UserID
	if x.Peer != nil {
		z.Peer = PoolPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.Query = x.Query
	z.Offset = x.Offset
	if x.Geo != nil {
		z.Geo = PoolGeoLocation.Get()
		x.Geo.DeepCopy(z.Geo)
	}
}

func (x *UpdateBotInlineSend) DeepCopy(z *UpdateBotInlineSend) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.UserID = x.UserID
	z.Query = x.Query
	z.ResultID = x.ResultID
	if x.Geo != nil {
		z.Geo = PoolGeoLocation.Get()
		x.Geo.DeepCopy(z.Geo)
	}
}

func (x *UpdateTeamCreated) DeepCopy(z *UpdateTeamCreated) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	if x.Team != nil {
		z.Team = PoolTeam.Get()
		x.Team.DeepCopy(z.Team)
	}
}

func (x *UpdateTeamMemberAdded) DeepCopy(z *UpdateTeamMemberAdded) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	if x.User != nil {
		z.User = PoolUser.Get()
		x.User.DeepCopy(z.User)
	}
	if x.Contact != nil {
		z.Contact = PoolContactUser.Get()
		x.Contact.DeepCopy(z.Contact)
	}
	z.AdderID = x.AdderID
	z.Hash = x.Hash
}

func (x *UpdateTeamMemberRemoved) DeepCopy(z *UpdateTeamMemberRemoved) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	z.UserID = x.UserID
	z.RemoverID = x.RemoverID
	z.Hash = x.Hash
}

func (x *UpdateTeamMemberStatus) DeepCopy(z *UpdateTeamMemberStatus) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	z.ChangerID = x.ChangerID
	z.Admin = x.Admin
}

func (x *UpdateTeamPhoto) DeepCopy(z *UpdateTeamPhoto) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	if x.Photo != nil {
		z.Photo = PoolTeamPhoto.Get()
		x.Photo.DeepCopy(z.Photo)
	}
}

func (x *UpdateTeam) DeepCopy(z *UpdateTeam) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.TeamID = x.TeamID
	z.Name = x.Name
}

func (x *UpdateCommunityMessage) DeepCopy(z *UpdateCommunityMessage) {
	z.TeamID = x.TeamID
	z.SenderID = x.SenderID
	z.ReceiverID = x.ReceiverID
	z.Body = x.Body
	z.CreatedOn = x.CreatedOn
	z.GlobalMsgID = x.GlobalMsgID
	for idx := range x.Entities {
		if x.Entities[idx] != nil {
			xx := PoolMessageEntity.Get()
			x.Entities[idx].DeepCopy(xx)
			z.Entities = append(z.Entities, xx)
		}
	}
	z.SenderMsgID = x.SenderMsgID
}

func (x *UpdateCommunityReadOutbox) DeepCopy(z *UpdateCommunityReadOutbox) {
	z.TeamID = x.TeamID
	z.SenderID = x.SenderID
	z.ReceiverID = x.ReceiverID
	z.SenderMsgID = x.SenderMsgID
}

func (x *UpdateCommunityTyping) DeepCopy(z *UpdateCommunityTyping) {
	z.TeamID = x.TeamID
	z.SenderID = x.SenderID
	z.ReceiverID = x.ReceiverID
	z.Action = x.Action
}

func (x *UpdateReaction) DeepCopy(z *UpdateReaction) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.MessageID = x.MessageID
	for idx := range x.Counter {
		if x.Counter[idx] != nil {
			xx := PoolReactionCounter.Get()
			x.Counter[idx].DeepCopy(xx)
			z.Counter = append(z.Counter, xx)
		}
	}
	z.TeamID = x.TeamID
	if x.Peer != nil {
		z.Peer = PoolPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	if x.Sender != nil {
		z.Sender = PoolUser.Get()
		x.Sender.DeepCopy(z.Sender)
	}
	z.YourReactions = append(z.YourReactions[:0], x.YourReactions...)
	z.Reaction = x.Reaction
}

func (x *UpdateCalendarEventAdded) DeepCopy(z *UpdateCalendarEventAdded) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	if x.Event != nil {
		z.Event = PoolCalendarEvent.Get()
		x.Event.DeepCopy(z.Event)
	}
}

func (x *UpdateCalendarEventRemoved) DeepCopy(z *UpdateCalendarEventRemoved) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	z.EventID = x.EventID
}

func (x *UpdateCalendarEventEdited) DeepCopy(z *UpdateCalendarEventEdited) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	if x.Event != nil {
		z.Event = PoolCalendarEvent.Get()
		x.Event.DeepCopy(z.Event)
	}
}

func (x *UpdateRedirect) DeepCopy(z *UpdateRedirect) {
	z.UCount = x.UCount
	z.UpdateID = x.UpdateID
	for idx := range x.Redirects {
		if x.Redirects[idx] != nil {
			xx := PoolClientRedirect.Get()
			x.Redirects[idx].DeepCopy(xx)
			z.Redirects = append(z.Redirects, xx)
		}
	}
	z.Empty = x.Empty
}

func (x *ClientRedirect) DeepCopy(z *ClientRedirect) {
	z.HostPort = x.HostPort
	z.Permanent = x.Permanent
	z.Target = x.Target
	z.Alternatives = append(z.Alternatives[:0], x.Alternatives...)
}

func (x *UpdatePhoneCall) DeepCopy(z *UpdatePhoneCall) {
	z.UCount = x.UCount
	z.TeamID = x.TeamID
	z.PeerID = x.PeerID
	z.PeerType = x.PeerType
	z.CallID = x.CallID
	z.UserID = x.UserID
	z.AccessHash = x.AccessHash
	z.Action = x.Action
	z.ActionData = append(z.ActionData[:0], x.ActionData...)
}

func (x *UpdateGetState) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateGetState, x)
}

func (x *UpdateGetDifference) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateGetDifference, x)
}

func (x *UpdateDifference) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateDifference, x)
}

func (x *UpdateTooLong) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateTooLong, x)
}

func (x *UpdateState) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateState, x)
}

func (x *UpdateMessageID) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateMessageID, x)
}

func (x *UpdateNewMessage) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateNewMessage, x)
}

func (x *UpdateMessageEdited) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateMessageEdited, x)
}

func (x *UpdateMessagesDeleted) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateMessagesDeleted, x)
}

func (x *UpdateReadHistoryInbox) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateReadHistoryInbox, x)
}

func (x *UpdateReadHistoryOutbox) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateReadHistoryOutbox, x)
}

func (x *UpdateMessagePinned) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateMessagePinned, x)
}

func (x *UpdateUserTyping) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateUserTyping, x)
}

func (x *UpdateUserStatus) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateUserStatus, x)
}

func (x *UpdateUsername) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateUsername, x)
}

func (x *UpdateUserPhoto) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateUserPhoto, x)
}

func (x *UpdateNotifySettings) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateNotifySettings, x)
}

func (x *UpdateGroupParticipantAdd) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateGroupParticipantAdd, x)
}

func (x *UpdateGroupParticipantDeleted) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateGroupParticipantDeleted, x)
}

func (x *UpdateGroupParticipantAdmin) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateGroupParticipantAdmin, x)
}

func (x *UpdateGroupAdmins) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateGroupAdmins, x)
}

func (x *UpdateGroupPhoto) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateGroupPhoto, x)
}

func (x *UpdateReadMessagesContents) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateReadMessagesContents, x)
}

func (x *UpdateAuthorizationReset) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateAuthorizationReset, x)
}

func (x *UpdateDraftMessage) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateDraftMessage, x)
}

func (x *UpdateDraftMessageCleared) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateDraftMessageCleared, x)
}

func (x *UpdateDialogPinned) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateDialogPinned, x)
}

func (x *UpdateDialogPinnedReorder) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateDialogPinnedReorder, x)
}

func (x *UpdateAccountPrivacy) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateAccountPrivacy, x)
}

func (x *UpdateLabelItemsAdded) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateLabelItemsAdded, x)
}

func (x *UpdateLabelItemsRemoved) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateLabelItemsRemoved, x)
}

func (x *UpdateLabelSet) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateLabelSet, x)
}

func (x *UpdateLabelDeleted) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateLabelDeleted, x)
}

func (x *UpdateUserBlocked) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateUserBlocked, x)
}

func (x *UpdateMessagePoll) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateMessagePoll, x)
}

func (x *UpdateBotCallbackQuery) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateBotCallbackQuery, x)
}

func (x *UpdateBotInlineQuery) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateBotInlineQuery, x)
}

func (x *UpdateBotInlineSend) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateBotInlineSend, x)
}

func (x *UpdateTeamCreated) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateTeamCreated, x)
}

func (x *UpdateTeamMemberAdded) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateTeamMemberAdded, x)
}

func (x *UpdateTeamMemberRemoved) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateTeamMemberRemoved, x)
}

func (x *UpdateTeamMemberStatus) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateTeamMemberStatus, x)
}

func (x *UpdateTeamPhoto) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateTeamPhoto, x)
}

func (x *UpdateTeam) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateTeam, x)
}

func (x *UpdateCommunityMessage) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateCommunityMessage, x)
}

func (x *UpdateCommunityReadOutbox) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateCommunityReadOutbox, x)
}

func (x *UpdateCommunityTyping) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateCommunityTyping, x)
}

func (x *UpdateReaction) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateReaction, x)
}

func (x *UpdateCalendarEventAdded) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateCalendarEventAdded, x)
}

func (x *UpdateCalendarEventRemoved) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateCalendarEventRemoved, x)
}

func (x *UpdateCalendarEventEdited) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateCalendarEventEdited, x)
}

func (x *UpdateRedirect) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdateRedirect, x)
}

func (x *ClientRedirect) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ClientRedirect, x)
}

func (x *UpdatePhoneCall) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_UpdatePhoneCall, x)
}

func (x *UpdateGetState) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateGetDifference) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateDifference) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateTooLong) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateState) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateMessageID) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateNewMessage) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateMessageEdited) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateMessagesDeleted) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateReadHistoryInbox) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateReadHistoryOutbox) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateMessagePinned) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateUserTyping) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateUserStatus) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateUsername) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateUserPhoto) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateNotifySettings) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateGroupParticipantAdd) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateGroupParticipantDeleted) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateGroupParticipantAdmin) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateGroupAdmins) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateGroupPhoto) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateReadMessagesContents) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateAuthorizationReset) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateDraftMessage) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateDraftMessageCleared) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateDialogPinned) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateDialogPinnedReorder) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateAccountPrivacy) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateLabelItemsAdded) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateLabelItemsRemoved) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateLabelSet) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateLabelDeleted) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateUserBlocked) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateMessagePoll) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateBotCallbackQuery) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateBotInlineQuery) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateBotInlineSend) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateTeamCreated) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateTeamMemberAdded) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateTeamMemberRemoved) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateTeamMemberStatus) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateTeamPhoto) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateTeam) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateCommunityMessage) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateCommunityReadOutbox) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateCommunityTyping) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateReaction) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateCalendarEventAdded) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateCalendarEventRemoved) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateCalendarEventEdited) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateRedirect) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ClientRedirect) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdatePhoneCall) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *UpdateGetState) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateGetDifference) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateDifference) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateTooLong) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateState) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateMessageID) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateNewMessage) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateMessageEdited) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateMessagesDeleted) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateReadHistoryInbox) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateReadHistoryOutbox) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateMessagePinned) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateUserTyping) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateUserStatus) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateUsername) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateUserPhoto) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateNotifySettings) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateGroupParticipantAdd) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateGroupParticipantDeleted) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateGroupParticipantAdmin) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateGroupAdmins) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateGroupPhoto) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateReadMessagesContents) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateAuthorizationReset) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateDraftMessage) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateDraftMessageCleared) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateDialogPinned) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateDialogPinnedReorder) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateAccountPrivacy) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateLabelItemsAdded) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateLabelItemsRemoved) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateLabelSet) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateLabelDeleted) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateUserBlocked) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateMessagePoll) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateBotCallbackQuery) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateBotInlineQuery) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateBotInlineSend) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateTeamCreated) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateTeamMemberAdded) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateTeamMemberRemoved) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateTeamMemberStatus) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateTeamPhoto) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateTeam) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateCommunityMessage) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateCommunityReadOutbox) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateCommunityTyping) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateReaction) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateCalendarEventAdded) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateCalendarEventRemoved) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateCalendarEventEdited) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdateRedirect) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *ClientRedirect) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *UpdatePhoneCall) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}
