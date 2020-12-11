package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_MessagesSend int64 = 2719516490

type poolMessagesSend struct {
	pool sync.Pool
}

func (p *poolMessagesSend) Get() *MessagesSend {
	x, ok := p.pool.Get().(*MessagesSend)
	if !ok {
		return &MessagesSend{}
	}
	return x
}

func (p *poolMessagesSend) Put(x *MessagesSend) {
	x.RandomID = 0
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.Body = ""
	x.ReplyTo = 0
	x.ClearDraft = false
	x.Entities = x.Entities[:0]
	p.pool.Put(x)
}

var PoolMessagesSend = poolMessagesSend{}

const C_MessagesSendMedia int64 = 1193569473

type poolMessagesSendMedia struct {
	pool sync.Pool
}

func (p *poolMessagesSendMedia) Get() *MessagesSendMedia {
	x, ok := p.pool.Get().(*MessagesSendMedia)
	if !ok {
		return &MessagesSendMedia{}
	}
	return x
}

func (p *poolMessagesSendMedia) Put(x *MessagesSendMedia) {
	x.RandomID = 0
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MediaType = 0
	x.MediaData = x.MediaData[:0]
	x.ReplyTo = 0
	x.ClearDraft = false
	p.pool.Put(x)
}

var PoolMessagesSendMedia = poolMessagesSendMedia{}

const C_MessagesEdit int64 = 2220778397

type poolMessagesEdit struct {
	pool sync.Pool
}

func (p *poolMessagesEdit) Get() *MessagesEdit {
	x, ok := p.pool.Get().(*MessagesEdit)
	if !ok {
		return &MessagesEdit{}
	}
	return x
}

func (p *poolMessagesEdit) Put(x *MessagesEdit) {
	x.RandomID = 0
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.Body = ""
	x.MessageID = 0
	x.Entities = x.Entities[:0]
	p.pool.Put(x)
}

var PoolMessagesEdit = poolMessagesEdit{}

const C_MessagesReadHistory int64 = 2718718556

type poolMessagesReadHistory struct {
	pool sync.Pool
}

func (p *poolMessagesReadHistory) Get() *MessagesReadHistory {
	x, ok := p.pool.Get().(*MessagesReadHistory)
	if !ok {
		return &MessagesReadHistory{}
	}
	return x
}

func (p *poolMessagesReadHistory) Put(x *MessagesReadHistory) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MaxID = 0
	p.pool.Put(x)
}

var PoolMessagesReadHistory = poolMessagesReadHistory{}

const C_MessagesGet int64 = 2530327223

type poolMessagesGet struct {
	pool sync.Pool
}

func (p *poolMessagesGet) Get() *MessagesGet {
	x, ok := p.pool.Get().(*MessagesGet)
	if !ok {
		return &MessagesGet{}
	}
	return x
}

func (p *poolMessagesGet) Put(x *MessagesGet) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MessagesIDs = x.MessagesIDs[:0]
	p.pool.Put(x)
}

var PoolMessagesGet = poolMessagesGet{}

const C_MessagesGetHistory int64 = 2587549819

type poolMessagesGetHistory struct {
	pool sync.Pool
}

func (p *poolMessagesGetHistory) Get() *MessagesGetHistory {
	x, ok := p.pool.Get().(*MessagesGetHistory)
	if !ok {
		return &MessagesGetHistory{}
	}
	return x
}

func (p *poolMessagesGetHistory) Put(x *MessagesGetHistory) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.Limit = 0
	x.MaxID = 0
	x.MinID = 0
	p.pool.Put(x)
}

var PoolMessagesGetHistory = poolMessagesGetHistory{}

const C_MessagesGetDialogs int64 = 91700887

type poolMessagesGetDialogs struct {
	pool sync.Pool
}

func (p *poolMessagesGetDialogs) Get() *MessagesGetDialogs {
	x, ok := p.pool.Get().(*MessagesGetDialogs)
	if !ok {
		return &MessagesGetDialogs{}
	}
	return x
}

func (p *poolMessagesGetDialogs) Put(x *MessagesGetDialogs) {
	x.Limit = 0
	x.Offset = 0
	x.ExcludePinned = false
	p.pool.Put(x)
}

var PoolMessagesGetDialogs = poolMessagesGetDialogs{}

const C_MessagesGetPinnedDialogs int64 = 2418490440

type poolMessagesGetPinnedDialogs struct {
	pool sync.Pool
}

func (p *poolMessagesGetPinnedDialogs) Get() *MessagesGetPinnedDialogs {
	x, ok := p.pool.Get().(*MessagesGetPinnedDialogs)
	if !ok {
		return &MessagesGetPinnedDialogs{}
	}
	return x
}

func (p *poolMessagesGetPinnedDialogs) Put(x *MessagesGetPinnedDialogs) {
	p.pool.Put(x)
}

var PoolMessagesGetPinnedDialogs = poolMessagesGetPinnedDialogs{}

const C_MessagesGetDialog int64 = 2013525138

type poolMessagesGetDialog struct {
	pool sync.Pool
}

func (p *poolMessagesGetDialog) Get() *MessagesGetDialog {
	x, ok := p.pool.Get().(*MessagesGetDialog)
	if !ok {
		return &MessagesGetDialog{}
	}
	return x
}

func (p *poolMessagesGetDialog) Put(x *MessagesGetDialog) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	p.pool.Put(x)
}

var PoolMessagesGetDialog = poolMessagesGetDialog{}

const C_MessagesSetTyping int64 = 493662630

type poolMessagesSetTyping struct {
	pool sync.Pool
}

func (p *poolMessagesSetTyping) Get() *MessagesSetTyping {
	x, ok := p.pool.Get().(*MessagesSetTyping)
	if !ok {
		return &MessagesSetTyping{}
	}
	return x
}

func (p *poolMessagesSetTyping) Put(x *MessagesSetTyping) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.Action = 0
	p.pool.Put(x)
}

var PoolMessagesSetTyping = poolMessagesSetTyping{}

const C_MessagesClearHistory int64 = 730920117

type poolMessagesClearHistory struct {
	pool sync.Pool
}

func (p *poolMessagesClearHistory) Get() *MessagesClearHistory {
	x, ok := p.pool.Get().(*MessagesClearHistory)
	if !ok {
		return &MessagesClearHistory{}
	}
	return x
}

func (p *poolMessagesClearHistory) Put(x *MessagesClearHistory) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MaxID = 0
	x.Delete = false
	p.pool.Put(x)
}

var PoolMessagesClearHistory = poolMessagesClearHistory{}

const C_MessagesDelete int64 = 4211128401

type poolMessagesDelete struct {
	pool sync.Pool
}

func (p *poolMessagesDelete) Get() *MessagesDelete {
	x, ok := p.pool.Get().(*MessagesDelete)
	if !ok {
		return &MessagesDelete{}
	}
	return x
}

func (p *poolMessagesDelete) Put(x *MessagesDelete) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MessageIDs = x.MessageIDs[:0]
	x.Revoke = false
	p.pool.Put(x)
}

var PoolMessagesDelete = poolMessagesDelete{}

const C_MessagesForward int64 = 2296400547

type poolMessagesForward struct {
	pool sync.Pool
}

func (p *poolMessagesForward) Get() *MessagesForward {
	x, ok := p.pool.Get().(*MessagesForward)
	if !ok {
		return &MessagesForward{}
	}
	return x
}

func (p *poolMessagesForward) Put(x *MessagesForward) {
	if x.FromPeer != nil {
		PoolInputPeer.Put(x.FromPeer)
		x.FromPeer = nil
	}
	if x.ToPeer != nil {
		PoolInputPeer.Put(x.ToPeer)
		x.ToPeer = nil
	}
	x.Silence = false
	x.MessageIDs = x.MessageIDs[:0]
	x.RandomID = 0
	p.pool.Put(x)
}

var PoolMessagesForward = poolMessagesForward{}

const C_MessagesReadContents int64 = 934027930

type poolMessagesReadContents struct {
	pool sync.Pool
}

func (p *poolMessagesReadContents) Get() *MessagesReadContents {
	x, ok := p.pool.Get().(*MessagesReadContents)
	if !ok {
		return &MessagesReadContents{}
	}
	return x
}

func (p *poolMessagesReadContents) Put(x *MessagesReadContents) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MessageIDs = x.MessageIDs[:0]
	p.pool.Put(x)
}

var PoolMessagesReadContents = poolMessagesReadContents{}

const C_MessagesSaveDraft int64 = 1884509359

type poolMessagesSaveDraft struct {
	pool sync.Pool
}

func (p *poolMessagesSaveDraft) Get() *MessagesSaveDraft {
	x, ok := p.pool.Get().(*MessagesSaveDraft)
	if !ok {
		return &MessagesSaveDraft{}
	}
	return x
}

func (p *poolMessagesSaveDraft) Put(x *MessagesSaveDraft) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.ReplyTo = 0
	x.Body = ""
	x.Entities = x.Entities[:0]
	x.EditedID = 0
	p.pool.Put(x)
}

var PoolMessagesSaveDraft = poolMessagesSaveDraft{}

const C_MessagesClearDraft int64 = 3502044240

type poolMessagesClearDraft struct {
	pool sync.Pool
}

func (p *poolMessagesClearDraft) Get() *MessagesClearDraft {
	x, ok := p.pool.Get().(*MessagesClearDraft)
	if !ok {
		return &MessagesClearDraft{}
	}
	return x
}

func (p *poolMessagesClearDraft) Put(x *MessagesClearDraft) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	p.pool.Put(x)
}

var PoolMessagesClearDraft = poolMessagesClearDraft{}

const C_MessagesToggleDialogPin int64 = 2713966841

type poolMessagesToggleDialogPin struct {
	pool sync.Pool
}

func (p *poolMessagesToggleDialogPin) Get() *MessagesToggleDialogPin {
	x, ok := p.pool.Get().(*MessagesToggleDialogPin)
	if !ok {
		return &MessagesToggleDialogPin{}
	}
	return x
}

func (p *poolMessagesToggleDialogPin) Put(x *MessagesToggleDialogPin) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.Pin = false
	p.pool.Put(x)
}

var PoolMessagesToggleDialogPin = poolMessagesToggleDialogPin{}

const C_MessagesReorderPinnedDialogs int64 = 2974098809

type poolMessagesReorderPinnedDialogs struct {
	pool sync.Pool
}

func (p *poolMessagesReorderPinnedDialogs) Get() *MessagesReorderPinnedDialogs {
	x, ok := p.pool.Get().(*MessagesReorderPinnedDialogs)
	if !ok {
		return &MessagesReorderPinnedDialogs{}
	}
	return x
}

func (p *poolMessagesReorderPinnedDialogs) Put(x *MessagesReorderPinnedDialogs) {
	x.Peers = x.Peers[:0]
	p.pool.Put(x)
}

var PoolMessagesReorderPinnedDialogs = poolMessagesReorderPinnedDialogs{}

const C_MessagesSendScreenShotNotification int64 = 679612941

type poolMessagesSendScreenShotNotification struct {
	pool sync.Pool
}

func (p *poolMessagesSendScreenShotNotification) Get() *MessagesSendScreenShotNotification {
	x, ok := p.pool.Get().(*MessagesSendScreenShotNotification)
	if !ok {
		return &MessagesSendScreenShotNotification{}
	}
	return x
}

func (p *poolMessagesSendScreenShotNotification) Put(x *MessagesSendScreenShotNotification) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.RandomID = 0
	x.ReplyTo = 0
	x.MinID = 0
	x.MaxID = 0
	p.pool.Put(x)
}

var PoolMessagesSendScreenShotNotification = poolMessagesSendScreenShotNotification{}

const C_MessagesSendReaction int64 = 1294935032

type poolMessagesSendReaction struct {
	pool sync.Pool
}

func (p *poolMessagesSendReaction) Get() *MessagesSendReaction {
	x, ok := p.pool.Get().(*MessagesSendReaction)
	if !ok {
		return &MessagesSendReaction{}
	}
	return x
}

func (p *poolMessagesSendReaction) Put(x *MessagesSendReaction) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MessageID = 0
	x.Reaction = ""
	p.pool.Put(x)
}

var PoolMessagesSendReaction = poolMessagesSendReaction{}

const C_MessagesDeleteReaction int64 = 3897816690

type poolMessagesDeleteReaction struct {
	pool sync.Pool
}

func (p *poolMessagesDeleteReaction) Get() *MessagesDeleteReaction {
	x, ok := p.pool.Get().(*MessagesDeleteReaction)
	if !ok {
		return &MessagesDeleteReaction{}
	}
	return x
}

func (p *poolMessagesDeleteReaction) Put(x *MessagesDeleteReaction) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MessageID = 0
	x.Reactions = x.Reactions[:0]
	p.pool.Put(x)
}

var PoolMessagesDeleteReaction = poolMessagesDeleteReaction{}

const C_MessagesGetReactionList int64 = 1241106883

type poolMessagesGetReactionList struct {
	pool sync.Pool
}

func (p *poolMessagesGetReactionList) Get() *MessagesGetReactionList {
	x, ok := p.pool.Get().(*MessagesGetReactionList)
	if !ok {
		return &MessagesGetReactionList{}
	}
	return x
}

func (p *poolMessagesGetReactionList) Put(x *MessagesGetReactionList) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MessageID = 0
	x.Hash = 0
	p.pool.Put(x)
}

var PoolMessagesGetReactionList = poolMessagesGetReactionList{}

const C_MessagesTogglePin int64 = 4009065684

type poolMessagesTogglePin struct {
	pool sync.Pool
}

func (p *poolMessagesTogglePin) Get() *MessagesTogglePin {
	x, ok := p.pool.Get().(*MessagesTogglePin)
	if !ok {
		return &MessagesTogglePin{}
	}
	return x
}

func (p *poolMessagesTogglePin) Put(x *MessagesTogglePin) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.MessageID = 0
	x.Silent = false
	p.pool.Put(x)
}

var PoolMessagesTogglePin = poolMessagesTogglePin{}

const C_MessagesDialogs int64 = 3615935362

type poolMessagesDialogs struct {
	pool sync.Pool
}

func (p *poolMessagesDialogs) Get() *MessagesDialogs {
	x, ok := p.pool.Get().(*MessagesDialogs)
	if !ok {
		return &MessagesDialogs{}
	}
	return x
}

func (p *poolMessagesDialogs) Put(x *MessagesDialogs) {
	x.Dialogs = x.Dialogs[:0]
	x.Users = x.Users[:0]
	x.Messages = x.Messages[:0]
	x.Count = 0
	x.UpdateID = 0
	x.Groups = x.Groups[:0]
	p.pool.Put(x)
}

var PoolMessagesDialogs = poolMessagesDialogs{}

const C_MessagesSent int64 = 3215955758

type poolMessagesSent struct {
	pool sync.Pool
}

func (p *poolMessagesSent) Get() *MessagesSent {
	x, ok := p.pool.Get().(*MessagesSent)
	if !ok {
		return &MessagesSent{}
	}
	return x
}

func (p *poolMessagesSent) Put(x *MessagesSent) {
	x.MessageID = 0
	x.RandomID = 0
	x.CreatedOn = 0
	p.pool.Put(x)
}

var PoolMessagesSent = poolMessagesSent{}

const C_MessagesMany int64 = 1993434083

type poolMessagesMany struct {
	pool sync.Pool
}

func (p *poolMessagesMany) Get() *MessagesMany {
	x, ok := p.pool.Get().(*MessagesMany)
	if !ok {
		return &MessagesMany{}
	}
	return x
}

func (p *poolMessagesMany) Put(x *MessagesMany) {
	x.Messages = x.Messages[:0]
	x.Users = x.Users[:0]
	x.Groups = x.Groups[:0]
	x.Continuous = false
	x.Empty = false
	p.pool.Put(x)
}

var PoolMessagesMany = poolMessagesMany{}

const C_MessagesReactionList int64 = 1464437214

type poolMessagesReactionList struct {
	pool sync.Pool
}

func (p *poolMessagesReactionList) Get() *MessagesReactionList {
	x, ok := p.pool.Get().(*MessagesReactionList)
	if !ok {
		return &MessagesReactionList{}
	}
	return x
}

func (p *poolMessagesReactionList) Put(x *MessagesReactionList) {
	x.List = x.List[:0]
	x.Users = x.Users[:0]
	x.Hash = 0
	x.Modified = false
	p.pool.Put(x)
}

var PoolMessagesReactionList = poolMessagesReactionList{}

const C_ReactionList int64 = 3980071153

type poolReactionList struct {
	pool sync.Pool
}

func (p *poolReactionList) Get() *ReactionList {
	x, ok := p.pool.Get().(*ReactionList)
	if !ok {
		return &ReactionList{}
	}
	return x
}

func (p *poolReactionList) Put(x *ReactionList) {
	x.Reaction = ""
	x.UserIDs = x.UserIDs[:0]
	p.pool.Put(x)
}

var PoolReactionList = poolReactionList{}

func init() {
	registry.RegisterConstructor(2719516490, "msg.MessagesSend")
	registry.RegisterConstructor(1193569473, "msg.MessagesSendMedia")
	registry.RegisterConstructor(2220778397, "msg.MessagesEdit")
	registry.RegisterConstructor(2718718556, "msg.MessagesReadHistory")
	registry.RegisterConstructor(2530327223, "msg.MessagesGet")
	registry.RegisterConstructor(2587549819, "msg.MessagesGetHistory")
	registry.RegisterConstructor(91700887, "msg.MessagesGetDialogs")
	registry.RegisterConstructor(2418490440, "msg.MessagesGetPinnedDialogs")
	registry.RegisterConstructor(2013525138, "msg.MessagesGetDialog")
	registry.RegisterConstructor(493662630, "msg.MessagesSetTyping")
	registry.RegisterConstructor(730920117, "msg.MessagesClearHistory")
	registry.RegisterConstructor(4211128401, "msg.MessagesDelete")
	registry.RegisterConstructor(2296400547, "msg.MessagesForward")
	registry.RegisterConstructor(934027930, "msg.MessagesReadContents")
	registry.RegisterConstructor(1884509359, "msg.MessagesSaveDraft")
	registry.RegisterConstructor(3502044240, "msg.MessagesClearDraft")
	registry.RegisterConstructor(2713966841, "msg.MessagesToggleDialogPin")
	registry.RegisterConstructor(2974098809, "msg.MessagesReorderPinnedDialogs")
	registry.RegisterConstructor(679612941, "msg.MessagesSendScreenShotNotification")
	registry.RegisterConstructor(1294935032, "msg.MessagesSendReaction")
	registry.RegisterConstructor(3897816690, "msg.MessagesDeleteReaction")
	registry.RegisterConstructor(1241106883, "msg.MessagesGetReactionList")
	registry.RegisterConstructor(4009065684, "msg.MessagesTogglePin")
	registry.RegisterConstructor(3615935362, "msg.MessagesDialogs")
	registry.RegisterConstructor(3215955758, "msg.MessagesSent")
	registry.RegisterConstructor(1993434083, "msg.MessagesMany")
	registry.RegisterConstructor(1464437214, "msg.MessagesReactionList")
	registry.RegisterConstructor(3980071153, "msg.ReactionList")
}

func (x *MessagesSend) DeepCopy(z *MessagesSend) {
	z.RandomID = x.RandomID
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.Body = x.Body
	z.ReplyTo = x.ReplyTo
	z.ClearDraft = x.ClearDraft
	for idx := range x.Entities {
		if x.Entities[idx] != nil {
			xx := PoolMessageEntity.Get()
			x.Entities[idx].DeepCopy(xx)
			z.Entities = append(z.Entities, xx)
		}
	}
}

func (x *MessagesSendMedia) DeepCopy(z *MessagesSendMedia) {
	z.RandomID = x.RandomID
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MediaType = x.MediaType
	z.MediaData = append(z.MediaData[:0], x.MediaData...)
	z.ReplyTo = x.ReplyTo
	z.ClearDraft = x.ClearDraft
}

func (x *MessagesEdit) DeepCopy(z *MessagesEdit) {
	z.RandomID = x.RandomID
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.Body = x.Body
	z.MessageID = x.MessageID
	for idx := range x.Entities {
		if x.Entities[idx] != nil {
			xx := PoolMessageEntity.Get()
			x.Entities[idx].DeepCopy(xx)
			z.Entities = append(z.Entities, xx)
		}
	}
}

func (x *MessagesReadHistory) DeepCopy(z *MessagesReadHistory) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MaxID = x.MaxID
}

func (x *MessagesGet) DeepCopy(z *MessagesGet) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MessagesIDs = append(z.MessagesIDs[:0], x.MessagesIDs...)
}

func (x *MessagesGetHistory) DeepCopy(z *MessagesGetHistory) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.Limit = x.Limit
	z.MaxID = x.MaxID
	z.MinID = x.MinID
}

func (x *MessagesGetDialogs) DeepCopy(z *MessagesGetDialogs) {
	z.Limit = x.Limit
	z.Offset = x.Offset
	z.ExcludePinned = x.ExcludePinned
}

func (x *MessagesGetPinnedDialogs) DeepCopy(z *MessagesGetPinnedDialogs) {
}

func (x *MessagesGetDialog) DeepCopy(z *MessagesGetDialog) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
}

func (x *MessagesSetTyping) DeepCopy(z *MessagesSetTyping) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.Action = x.Action
}

func (x *MessagesClearHistory) DeepCopy(z *MessagesClearHistory) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MaxID = x.MaxID
	z.Delete = x.Delete
}

func (x *MessagesDelete) DeepCopy(z *MessagesDelete) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MessageIDs = append(z.MessageIDs[:0], x.MessageIDs...)
	z.Revoke = x.Revoke
}

func (x *MessagesForward) DeepCopy(z *MessagesForward) {
	if x.FromPeer != nil {
		z.FromPeer = PoolInputPeer.Get()
		x.FromPeer.DeepCopy(z.FromPeer)
	}
	if x.ToPeer != nil {
		z.ToPeer = PoolInputPeer.Get()
		x.ToPeer.DeepCopy(z.ToPeer)
	}
	z.Silence = x.Silence
	z.MessageIDs = append(z.MessageIDs[:0], x.MessageIDs...)
	z.RandomID = x.RandomID
}

func (x *MessagesReadContents) DeepCopy(z *MessagesReadContents) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MessageIDs = append(z.MessageIDs[:0], x.MessageIDs...)
}

func (x *MessagesSaveDraft) DeepCopy(z *MessagesSaveDraft) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.ReplyTo = x.ReplyTo
	z.Body = x.Body
	for idx := range x.Entities {
		if x.Entities[idx] != nil {
			xx := PoolMessageEntity.Get()
			x.Entities[idx].DeepCopy(xx)
			z.Entities = append(z.Entities, xx)
		}
	}
	z.EditedID = x.EditedID
}

func (x *MessagesClearDraft) DeepCopy(z *MessagesClearDraft) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
}

func (x *MessagesToggleDialogPin) DeepCopy(z *MessagesToggleDialogPin) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.Pin = x.Pin
}

func (x *MessagesReorderPinnedDialogs) DeepCopy(z *MessagesReorderPinnedDialogs) {
	for idx := range x.Peers {
		if x.Peers[idx] != nil {
			xx := PoolInputPeer.Get()
			x.Peers[idx].DeepCopy(xx)
			z.Peers = append(z.Peers, xx)
		}
	}
}

func (x *MessagesSendScreenShotNotification) DeepCopy(z *MessagesSendScreenShotNotification) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.RandomID = x.RandomID
	z.ReplyTo = x.ReplyTo
	z.MinID = x.MinID
	z.MaxID = x.MaxID
}

func (x *MessagesSendReaction) DeepCopy(z *MessagesSendReaction) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MessageID = x.MessageID
	z.Reaction = x.Reaction
}

func (x *MessagesDeleteReaction) DeepCopy(z *MessagesDeleteReaction) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MessageID = x.MessageID
	z.Reactions = append(z.Reactions[:0], x.Reactions...)
}

func (x *MessagesGetReactionList) DeepCopy(z *MessagesGetReactionList) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MessageID = x.MessageID
	z.Hash = x.Hash
}

func (x *MessagesTogglePin) DeepCopy(z *MessagesTogglePin) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.MessageID = x.MessageID
	z.Silent = x.Silent
}

func (x *MessagesDialogs) DeepCopy(z *MessagesDialogs) {
	for idx := range x.Dialogs {
		if x.Dialogs[idx] != nil {
			xx := PoolDialog.Get()
			x.Dialogs[idx].DeepCopy(xx)
			z.Dialogs = append(z.Dialogs, xx)
		}
	}
	for idx := range x.Users {
		if x.Users[idx] != nil {
			xx := PoolUser.Get()
			x.Users[idx].DeepCopy(xx)
			z.Users = append(z.Users, xx)
		}
	}
	for idx := range x.Messages {
		if x.Messages[idx] != nil {
			xx := PoolUserMessage.Get()
			x.Messages[idx].DeepCopy(xx)
			z.Messages = append(z.Messages, xx)
		}
	}
	z.Count = x.Count
	z.UpdateID = x.UpdateID
	for idx := range x.Groups {
		if x.Groups[idx] != nil {
			xx := PoolGroup.Get()
			x.Groups[idx].DeepCopy(xx)
			z.Groups = append(z.Groups, xx)
		}
	}
}

func (x *MessagesSent) DeepCopy(z *MessagesSent) {
	z.MessageID = x.MessageID
	z.RandomID = x.RandomID
	z.CreatedOn = x.CreatedOn
}

func (x *MessagesMany) DeepCopy(z *MessagesMany) {
	for idx := range x.Messages {
		if x.Messages[idx] != nil {
			xx := PoolUserMessage.Get()
			x.Messages[idx].DeepCopy(xx)
			z.Messages = append(z.Messages, xx)
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
	z.Continuous = x.Continuous
	z.Empty = x.Empty
}

func (x *MessagesReactionList) DeepCopy(z *MessagesReactionList) {
	for idx := range x.List {
		if x.List[idx] != nil {
			xx := PoolReactionList.Get()
			x.List[idx].DeepCopy(xx)
			z.List = append(z.List, xx)
		}
	}
	for idx := range x.Users {
		if x.Users[idx] != nil {
			xx := PoolUser.Get()
			x.Users[idx].DeepCopy(xx)
			z.Users = append(z.Users, xx)
		}
	}
	z.Hash = x.Hash
	z.Modified = x.Modified
}

func (x *ReactionList) DeepCopy(z *ReactionList) {
	z.Reaction = x.Reaction
	z.UserIDs = append(z.UserIDs[:0], x.UserIDs...)
}

func (x *MessagesSend) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesSend, x)
}

func (x *MessagesSendMedia) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesSendMedia, x)
}

func (x *MessagesEdit) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesEdit, x)
}

func (x *MessagesReadHistory) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesReadHistory, x)
}

func (x *MessagesGet) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesGet, x)
}

func (x *MessagesGetHistory) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesGetHistory, x)
}

func (x *MessagesGetDialogs) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesGetDialogs, x)
}

func (x *MessagesGetPinnedDialogs) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesGetPinnedDialogs, x)
}

func (x *MessagesGetDialog) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesGetDialog, x)
}

func (x *MessagesSetTyping) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesSetTyping, x)
}

func (x *MessagesClearHistory) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesClearHistory, x)
}

func (x *MessagesDelete) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesDelete, x)
}

func (x *MessagesForward) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesForward, x)
}

func (x *MessagesReadContents) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesReadContents, x)
}

func (x *MessagesSaveDraft) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesSaveDraft, x)
}

func (x *MessagesClearDraft) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesClearDraft, x)
}

func (x *MessagesToggleDialogPin) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesToggleDialogPin, x)
}

func (x *MessagesReorderPinnedDialogs) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesReorderPinnedDialogs, x)
}

func (x *MessagesSendScreenShotNotification) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesSendScreenShotNotification, x)
}

func (x *MessagesSendReaction) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesSendReaction, x)
}

func (x *MessagesDeleteReaction) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesDeleteReaction, x)
}

func (x *MessagesGetReactionList) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesGetReactionList, x)
}

func (x *MessagesTogglePin) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesTogglePin, x)
}

func (x *MessagesDialogs) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesDialogs, x)
}

func (x *MessagesSent) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesSent, x)
}

func (x *MessagesMany) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesMany, x)
}

func (x *MessagesReactionList) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessagesReactionList, x)
}

func (x *ReactionList) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ReactionList, x)
}

func (x *MessagesSend) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesSendMedia) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesEdit) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesReadHistory) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesGet) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesGetHistory) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesGetDialogs) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesGetPinnedDialogs) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesGetDialog) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesSetTyping) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesClearHistory) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesDelete) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesForward) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesReadContents) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesSaveDraft) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesClearDraft) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesToggleDialogPin) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesReorderPinnedDialogs) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesSendScreenShotNotification) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesSendReaction) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesDeleteReaction) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesGetReactionList) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesTogglePin) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesDialogs) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesSent) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesMany) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesReactionList) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ReactionList) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessagesSend) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesSendMedia) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesEdit) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesReadHistory) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesGet) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesGetHistory) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesGetDialogs) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesGetPinnedDialogs) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesGetDialog) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesSetTyping) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesClearHistory) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesDelete) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesForward) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesReadContents) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesSaveDraft) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesClearDraft) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesToggleDialogPin) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesReorderPinnedDialogs) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesSendScreenShotNotification) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesSendReaction) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesDeleteReaction) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesGetReactionList) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesTogglePin) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesDialogs) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesSent) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesMany) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessagesReactionList) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *ReactionList) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}
