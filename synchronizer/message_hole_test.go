package synchronizer

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"git.ronaksoftware.com/ronak/riversdk/repo/dto"

	"git.ronaksoftware.com/ronak/riversdk/repo"
)

func TestIsMessageInHole(t *testing.T) {
	// Create DB and Tables
	err := repo.InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	repo.Ctx().LogMode(true)
	defer repo.Ctx().Close()
	defer os.Remove("river.db")

	// Insert Chunck Data
	dtoHoles := []dto.MessageHoles{
		dto.MessageHoles{
			PeerID: 1,
			MinID:  sql.NullInt64{Int64: 10, Valid: true},
			MaxID:  20,
		},
		dto.MessageHoles{
			PeerID: 1,
			MinID:  sql.NullInt64{Int64: 30, Valid: true},
			MaxID:  40,
		},
		dto.MessageHoles{
			PeerID: 1,
			MinID:  sql.NullInt64{Int64: 50, Valid: true},
			MaxID:  60,
		},
	}

	for _, v := range dtoHoles {
		err := repo.Ctx().MessageHoles.Save(v.PeerID, v.MinID.Int64, v.MaxID)
		if err != nil {
			t.Error(err)
		}
	}

	// surrended mode
	holes, err := repo.Ctx().MessageHoles.GetHoles(1, 0, 61)
	if len(holes) != 3 || err != nil {
		t.Error("Count is not equal to 3")
	}
	//minside mode
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 5, 15)
	if len(holes) != 1 || err != nil {
		t.Error("minside mode")
	}

	//maxside mode
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 15, 25)
	if len(holes) != 1 || err != nil {
		t.Error("maxside mode")
	}

	//inside mode
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 12, 18)
	if len(holes) != 1 || err != nil {
		t.Error("inside mode")
	}

	//inside exact size mode
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 10, 20)
	if len(holes) != 1 || err != nil {
		t.Error("inside exact size mode")
	}

	//inside min overlap
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 10, 15)
	if len(holes) != 1 || err != nil {
		t.Error("inside min overlap")
	}
	//inside max overlap
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 15, 20)
	if len(holes) != 1 || err != nil {
		t.Error("inside max overlap")
	}
}

func TestFillMessageHolesSurrendedMode(t *testing.T) {
	// Create DB and Tables
	err := repo.InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	repo.Ctx().LogMode(true)
	defer repo.Ctx().Close()
	defer os.Remove("river.db")

	// Insert CHunck Data
	dtoHoles := []dto.MessageHoles{
		dto.MessageHoles{
			PeerID: 1,
			MinID:  sql.NullInt64{Int64: 10, Valid: true},
			MaxID:  20,
		},
		dto.MessageHoles{
			PeerID: 1,
			MinID:  sql.NullInt64{Int64: 30, Valid: true},
			MaxID:  40,
		},
		dto.MessageHoles{
			PeerID: 1,
			MinID:  sql.NullInt64{Int64: 50, Valid: true},
			MaxID:  60,
		},
	}

	for _, v := range dtoHoles {
		err := repo.Ctx().MessageHoles.Save(v.PeerID, v.MinID.Int64, v.MaxID)
		if err != nil {
			t.Error(err)
		}
	}

	// surrended mode
	holes, err := repo.Ctx().MessageHoles.GetHoles(1, 0, 61)
	if len(holes) != 3 || err != nil {
		t.Error("Count is not equal to 3")
	}

	err = fillMessageHoles(1, 0, 61)
	if err != nil {
		t.Error(err)
	}

	// TODO : Check The result
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 0, 61)
	if len(holes) != 0 || err != nil {
		t.Error("Count is not equal to 0")
	}
}

func TestFillMessageHolesMinSideMode(t *testing.T) {
	// Create DB and Tables
	err := repo.InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	repo.Ctx().LogMode(true)
	defer repo.Ctx().Close()
	defer os.Remove("river.db")

	// Insert CHunck Data
	dtoHoles := []dto.MessageHoles{
		dto.MessageHoles{
			PeerID: 1,
			MinID:  sql.NullInt64{Int64: 10, Valid: true},
			MaxID:  20,
		},
	}

	for _, v := range dtoHoles {
		err := repo.Ctx().MessageHoles.Save(v.PeerID, v.MinID.Int64, v.MaxID)
		if err != nil {
			t.Error(err)
		}
	}

	//minside mode
	holes, err := repo.Ctx().MessageHoles.GetHoles(1, 5, 15)
	if len(holes) != 1 || err != nil {
		t.Error("minside mode")
	}

	err = fillMessageHoles(1, 5, 15)
	if err != nil {
		t.Error(err)
	}

	// TODO : Check The result
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 5, 15)
	if len(holes) != 0 || err != nil {
		t.Error("minside mode find hole again")
	}
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 16, 20)
	if len(holes) != 1 || err != nil {
		t.Error("minside mode no hole created")
	}
	if holes[0].MinID.Int64 != 16 || holes[0].MaxID != 20 {
		t.Error("minside mode create wrong hole")
	}
}

func TestFillMessageHolesMaxSideMode(t *testing.T) {
	// Create DB and Tables
	err := repo.InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	repo.Ctx().LogMode(true)
	defer repo.Ctx().Close()
	defer os.Remove("river.db")

	// Insert CHunck Data
	dtoHoles := []dto.MessageHoles{
		dto.MessageHoles{
			PeerID: 1,
			MinID:  sql.NullInt64{Int64: 10, Valid: true},
			MaxID:  20,
		},
	}

	for _, v := range dtoHoles {
		err := repo.Ctx().MessageHoles.Save(v.PeerID, v.MinID.Int64, v.MaxID)
		if err != nil {
			t.Error(err)
		}
	}

	//maxside mode
	holes, err := repo.Ctx().MessageHoles.GetHoles(1, 15, 25)
	if len(holes) != 1 || err != nil {
		t.Error("maxside mode")
	}

	err = fillMessageHoles(1, 15, 25)
	if err != nil {
		t.Error(err)
	}

	// TODO : Check The result
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 15, 25)
	if len(holes) != 0 || err != nil {
		t.Error("maxside mode find hole again")
	}
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 10, 14)
	if len(holes) != 1 || err != nil {
		t.Error("maxside mode no hole created")
	}
	if holes[0].MinID.Int64 != 10 || holes[0].MaxID != 14 {
		t.Error("maxside mode create wrong hole")
	}
}

func TestFillMessageHolesInsideMode(t *testing.T) {
	// Create DB and Tables
	err := repo.InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	repo.Ctx().LogMode(true)
	defer repo.Ctx().Close()
	defer os.Remove("river.db")

	// Insert CHunck Data
	dtoHoles := []dto.MessageHoles{
		dto.MessageHoles{
			PeerID: 1,
			MinID:  sql.NullInt64{Int64: 10, Valid: true},
			MaxID:  20,
		},
	}

	for _, v := range dtoHoles {
		err := repo.Ctx().MessageHoles.Save(v.PeerID, v.MinID.Int64, v.MaxID)
		if err != nil {
			t.Error(err)
		}
	}

	//inside mode
	holes, err := repo.Ctx().MessageHoles.GetHoles(1, 12, 18)
	if len(holes) != 1 || err != nil {
		t.Error("inside mode")
	}
	err = fillMessageHoles(1, 12, 18)
	if err != nil {
		t.Error(err)
	}

	// TODO : Check The result
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 12, 18)
	if len(holes) != 0 || err != nil {
		t.Error("inside mode find hole again")
	}
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 10, 11)
	if len(holes) != 1 || err != nil {
		t.Error("inside mode no hole created")
	}
	if holes[0].MinID.Int64 != 10 || holes[0].MaxID != 11 {
		t.Error("inside mode create wrong hole")
	}
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 19, 20)
	if len(holes) != 1 || err != nil {
		t.Error("inside mode no hole created")
	}
	if holes[0].MinID.Int64 != 19 || holes[0].MaxID != 20 {
		t.Error("inside mode create wrong hole")
	}
}

func TestFillMessageHolesInsideExactMode(t *testing.T) {
	// Create DB and Tables
	err := repo.InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	repo.Ctx().LogMode(true)
	defer repo.Ctx().Close()
	defer os.Remove("river.db")

	// Insert CHunck Data
	dtoHoles := []dto.MessageHoles{
		dto.MessageHoles{
			PeerID: 1,
			MinID:  sql.NullInt64{Int64: 10, Valid: true},
			MaxID:  20,
		},
	}

	for _, v := range dtoHoles {
		err := repo.Ctx().MessageHoles.Save(v.PeerID, v.MinID.Int64, v.MaxID)
		if err != nil {
			t.Error(err)
		}
	}

	//inside exact size mode
	holes, err := repo.Ctx().MessageHoles.GetHoles(1, 10, 20)
	if len(holes) != 1 || err != nil {
		t.Error("inside exact size mode")
	}

	err = fillMessageHoles(1, 10, 20)
	if err != nil {
		t.Error(err)
	}

	// TODO : Check The result
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 10, 20)
	if len(holes) != 0 || err != nil {
		t.Error("inside exact mode find hole again")
	}
}

func TestFillMessageHolesInsideMinOverlap(t *testing.T) {
	// Create DB and Tables
	err := repo.InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	repo.Ctx().LogMode(true)
	defer repo.Ctx().Close()
	defer os.Remove("river.db")

	// Insert CHunck Data
	dtoHoles := []dto.MessageHoles{
		dto.MessageHoles{
			PeerID: 1,
			MinID:  sql.NullInt64{Int64: 10, Valid: true},
			MaxID:  20,
		},
	}

	for _, v := range dtoHoles {
		err := repo.Ctx().MessageHoles.Save(v.PeerID, v.MinID.Int64, v.MaxID)
		if err != nil {
			t.Error(err)
		}
	}

	//inside min overlap
	holes, err := repo.Ctx().MessageHoles.GetHoles(1, 10, 15)
	if len(holes) != 1 || err != nil {
		t.Error("inside min overlap")
	}

	err = fillMessageHoles(1, 10, 15)
	if err != nil {
		t.Error(err)
	}

	// TODO : Check The result
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 10, 15)
	if len(holes) != 0 || err != nil {
		t.Error("inside min overlap mode find hole again")
	}
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 16, 20)
	if len(holes) != 1 || err != nil {
		t.Error("inside min overlap mode no hole created")
	}
	if holes[0].MinID.Int64 != 16 || holes[0].MaxID != 20 {
		t.Error("inside min overlap mode create wrong hole")
	}
}

func TestFillMessageHolesInsideMaxOverlap(t *testing.T) {
	// Create DB and Tables
	err := repo.InitRepo("sqlite3", "river.db")
	if err != nil {
		log.Fatal(err)
	}
	repo.Ctx().LogMode(true)
	defer repo.Ctx().Close()
	defer os.Remove("river.db")

	// Insert CHunck Data
	dtoHoles := []dto.MessageHoles{
		dto.MessageHoles{
			PeerID: 1,
			MinID:  sql.NullInt64{Int64: 10, Valid: true},
			MaxID:  20,
		},
	}

	for _, v := range dtoHoles {
		err := repo.Ctx().MessageHoles.Save(v.PeerID, v.MinID.Int64, v.MaxID)
		if err != nil {
			t.Error(err)
		}
	}

	//inside max overlap
	holes, err := repo.Ctx().MessageHoles.GetHoles(1, 15, 20)
	if len(holes) != 1 || err != nil {
		t.Error("inside max overlap")
	}

	err = fillMessageHoles(1, 15, 20)
	if err != nil {
		t.Error(err)
	}

	// TODO : Check The result

	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 15, 20)
	if len(holes) != 0 || err != nil {
		t.Error("inside max overlap mode find hole again")
	}
	holes, err = repo.Ctx().MessageHoles.GetHoles(1, 10, 14)
	if len(holes) != 1 || err != nil {
		t.Error("inside max overlap mode no hole created")
	}
	if holes[0].MinID.Int64 != 10 || holes[0].MaxID != 14 {
		t.Error("inside max overlap mode create wrong hole")
	}
}

func TestGetMaxClosestHole(t *testing.T) {

	holes := []dto.MessageHoles{
		dto.MessageHoles{
			PeerID: 1,
			MinID:  sql.NullInt64{Int64: 0, Valid: true},
			MaxID:  692,
		},
	}
	dto := GetMaxClosestHole(982, holes)
	if dto == nil {
		t.Error("Failed :/ ")
	}

}

func TestGetMonClosestHole(t *testing.T) {

	holes := []dto.MessageHoles{
		dto.MessageHoles{
			PeerID: 1,
			MinID:  sql.NullInt64{Int64: 0, Valid: true},
			MaxID:  635,
		},
	}
	dto := GetMinClosestHole(933, holes)
	if dto == nil {
		t.Error("Failed :/ ")
	}

}
