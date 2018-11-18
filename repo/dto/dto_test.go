package dto

import (
	"testing"
)

func BenchmarkMapJson(b *testing.B) {
	src := &Dialogs{
		AccessHash:      1,
		LastUpdate:      2,
		NotifyFlags:     3,
		NotifyMuteUntil: 4,
		NotifySound:     "0",
		PeerID:          5,
		PeerType:        6,
		ReadInboxMaxID:  7,
		ReadOutboxMaxID: 8,
		TopMessageID:    9,
		UnreadCount:     10,
	}
	des := new(Dialogs)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		MapJson(src, des)
	}

	// fmt.Println("src :", src)
	// fmt.Println("des :", des)
}

func BenchmarkMap(b *testing.B) {
	src := &Dialogs{
		AccessHash:      1,
		LastUpdate:      2,
		NotifyFlags:     3,
		NotifyMuteUntil: 4,
		NotifySound:     "0",
		PeerID:          5,
		PeerType:        6,
		ReadInboxMaxID:  7,
		ReadOutboxMaxID: 8,
		TopMessageID:    9,
		UnreadCount:     10,
	}
	des := new(Dialogs)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Map(src, des)
	}

	// fmt.Println("src :", src)
	// fmt.Println("des :", des)
}
