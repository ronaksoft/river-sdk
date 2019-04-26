package repo

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"

	"github.com/doug-martin/goqu"
	_ "github.com/mattn/go-sqlite3"
)

func BenchmarkInsertORM(b *testing.B) {
	// Create DB and Tables
	err := InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	defer Ctx().Close()
	//defer os.Remove("river.db")

	b.ResetTimer()

	for i := 0; i < 100; i++ {
		m := new(msg.UserMessage)
		m.ID = domain.SequentialUniqueID()
		m.PeerID = 123456789
		m.PeerType = 1
		m.CreatedOn = time.Now().Unix()
		m.Body = fmt.Sprintf("Test %v", i)
		m.SenderID = 987654321
		Ctx().Messages.SaveMessage(m)
	}

}

func BenchmarkInsertRAW(b *testing.B) {
	// Create DB and Tables
	err := InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	defer Ctx().Close()
	//defer os.Remove("river.db")

	// Open DB
	db, err := sql.Open("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// insert
	stmt, err := db.Prepare(`INSERT INTO messages 
	( ID, PeerID, PeerType, CreatedOn, Body, SenderID, EditedOn, FwdSenderID, FwdChannelID, FwdChannelMessageID, Flags, MessageType, ContentRead, Inbox, ReplyTo, MessageAction )
	VALUES
	(?,?,?,?,?,?,0,0,0,0,0,0,0,0,0,0)`)

	if err != nil {
		log.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < 100; i++ {
		m := new(msg.UserMessage)
		m.ID = domain.SequentialUniqueID()
		m.PeerID = 123456789
		m.PeerType = 1
		m.CreatedOn = time.Now().Unix()
		m.Body = fmt.Sprintf("Test %v", i)
		m.SenderID = 987654321

		_, err := stmt.Exec(m.ID, m.PeerID, m.PeerType, m.CreatedOn, m.Body, m.SenderID)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func BenchmarkInsertBatch(b *testing.B) {
	// Create DB and Tables
	err := InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	defer Ctx().Close()
	//defer os.Remove("river.db")

	db, err := sql.Open("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	batchSB := new(strings.Builder)

	for i := 0; i < 100; i++ {
		m := new(msg.UserMessage)
		m.ID = domain.SequentialUniqueID()
		m.PeerID = 123456789
		m.PeerType = 1
		m.CreatedOn = time.Now().Unix()
		m.Body = fmt.Sprintf("Test %v", i)
		m.SenderID = 987654321

		qb := goqu.New("", nil)

		str := qb.From("messages").Insert(goqu.Record{
			"ID":                  m.ID,
			"PeerID":              m.PeerID,
			"PeerType":            m.PeerType,
			"CreatedOn":           m.CreatedOn,
			"Body":                m.Body,
			"SenderID":            m.SenderID,
			"EditedOn":            m.EditedOn,
			"FwdSenderID":         m.SenderID,
			"FwdChannelID":        m.FwdChannelID,
			"FwdChannelMessageID": m.FwdChannelMessageID,
			"Flags":               m.Flags,
			"MessageType":         m.MessageType,
			"ContentRead":         m.ContentRead,
			"Inbox":               m.Inbox,
			"ReplyTo":             m.ReplyTo,
			"MessageAction":       m.MessageAction,
		}).Sql
		batchSB.WriteString(str + ";")
	}
	qry := batchSB.String()

	b.ResetTimer()

	_, err = db.Exec(qry)
	if err != nil {
		log.Fatal(err)
	}
}
