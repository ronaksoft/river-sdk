package tools

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

/*
   Creation Time: 2019 - Nov - 25
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func stringPrinter(s string) {
	for i := 0; i < 5; i++ {
		fmt.Println(s)
		time.Sleep(time.Second)
	}
}

func bytesPrinter(b []byte) {
	for i := 0; i < 5; i++ {
		fmt.Println(b)
		time.Sleep(time.Second)
	}
}

// func TestByteToStr1(t *testing.T) {
// 	b := []byte("Sample Text")
// 	go stringPrinter(ByteToStr(b))
// 	for i := 0; i < 10; i++ {
// 		j := RandomInt(len(b))
// 		b[j] = []byte(RandomID(1))[0]
// 		time.Sleep(time.Millisecond * 100)
// 	}
// }
//
// func TestByteToStr2(t *testing.T) {
// 	b := []byte("Sample Text")
// 	go stringPrinter(ByteToStr(b))
// 	for i := 0; i < 10; i++ {
// 		b = []byte(RandomID(10))
// 		_ = b
// 		time.Sleep(time.Millisecond * 100)
// 		runtime.GC()
// 	}
// }
//
// func TestStrToByte(t *testing.T) {
// 	s := "Sample Text"
// 	go bytesPrinter(StrToByte(s))
// 	for i := 0; i < 10; i++ {
// 		s = RandomID(10)
// 		time.Sleep(time.Millisecond * 100)
// 	}
// }

func BenchmarkRandomInt64(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = RandomInt64(0)
		}
	})
}

func BenchmarkRandomID(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = RandomID(10)
		}
	})
}

func TestRandomID(t *testing.T) {
	x := RandomID(10)
	fmt.Println(x)
	for i := 0; i < 1000; i++ {
		RandomID(10)
	}

	time.Sleep(time.Second)
	runtime.GC()
	fmt.Println(x)
}

func TestDeleteItemFromArray(t *testing.T) {
	x := []int32{1, 2, 3, 4, 5, 6, 7, 8}
	fmt.Println("Before:", x)
	DeleteItemFromArray(&x, 3)
	fmt.Println("After:", x)

	for idx := 0; idx < len(x); idx++ {
		DeleteItemFromArray(&x, idx)
		fmt.Println("Deleted:", idx, x)
		idx--
	}
	fmt.Println("After All:", x)
}
