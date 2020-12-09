package riversdk

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"github.com/dgraph-io/badger/v2"
	"go.uber.org/zap"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/*
   Creation Time: 2020 - Jul - 25
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

// GenSrpHash generates a hash to be used in AuthCheckPassword and other related apis
func GenSrpHash(password []byte, algorithm int64, algorithmData []byte) []byte {
	switch algorithm {
	case msg.C_PasswordAlgorithmVer6A:
		algo := &msg.PasswordAlgorithmVer6A{}
		err := algo.Unmarshal(algorithmData)

		if err != nil {
			logs.Warn("Error On Unmarshal Algorithm", zap.Error(err))
			return nil
		}

		p := big.NewInt(0).SetBytes(algo.P)
		x := big.NewInt(0).SetBytes(domain.PH2(password, algo.Salt1, algo.Salt2))
		v := big.NewInt(0).Exp(big.NewInt(int64(algo.G)), x, p)

		return v.Bytes()
	default:
		return nil
	}
}

// GenInputPassword  accepts AccountPassword marshaled as argument and return InputPassword marshaled
func GenInputPassword(password []byte, accountPasswordBytes []byte) []byte {
	ap := &msg.AccountPassword{}
	err := ap.Unmarshal(accountPasswordBytes)

	algo := &msg.PasswordAlgorithmVer6A{}
	err = algo.Unmarshal(ap.AlgorithmData)
	if err != nil {
		logs.Warn("Error On GenInputPassword", zap.Error(err))
		return nil
	}

	p := big.NewInt(0).SetBytes(algo.P)
	g := big.NewInt(0).SetInt64(int64(algo.G))
	k := big.NewInt(0).SetBytes(domain.K(p, g))

	x := big.NewInt(0).SetBytes(domain.PH2(password, algo.Salt1, algo.Salt2))
	v := big.NewInt(0).Exp(g, x, p)
	a := big.NewInt(0).SetBytes(ap.RandomData)
	ga := big.NewInt(0).Exp(g, a, p)
	gb := big.NewInt(0).SetBytes(ap.SrpB)
	u := big.NewInt(0).SetBytes(domain.U(ga, gb))
	kv := big.NewInt(0).Mod(big.NewInt(0).Mul(k, v), p)
	t := big.NewInt(0).Mod(big.NewInt(0).Sub(gb, kv), p)
	if t.Sign() < 0 {
		t.Add(t, p)
	}
	sa := big.NewInt(0).Exp(t, big.NewInt(0).Add(a, big.NewInt(0).Mul(u, x)), p)
	m1 := domain.MM(p, g, algo.Salt1, algo.Salt2, ga, gb, sa)

	inputPassword := &msg.InputPassword{
		SrpID: ap.SrpID,
		A:     domain.Pad(ga),
		M1:    m1,
	}
	res, _ := inputPassword.Marshal()
	return res
}

// SanitizeQuestionAnswer
func SanitizeQuestionAnswer(answer string) string {
	return strings.ToLower(strings.TrimSpace(answer))
}

// GetCountryCode
func GetCountryCode(phone string) string {
	return domain.GetCountryCode(phone)
}

// Version returns the version of the current SDK
func Version() string {
	return domain.SDKVersion
}

// BadgerSupport identifies the version of the badger
func BadgerSupport(dbDir string) bool {
	if strings.HasPrefix(dbDir, "file://") {
		dbDir = dbDir[7:]
	}
	f, err := os.Open(filepath.Join(dbDir, "MANIFEST"))
	if err != nil {
		return false
	}
	_, _, err = badger.ReplayManifestFile(f)
	if err != nil {
		return false
	}
	return true
}

// LastSeenEstimate
func LastSeenEstimate(ts int64) int {
	if ts == 0 {
		return domain.LastSeenUnknown
	}

	now := domain.Now()
	seen := time.Unix(ts, 0)

	if now.Unix() < seen.Unix() {
		return domain.LastSeenUnknown
	}

	// We do Waterfall check
	// 1. FewSeconds / Recently
	if now.Sub(seen) < time.Minute {
		return domain.LastSeenFewSeconds
	}
	if now.Sub(seen) < time.Hour {
		return domain.LastSeenFewMinutes
	}

	ny, nw := now.ISOWeek()
	sy, sw := seen.ISOWeek()
	// 2. Same Week/Year
	if ny == sy && nw == sw {
		if now.YearDay() == seen.YearDay() {
			return domain.LastSeenToday
		} else if now.YearDay() == seen.YearDay()+1 {
			return domain.LastSeenYesterday
		}
		return domain.LastSeenThisWeek
	}

	// 3. Same Year
	if ny == sy {
		if nw == sw+1 {
			return domain.LastSeenLastWeek
		}
		if now.Month() == seen.Month() {
			return domain.LastSeenThisMonth
		}
	}

	return domain.LastSeenLongTimeAgo
}
