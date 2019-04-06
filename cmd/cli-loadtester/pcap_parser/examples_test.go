package pcap_parser

import (
	"testing"
	"time"
)

func TestParser(t *testing.T) {
	pcapFilePath := "dump.pcap"

	// res, err := Parse_wsutil(pcapFilePath)
	res, err := Parse(pcapFilePath)
	if err != nil {
		t.Error(err)
	}

	sw := time.Now()
	for r := range res {
		t.Logf("%d \t\t %d \t\t SRC: %v:%v \t DES: %v:%v \t\t AuthID:%d\n", r.RowNo, r.Counter, r.SrcIP, r.SrcPort, r.DstIP, r.DstPort, r.Message.AuthID)
	}
	elapsed := time.Since(sw)
	t.Logf("Elapsed : %v", elapsed)
}

func TestParser_wsutil(t *testing.T) {
	pcapFilePath := "dump.pcap"

	res, err := Parse_wsutil(pcapFilePath)
	if err != nil {
		t.Error(err)
	}

	sw := time.Now()
	for r := range res {
		t.Logf("%d \t\t %d \t\t SRC: %v:%v \t DES: %v:%v \t\t AuthID:%d\n", r.RowNo, r.Counter, r.SrcIP, r.SrcPort, r.DstIP, r.DstPort, r.Message.AuthID)
	}
	elapsed := time.Since(sw)
	t.Logf("Elapsed : %v", elapsed)
}
