package report

import (
	"fmt"
	"strings"

	"git.ronaksoftware.com/ronak/riversdk/log"
	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/pcap_parser"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// PcapRequest feed process type
type PcapRequest struct {
	ReqID        uint64
	AuthIDs      map[int64]bool
	RequestsList []string
	ResponseList []string
}

func (r *PcapRequest) String() string {
	return fmt.Sprintf("ReqID:%d \t AuthIDs[%v] \t Req[%v] \t Res[%v]", r.ReqID, r.AuthIDs, r.RequestsList, r.ResponseList)
}

// PcapReport pcap parser reporter
type PcapReport struct {
	// all captured requests and its responses
	Requests map[uint64]*PcapRequest

	// CorruptedRequest : requestID that have more than one item on its AuthIDs
	CorruptedRequests []*PcapRequest
	// DuplicatedRequest : requestID that have more than one item on its requestList
	DuplicatedRequest []*PcapRequest
	// DuplicatedResponse : requestID that have more than one item on its responseList
	DuplicatedResponse []*PcapRequest
	// TimedoutRequest : requestID that have zero item on its responseList
	TimedoutRequest []*PcapRequest

	ConstructorCounter map[int64]int64

	isProcessed bool
}

// NewPcapReport create new instance
func NewPcapReport() *PcapReport {
	r := new(PcapReport)
	r.Requests = make(map[uint64]*PcapRequest)
	r.CorruptedRequests = make([]*PcapRequest, 0)
	r.DuplicatedRequest = make([]*PcapRequest, 0)
	r.DuplicatedResponse = make([]*PcapRequest, 0)
	r.TimedoutRequest = make([]*PcapRequest, 0)
	r.ConstructorCounter = make(map[int64]int64)

	return r
}

// Feed insert data to reporter
func (r *PcapReport) Feed(p *pcap_parser.ParsedWS) error {

	act, ok := shared.GetCachedActorByAuthID(p.Message.AuthID)
	if !ok && p.Message.AuthID != 0 {
		return fmt.Errorf("Actor does not exist for this authID : %d", p.Message.AuthID)
	}
	envelop, err := decryptProto(act, p.Message)

	if err != nil {
		return err
	}
	// extract only messages reposnses and skip updates
	messages, _ := extractMessages(envelop)
	for _, m := range messages {

		r.ConstructorCounter[m.Constructor] = r.ConstructorCounter[m.Constructor] + 1
		// create report params
		req, ok := r.Requests[m.RequestID]
		if ok {
			// AuthID == 0 means unencrypted message
			if p.Message.AuthID != 0 {
				req.AuthIDs[p.Message.AuthID] = true
			}
			// response
			if p.SrcPort == shared.ServerPort {
				req.ResponseList = append(req.ResponseList, msg.ConstructorNames[m.Constructor])
			} else {
				// request
				req.RequestsList = append(req.RequestsList, msg.ConstructorNames[m.Constructor])
			}

		} else {
			req := &PcapRequest{
				ReqID:        m.RequestID,
				AuthIDs:      make(map[int64]bool),
				RequestsList: make([]string, 0),
				ResponseList: make([]string, 0),
			}

			req.AuthIDs[p.Message.AuthID] = true
			// response
			if p.SrcPort == shared.ServerPort {
				req.ResponseList = append(req.ResponseList, msg.ConstructorNames[m.Constructor])
			} else {
				// request
				req.RequestsList = append(req.RequestsList, msg.ConstructorNames[m.Constructor])
			}

			r.Requests[m.RequestID] = req
		}
	}
	r.isProcessed = false
	return nil
}

// ProcessFeeds process feed data and generate report items
func (r *PcapReport) ProcessFeeds() {

	for _, val := range r.Requests {
		// CorruptedRequest : requestID that have more than one item on its AuthIDs
		if len(val.AuthIDs) > 1 {
			r.CorruptedRequests = append(r.CorruptedRequests, val)
		}
		// DuplicatedRequest : requestID that have more than one item on its requestList
		if len(val.RequestsList) > 1 {
			r.DuplicatedRequest = append(r.DuplicatedRequest, val)
		}
		// DuplicatedResponse : requestID that have more than one item on its responseList
		if len(val.ResponseList) > 1 {
			r.DuplicatedResponse = append(r.DuplicatedResponse, val)
		}
		// TimedoutRequest : requestID that have zero item on its responseList
		if len(val.ResponseList) == 0 {
			r.TimedoutRequest = append(r.TimedoutRequest, val)
		}
	}
	r.isProcessed = true
}

func (r *PcapReport) String() string {
	if !r.isProcessed {
		r.ProcessFeeds()
	}
	sb := strings.Builder{}
	sb.WriteString("\n\nPacp Parser Reqult : \n\n")
	for _, v := range r.CorruptedRequests {
		sb.WriteString("\nCRP Req: " + v.String())
	}
	for _, v := range r.DuplicatedRequest {
		sb.WriteString("\nDUP Req : " + v.String())
	}

	for _, v := range r.DuplicatedResponse {
		sb.WriteString("\nDUP Res : " + v.String())
	}
	for _, v := range r.TimedoutRequest {
		sb.WriteString("\nTMO Req : " + v.String())
	}
	sb.WriteString("\n\nPacp Parser Summary : \n")
	sb.WriteString(fmt.Sprintf("\n\t Total Corrupted Requests : %d", len(r.CorruptedRequests)))
	sb.WriteString(fmt.Sprintf("\n\t Total Duplicate Requests : %d", len(r.DuplicatedRequest)))
	sb.WriteString(fmt.Sprintf("\n\t Total Duplicate Response : %d", len(r.DuplicatedResponse)))
	sb.WriteString(fmt.Sprintf("\n\t Total Timeouted Requests : %d", len(r.TimedoutRequest)))

	sb.WriteString("\n\t Received Messages : \n")
	for con, count := range r.ConstructorCounter {
		sb.WriteString(fmt.Sprintf("\n\t\t %s : %d", msg.ConstructorNames[con], count))
	}

	return sb.String()
}

func extractMessages(m *msg.MessageEnvelope) ([]*msg.MessageEnvelope, []*msg.UpdateContainer) {
	messages := make([]*msg.MessageEnvelope, 0)
	updates := make([]*msg.UpdateContainer, 0)
	switch m.Constructor {
	case msg.C_MessageContainer:
		x := new(msg.MessageContainer)
		err := x.Unmarshal(m.Message)
		if err == nil {
			for _, env := range x.Envelopes {
				msgs, upds := extractMessages(env)
				messages = append(messages, msgs...)
				updates = append(updates, upds...)
			}
		}
	case msg.C_UpdateContainer:
		x := new(msg.UpdateContainer)
		err := x.Unmarshal(m.Message)
		if err == nil {
			updates = append(updates, x)
		}
	case msg.C_Error:
		e := new(msg.Error)
		e.Unmarshal(m.Message)
		// its general error
		if m.RequestID == 0 {
			x := new(msg.Error)
			err := x.Unmarshal(m.Message)
			if err == nil {
				log.LOG_Error("PcapReport::extractMessages() Received General Error", zap.String("Code", x.Code), zap.String("Items", x.Items))
			} else {
				log.LOG_Error("PcapReport::extractMessages() Received General Error and failed to unmarshal it", zap.Error(err))
			}
		} else {
			// callback delegate will handle it
			messages = append(messages, m)
		}

	default:
		messages = append(messages, m)
	}
	return messages, updates
}

// FeedPacket insert data to reporter
func (r *PcapReport) FeedPacket(p *msg.ProtoMessage, isResponse bool) error {
	act, ok := shared.GetCachedActorByAuthID(p.AuthID)
	if !ok && p.AuthID != 0 {
		return fmt.Errorf("Actor does not exist for this authID : %d", p.AuthID)
	}

	envelop, err := decryptProto(act, p)

	if err != nil {
		return err
	}
	// extract only messages reposnses and skip updates
	messages, _ := extractMessages(envelop)
	for _, m := range messages {

		r.ConstructorCounter[m.Constructor] = r.ConstructorCounter[m.Constructor] + 1
		// create report params
		req, ok := r.Requests[m.RequestID]
		if ok {
			req.AuthIDs[p.AuthID] = true
			// response
			if isResponse {
				req.ResponseList = append(req.ResponseList, msg.ConstructorNames[m.Constructor])
			} else {
				// request
				req.RequestsList = append(req.RequestsList, msg.ConstructorNames[m.Constructor])
			}

		} else {
			req := &PcapRequest{
				ReqID:        m.RequestID,
				AuthIDs:      make(map[int64]bool),
				RequestsList: make([]string, 0),
				ResponseList: make([]string, 0),
			}

			req.AuthIDs[p.AuthID] = true
			// response
			if isResponse {
				req.ResponseList = append(req.ResponseList, msg.ConstructorNames[m.Constructor])
			} else {
				// request
				req.RequestsList = append(req.RequestsList, msg.ConstructorNames[m.Constructor])
			}

			r.Requests[m.RequestID] = req
		}
	}
	r.isProcessed = false
	return nil
}

func decryptProto(act shared.Acter, protMsg *msg.ProtoMessage) (*msg.MessageEnvelope, error) {
	if protMsg.AuthID == 0 {
		env := new(msg.MessageEnvelope)
		err := env.Unmarshal(protMsg.Payload)
		if err != nil {
			return nil, fmt.Errorf("decryptProto() AuthID=0 Unmarshal protMsg.Payload , err : %s", err.Error())
		}
		return env, nil
	}
	if act == nil {
		return nil, fmt.Errorf("decryptProto() when protMsg have authID & encrypted, actor can't be null")
	}

	authID, authKey := act.GetAuthInfo()
	if authID != protMsg.AuthID {
		return nil, fmt.Errorf("decryptProto() Actor AuthID:%d is not equal to ProtoMsg AuthID :%d", authID, protMsg.AuthID)
	}
	decryptedBytes, err := domain.Decrypt(authKey, protMsg.MessageKey, protMsg.Payload)
	if err != nil {
		return nil, fmt.Errorf("decryptProto() -> domain.Decrypt() , err : %s", err.Error())
	}
	encryptedPayload := new(msg.ProtoEncryptedPayload)
	err = encryptedPayload.Unmarshal(decryptedBytes)
	if err != nil {
		return nil, fmt.Errorf("decryptProto() -> Unmarshal(decryptedBytes) , err : %s", err.Error())
	}

	return encryptedPayload.Envelope, nil
}
