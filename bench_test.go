package main

import (
	"testing"

	v "git.vegner.org/vsvegner/gomtc/internal/app/validation"
)

func BenchmarkNetParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if !v.IsIP("192.168.65.254") {
			b.Fatalf("Unexpected string: 192.168.65.254")
		}
	}
}

// func BenchmarkNetaddrParse(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		if !isIPnetaddr("192.168.65.254") {
// 			b.Fatalf("Unexpected string: 192.168.65.254")
// 		}
// 	}
// }
// func BenchmarkStrconvAtoi(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		if !is_ipv4("192.168.65.254") {
// 			b.Fatalf("Unexpected string: 192.168.65.254")
// 		}
// 	}
// }
// func BenchmarkRegexp(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		if !validIP4("192.168.65.254") {
// 			b.Fatalf("Unexpected string: 192.168.65.254")
// 		}
// 	}
// }
func BenchmarkValidateMac(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if !v.IsMac("00:00:00:FF:FF:FF") {
			b.Fatalf("Unexpected string: 00:00:00:FF:FF:FF")
		}
	}
}

// func BenchmarkIsMac(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		if !validateMac("00:00:00:FF:FF:FF") {
// 			b.Fatalf("Unexpected string: 00:00:00:FF:FF:FF")
// 		}
// 	}
// }

// func BenchmarkIsMacNetaddr(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		if !isMacNetaddr("00:00:00:FF:FF:FF") {
// 			b.Fatalf("Unexpected string: 00:00:00:FF:FF:FF")
// 		}
// 	}
// }
