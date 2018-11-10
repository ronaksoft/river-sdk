package repo

import (
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/domain"

	"git.ronaksoftware.com/ronak/riversdk/msg"

	_ "github.com/mattn/go-sqlite3"
)

func BenchmarkInsert(b *testing.B) {
	err := InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		m := new(msg.UserMessage)
		m.ID = domain.SequentialUniqueID()
		m.PeerID = 123456789
		m.PeerType = 1
		m.CreatedOn = time.Now().Unix()
		m.Body = fmt.Sprintf("Test %v", i)
		m.SenderID = 987654321
		Ctx().Messages.SaveMessage(m)
	}

	// os.Remove("river.db")
}

func BenchmarkInsertRAW(b *testing.B) {
	err := InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}

	// insert
	stmt, err := db.Prepare(`INSERT INTO messages 
	(
	ID,
	PeerID,
	PeerType,
	CreatedOn,
	Body,
	SenderID,
	EditedOn,
	FwdSenderID	,
	FwdChannelID,
	FwdChannelMessageID	,
	Flags	,
	MessageType	,
	ContentRead	,
	Inbox,
	ReplyTo	,
	MessageAction
	)
	VALUES
	(
	?,
	?,
	?,
	?,
	?,
	?,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0
	)`)

	if err != nil {
		log.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
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

	// os.Remove("river.db")
}
