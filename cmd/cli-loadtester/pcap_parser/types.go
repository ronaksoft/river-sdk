package pcap_parser

import (
	"encoding/binary"
	"net"
	"unsafe"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
)

// ParsedWS parsed package with required info
type ParsedWS struct {
	RowNo   int // increases for any packet in pcap file
	Counter int // increases only for vaild ws packet
	SrcIP   net.IP
	DstIP   net.IP
	SrcPort uint16
	DstPort uint16
	Message *msg.ProtoMessage
}

// WS websocket frame header is 8 bytes
type WS struct {
	Header   []byte // 8 bytes
	Payload  []byte
	IsMasked bool
}

// NewWS unmarshal websocket from payload buff
func NewWS(tcpPayload []byte) *WS {

	n := len(tcpPayload)
	if n > 8 {
		ws := new(WS)
		ws.IsMasked = (tcpPayload[1] >> 7) > 0
		if ws.IsMasked {
			ws.Header = make([]byte, 8)
			ws.Payload = make([]byte, n-8)
			copy(ws.Header, tcpPayload[0:8])
			copy(ws.Payload, tcpPayload[8:])
		} else {
			ws.Header = make([]byte, 4)
			ws.Payload = make([]byte, n-4)
			copy(ws.Header, tcpPayload[0:4])
			copy(ws.Payload, tcpPayload[4:])
		}
		return ws
	}
	return nil
}

// Fin 1 bit Header[0] X000-0000
func (w *WS) Fin() bool {
	return (w.Header[0] >> 7) > 0
}

// Reserved 3 bit Header[0] 0XXX-0000
func (w *WS) Reserved() (resv1, resv2, resv3 bool) {
	resv1 = ((w.Header[0] >> 4) & 0x4) > 0
	resv2 = ((w.Header[0] >> 4) & 0x2) > 0
	resv3 = ((w.Header[0] >> 4) & 0x1) > 0
	return
}

// OpCode 4 bit Header[0] 0000-XXXX
func (w *WS) OpCode() int {
	return int(w.Header[0] & 0xF)
}

// Mask 1 bit Header[1] X000-0000
func (w *WS) Mask() bool {
	//return (w.Header[1] >> 7) > 0
	return w.IsMasked
}

// PayloadLength 7 bit Header[1] 0XXX-XXXX
func (w *WS) PayloadLength() int {
	return int(w.Header[1] & 0x7F)
}

// ExtendedPayloadLength 16 bit Header[2:4] XXXX-XXXX XXXX-XXXX
func (w *WS) ExtendedPayloadLength() int {
	len := binary.BigEndian.Uint16(w.Header[2:4])
	return int(len)
}

// MaskingKey 32 bit Header[4:8] XXXX-XXXX XXXX-XXXX XXXX-XXXX XXXX-XXXX
func (w *WS) MaskingKey() [4]byte {
	res := [4]byte{0, 0, 0, 0}
	if w.IsMasked {
		copy(res[:], w.Header[4:8])
	}
	return res
}

// UnmaskPayload (Payload XOR MaskingKey)
func (w *WS) UnmaskPayload() []byte {
	if w.IsMasked {
		key := w.MaskingKey()
		cipher(w.Payload, key, 0)
		// set masked flag to false
		w.Header[1] = w.Header[1] & 0x7F
	}
	return w.Payload
}

// IsValid simply compare len of payload and length in header :/
func (w *WS) IsValid() (bool, *msg.ProtoMessage) {
	x := new(msg.ProtoMessage)
	err := x.Unmarshal(w.UnmaskPayload())
	if err != nil {
		return false, nil
	}
	return true, x
}

// Cipher copied from github.com/gobwas/ws/cipher.go xD
func cipher(payload []byte, mask [4]byte, offset int) {
	n := len(payload)
	if n < 8 {
		for i := 0; i < n; i++ {
			payload[i] ^= mask[(offset+i)%4]
		}
		return
	}

	// Calculate position in mask due to previously processed bytes number.
	mpos := offset % 4
	// Count number of bytes will processed one by one from the beginning of payload.
	ln := remain[mpos]
	// Count number of bytes will processed one by one from the end of payload.
	// This is done to process payload by 8 bytes in each iteration of main loop.
	rn := (n - ln) % 8

	for i := 0; i < ln; i++ {
		payload[i] ^= mask[(mpos+i)%4]
	}
	for i := n - rn; i < n; i++ {
		payload[i] ^= mask[(mpos+i)%4]
	}

	// We should cast mask to uint32 with unsafe instead of encoding.BigEndian
	// to avoid care of os dependent byte order. That is, on any endianess mask
	// and payload will be presented with the same order. In other words, we
	// could not use encoding.BigEndian on xoring payload as uint64.
	m := *(*uint32)(unsafe.Pointer(&mask))
	m2 := uint64(m)<<32 | uint64(m)

	// Get pointer to payload at ln index to
	// skip manual processed bytes above.
	p := uintptr(unsafe.Pointer(&payload[ln]))
	// Also skip right part as the division by 8 remainder.
	// Divide it by 8 to get number of uint64 parts remaining to process.
	n = (n - rn) >> 3
	// Process the rest of bytes as uint64.
	for i := 0; i < n; i, p = i+1, p+8 {
		v := (*uint64)(unsafe.Pointer(p))
		*v = *v ^ m2
	}
}

// copied from github.com/gobwas/ws/cipher.go xD
// remain maps position in masking key [0,4) to number
// of bytes that need to be processed manually inside Cipher().
var remain = [4]int{0, 3, 2, 1}
