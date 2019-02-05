package pcap_parser

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func Parse(pcapFile string) (chan *ParsedWS, error) {

	result := make(chan *ParsedWS)

	handle, err := pcap.OpenOffline(pcapFile)
	if err != nil {
		return nil, err
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	go parsePackects(result, packetSource)

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
