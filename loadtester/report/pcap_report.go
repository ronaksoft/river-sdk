package report

import (
	"git.ronaksoftware.com/ronak/riversdk/loadtester/pcap_parser"
)

type PcapReport struct {
}

func NewPcapReport() *PcapReport {
	r := new(PcapReport)
	return r
}

func (r *PcapReport) Feed(p *pcap_parser.ParsedWS) {

}

func (r *PcapReport) String() string {
	return ""
}
