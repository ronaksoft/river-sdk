package domain

import (
	"crypto/sha256"
	"math/big"
)

/*
   Creation Time: 2020 - Jan - 11
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

const (
	bitSize     = 2048
	SrpByteSize = bitSize / 8
)

func H(data ...[]byte) []byte {
	h := sha256.New()
	for _, d := range data {
		h.Write(d)
	}
	return h.Sum(nil)
}

func SH(data, salt []byte) []byte {
	b := make([]byte, 0, len(data)+2*len(salt))
	b = append(b, salt...)
	b = append(b, data...)
	b = append(b, salt...)
	return H(b)
}

func PH1(password, salt1, salt2 []byte) []byte {
	return SH(SH(password, salt1), salt2)
}

func PH2(password, salt1, salt2 []byte) []byte {
	return SH(SH(PH1(password, salt1, salt2), salt1), salt2)
}

func K(p, g *big.Int) []byte {
	return H(append(Pad(p), Pad(g)...))
}

func U(ga, gb *big.Int) []byte {
	return H(append(Pad(ga), Pad(gb)...))
}

func MM(p, g *big.Int, s1, s2 []byte, ga, gb, sb *big.Int) []byte {
	return H(H(Pad(p)), H(Pad(g)), H(s1), H(s2), H(Pad(ga)), H(Pad(gb)), H(Pad(sb)))
}

// Pad x to n bytes if needed
func Pad(x *big.Int) []byte {
	b := x.Bytes()
	if len(b) < SrpByteSize {
		z := SrpByteSize - len(b)
		p := make([]byte, SrpByteSize)
		for i := 0; i < z; i++ {
			p[i] = 0
		}

		copy(p[z:], b)
		b = p
	}
	return b
}
