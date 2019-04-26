package pcap_parser

import (
	"bytes"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/shared"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func Parse_wsutil(pcapFile string) (chan *ParsedWS, error) {

	result := make(chan *ParsedWS)

	handle, err := pcap.OpenOffline(pcapFile)
	if err != nil {
		return nil, err
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	go parsePackects_wsutil(result, packetSource)

	return result, nil

}

func parsePackects_wsutil(chRes chan *ParsedWS, src *gopacket.PacketSource) {

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
		r := bytes.NewReader(tcp.Payload)
		// Read message that client sends to server
		if uint16(tcp.SrcPort) != shared.ServerPort {
			msgs := make([]wsutil.Message, 0)
			var err error
			msgs, err = wsutil.ReadClientMessage(r, msgs)
			if err == nil {
				for _, m := range msgs {
					protoMsg := new(msg.ProtoMessage)
					err = protoMsg.Unmarshal(m.Payload)
					if err == nil {
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
		}

		// Read message that server sends to client
		if uint16(tcp.SrcPort) == shared.ServerPort {
			msgs := make([]wsutil.Message, 0)
			var err error
			msgs, err = wsutil.ReadServerMessage(r, msgs)
			if err == nil {
				for _, m := range msgs {
					protoMsg := new(msg.ProtoMessage)
					err = protoMsg.Unmarshal(m.Payload)
					if err == nil {
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
		}

	}
}
