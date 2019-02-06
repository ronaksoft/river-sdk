package pcap_parser

import (
	"encoding/binary"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
)

var (
	reassmblerTCP map[uint32]gopacket.Packet
)

func Parse(pcapFile string) (chan *ParsedWS, error) {

	result := make(chan *ParsedWS)

	handle, err := pcap.OpenOffline(pcapFile)
	if err != nil {
		return nil, err
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	go parsePackects_withAssembly(result, packetSource)

	return result, nil

}

func parsePackects(chRes chan *ParsedWS, src *gopacket.PacketSource) {

	defer close(chRes)
	rowNo := 0
	counter := 0

	for packet := range src.Packets() {
		rowNo++
		if packet == nil {
			continue
		}

		if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
			continue
		}

		ip := packet.NetworkLayer().(*layers.IPv4)
		tcp := packet.TransportLayer().(*layers.TCP)

		// if rowNo == 14125 {
		// 	rowNo = 14125
		// }
		ws := NewWS(tcp.Payload)
		if ws == nil {
			continue
		}
		isValid, protoMsg := ws.IsValid()
		if isValid {
			counter++
			res := &ParsedWS{
				RowNo:   rowNo,
				Counter: counter,
				SrcIP:   ip.SrcIP,
				DstIP:   ip.DstIP,
				SrcPort: uint16(tcp.SrcPort),
				DstPort: uint16(tcp.DstPort),
				Message: protoMsg,
			}
			chRes <- res
		}
	}

}

func parsePackects_withAssembly(chRes chan *ParsedWS, src *gopacket.PacketSource) {

	defer close(chRes)

	// Set up assembly
	streamFactory := &wsFactory{
		ch: chRes,
	}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)
	packets := src.Packets()
	ticker := time.Tick(shared.DefaultTimeout)
	// Read in packets, pass to assembler.
	for {
		select {
		case packet := <-packets:
			// nil packet is EOF
			if packet == nil {
				assembler.FlushAll()
				return
			}
			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				continue
			}
			tcp := packet.TransportLayer().(*layers.TCP)
			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)
		case <-ticker:
			assembler.FlushOlderThan(time.Now().Add(shared.DefaultTimeout * -1))
		}
	}

}

// WSFactory implements tcpassembly.StreamFactory
type wsFactory struct {
	ch chan *ParsedWS
}

func (h *wsFactory) New(net, tcp gopacket.Flow) tcpassembly.Stream {
	s := &wsStream{
		ch:   h.ch,
		data: make([]byte, 0),
		net:  net,
		tcp:  tcp,
	}
	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return s
}

// httpStream will handle the actual decoding of http requests.
type wsStream struct {
	ch       chan *ParsedWS
	data     []byte
	net, tcp gopacket.Flow
}

// Reassembled is called whenever new packet data is available for reading.
// Reassembly objects contain stream data IN ORDER.
func (s *wsStream) Reassembled(reassemblies []tcpassembly.Reassembly) {

	buff := make([]byte, 0)
	for _, reassembly := range reassemblies {
		buff = append(buff, reassembly.Bytes...)
	}

	n := len(buff)
	if n > 0 {
		s.data = make([]byte, n, n)
		copy(s.data, buff)
	}

}

// ReassemblyComplete is called when the TCP assembler believes a stream has
func (s *wsStream) ReassemblyComplete() {

	ws := NewWS(s.data)
	if ws == nil {
		return
	}
	isValid, protoMsg := ws.IsValid()

	if isValid {
		res := &ParsedWS{
			RowNo:   0,
			Counter: 0,
			SrcIP:   s.net.Src().Raw(),
			DstIP:   s.net.Dst().Raw(),
			SrcPort: binary.BigEndian.Uint16(s.tcp.Src().Raw()),
			DstPort: binary.BigEndian.Uint16(s.tcp.Dst().Raw()),
			Message: protoMsg,
		}
		s.ch <- res
	}

}
