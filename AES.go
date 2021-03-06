package main

import (
	"encoding/hex"
	"math/rand"
)

// GF256 is a variable in Galois field of 2^8 that is used by AES
type GF256 struct {
	i uint8
}

func (B0 GF256) add(B1 GF256) GF256 {
	return GF256{B0.i ^ B1.i}
}

func (B0 GF256) mult(B1 GF256) GF256 {
	var a, b, p uint8
	a = B0.i
	b = B1.i
	for index := 0; index < 8; index++ {
		if b&1 != 0 {
			p = p ^ a
		}
		carry := (a&0x80 != 0)
		a = a << 1
		if carry {
			a = a ^ 0b11011
		}
		b = b >> 1
	}
	return GF256{p}
}

func (B0 GF256) rotl8(n int) GF256 {
	return GF256{B0.i<<n | B0.i>>(8-n)}

}

func (B0 GF256) sbox() GF256 {
	box := [256]uint8{
		0x63, 0x7c, 0x77, 0x7b, 0xf2, 0x6b, 0x6f, 0xc5, 0x30, 0x01, 0x67, 0x2b, 0xfe, 0xd7, 0xab, 0x76,
		0xca, 0x82, 0xc9, 0x7d, 0xfa, 0x59, 0x47, 0xf0, 0xad, 0xd4, 0xa2, 0xaf, 0x9c, 0xa4, 0x72, 0xc0,
		0xb7, 0xfd, 0x93, 0x26, 0x36, 0x3f, 0xf7, 0xcc, 0x34, 0xa5, 0xe5, 0xf1, 0x71, 0xd8, 0x31, 0x15,
		0x04, 0xc7, 0x23, 0xc3, 0x18, 0x96, 0x05, 0x9a, 0x07, 0x12, 0x80, 0xe2, 0xeb, 0x27, 0xb2, 0x75,
		0x09, 0x83, 0x2c, 0x1a, 0x1b, 0x6e, 0x5a, 0xa0, 0x52, 0x3b, 0xd6, 0xb3, 0x29, 0xe3, 0x2f, 0x84,
		0x53, 0xd1, 0x00, 0xed, 0x20, 0xfc, 0xb1, 0x5b, 0x6a, 0xcb, 0xbe, 0x39, 0x4a, 0x4c, 0x58, 0xcf,
		0xd0, 0xef, 0xaa, 0xfb, 0x43, 0x4d, 0x33, 0x85, 0x45, 0xf9, 0x02, 0x7f, 0x50, 0x3c, 0x9f, 0xa8,
		0x51, 0xa3, 0x40, 0x8f, 0x92, 0x9d, 0x38, 0xf5, 0xbc, 0xb6, 0xda, 0x21, 0x10, 0xff, 0xf3, 0xd2,
		0xcd, 0x0c, 0x13, 0xec, 0x5f, 0x97, 0x44, 0x17, 0xc4, 0xa7, 0x7e, 0x3d, 0x64, 0x5d, 0x19, 0x73,
		0x60, 0x81, 0x4f, 0xdc, 0x22, 0x2a, 0x90, 0x88, 0x46, 0xee, 0xb8, 0x14, 0xde, 0x5e, 0x0b, 0xdb,
		0xe0, 0x32, 0x3a, 0x0a, 0x49, 0x06, 0x24, 0x5c, 0xc2, 0xd3, 0xac, 0x62, 0x91, 0x95, 0xe4, 0x79,
		0xe7, 0xc8, 0x37, 0x6d, 0x8d, 0xd5, 0x4e, 0xa9, 0x6c, 0x56, 0xf4, 0xea, 0x65, 0x7a, 0xae, 0x08,
		0xba, 0x78, 0x25, 0x2e, 0x1c, 0xa6, 0xb4, 0xc6, 0xe8, 0xdd, 0x74, 0x1f, 0x4b, 0xbd, 0x8b, 0x8a,
		0x70, 0x3e, 0xb5, 0x66, 0x48, 0x03, 0xf6, 0x0e, 0x61, 0x35, 0x57, 0xb9, 0x86, 0xc1, 0x1d, 0x9e,
		0xe1, 0xf8, 0x98, 0x11, 0x69, 0xd9, 0x8e, 0x94, 0x9b, 0x1e, 0x87, 0xe9, 0xce, 0x55, 0x28, 0xdf,
		0x8c, 0xa1, 0x89, 0x0d, 0xbf, 0xe6, 0x42, 0x68, 0x41, 0x99, 0x2d, 0x0f, 0xb0, 0x54, 0xbb, 0x16}
	return GF256{box[B0.i]}
}

func (B0 GF256) invsbox() GF256 {
	box := [256]uint8{
		0x52, 0x09, 0x6a, 0xd5, 0x30, 0x36, 0xa5, 0x38, 0xbf, 0x40, 0xa3, 0x9e, 0x81, 0xf3, 0xd7, 0xfb,
		0x7c, 0xe3, 0x39, 0x82, 0x9b, 0x2f, 0xff, 0x87, 0x34, 0x8e, 0x43, 0x44, 0xc4, 0xde, 0xe9, 0xcb,
		0x54, 0x7b, 0x94, 0x32, 0xa6, 0xc2, 0x23, 0x3d, 0xee, 0x4c, 0x95, 0x0b, 0x42, 0xfa, 0xc3, 0x4e,
		0x08, 0x2e, 0xa1, 0x66, 0x28, 0xd9, 0x24, 0xb2, 0x76, 0x5b, 0xa2, 0x49, 0x6d, 0x8b, 0xd1, 0x25,
		0x72, 0xf8, 0xf6, 0x64, 0x86, 0x68, 0x98, 0x16, 0xd4, 0xa4, 0x5c, 0xcc, 0x5d, 0x65, 0xb6, 0x92,
		0x6c, 0x70, 0x48, 0x50, 0xfd, 0xed, 0xb9, 0xda, 0x5e, 0x15, 0x46, 0x57, 0xa7, 0x8d, 0x9d, 0x84,
		0x90, 0xd8, 0xab, 0x00, 0x8c, 0xbc, 0xd3, 0x0a, 0xf7, 0xe4, 0x58, 0x05, 0xb8, 0xb3, 0x45, 0x06,
		0xd0, 0x2c, 0x1e, 0x8f, 0xca, 0x3f, 0x0f, 0x02, 0xc1, 0xaf, 0xbd, 0x03, 0x01, 0x13, 0x8a, 0x6b,
		0x3a, 0x91, 0x11, 0x41, 0x4f, 0x67, 0xdc, 0xea, 0x97, 0xf2, 0xcf, 0xce, 0xf0, 0xb4, 0xe6, 0x73,
		0x96, 0xac, 0x74, 0x22, 0xe7, 0xad, 0x35, 0x85, 0xe2, 0xf9, 0x37, 0xe8, 0x1c, 0x75, 0xdf, 0x6e,
		0x47, 0xf1, 0x1a, 0x71, 0x1d, 0x29, 0xc5, 0x89, 0x6f, 0xb7, 0x62, 0x0e, 0xaa, 0x18, 0xbe, 0x1b,
		0xfc, 0x56, 0x3e, 0x4b, 0xc6, 0xd2, 0x79, 0x20, 0x9a, 0xdb, 0xc0, 0xfe, 0x78, 0xcd, 0x5a, 0xf4,
		0x1f, 0xdd, 0xa8, 0x33, 0x88, 0x07, 0xc7, 0x31, 0xb1, 0x12, 0x10, 0x59, 0x27, 0x80, 0xec, 0x5f,
		0x60, 0x51, 0x7f, 0xa9, 0x19, 0xb5, 0x4a, 0x0d, 0x2d, 0xe5, 0x7a, 0x9f, 0x93, 0xc9, 0x9c, 0xef,
		0xa0, 0xe0, 0x3b, 0x4d, 0xae, 0x2a, 0xf5, 0xb0, 0xc8, 0xeb, 0xbb, 0x3c, 0x83, 0x53, 0x99, 0x61,
		0x17, 0x2b, 0x04, 0x7e, 0xba, 0x77, 0xd6, 0x26, 0xe1, 0x69, 0x14, 0x63, 0x55, 0x21, 0x0c, 0x7d}
	return GF256{box[B0.i]}
}

func mixColumn(C [4]GF256) [4]GF256 {
	return [4]GF256{
		(C[0].mult(GF256{2})).add(C[1].mult(GF256{3})).add(C[2].mult(GF256{1})).add(C[3].mult(GF256{1})),
		(C[0].mult(GF256{1})).add(C[1].mult(GF256{2})).add(C[2].mult(GF256{3})).add(C[3].mult(GF256{1})),
		(C[0].mult(GF256{1})).add(C[1].mult(GF256{1})).add(C[2].mult(GF256{2})).add(C[3].mult(GF256{3})),
		(C[0].mult(GF256{3})).add(C[1].mult(GF256{1})).add(C[2].mult(GF256{1})).add(C[3].mult(GF256{2}))}
}

func unMixColumn(C [4]GF256) [4]GF256 {
	return [4]GF256{
		(C[0].mult(GF256{14})).add(C[1].mult(GF256{11})).add(C[2].mult(GF256{13})).add(C[3].mult(GF256{9})),
		(C[0].mult(GF256{9})).add(C[1].mult(GF256{14})).add(C[2].mult(GF256{11})).add(C[3].mult(GF256{13})),
		(C[0].mult(GF256{13})).add(C[1].mult(GF256{9})).add(C[2].mult(GF256{14})).add(C[3].mult(GF256{11})),
		(C[0].mult(GF256{11})).add(C[1].mult(GF256{13})).add(C[2].mult(GF256{9})).add(C[3].mult(GF256{14}))}
}

func mixColumns(A [4][4]GF256) [4][4]GF256 {
	c0 := mixColumn([4]GF256{A[0][0], A[1][0], A[2][0], A[3][0]})
	c1 := mixColumn([4]GF256{A[0][1], A[1][1], A[2][1], A[3][1]})
	c2 := mixColumn([4]GF256{A[0][2], A[1][2], A[2][2], A[3][2]})
	c3 := mixColumn([4]GF256{A[0][3], A[1][3], A[2][3], A[3][3]})
	return [4][4]GF256{
		{c0[0], c1[0], c2[0], c3[0]},
		{c0[1], c1[1], c2[1], c3[1]},
		{c0[2], c1[2], c2[2], c3[2]},
		{c0[3], c1[3], c2[3], c3[3]}}
}

func unMixColumns(A [4][4]GF256) [4][4]GF256 {
	c0 := unMixColumn([4]GF256{A[0][0], A[1][0], A[2][0], A[3][0]})
	c1 := unMixColumn([4]GF256{A[0][1], A[1][1], A[2][1], A[3][1]})
	c2 := unMixColumn([4]GF256{A[0][2], A[1][2], A[2][2], A[3][2]})
	c3 := unMixColumn([4]GF256{A[0][3], A[1][3], A[2][3], A[3][3]})
	return [4][4]GF256{
		{c0[0], c1[0], c2[0], c3[0]},
		{c0[1], c1[1], c2[1], c3[1]},
		{c0[2], c1[2], c2[2], c3[2]},
		{c0[3], c1[3], c2[3], c3[3]}}
}

func shiftRows(A [4][4]GF256) [4][4]GF256 {
	return [4][4]GF256{
		{A[0][0], A[0][1], A[0][2], A[0][3]},
		{A[1][1], A[1][2], A[1][3], A[1][0]},
		{A[2][2], A[2][3], A[2][0], A[2][1]},
		{A[3][3], A[3][0], A[3][1], A[3][2]}}
}

func unShiftRows(A [4][4]GF256) [4][4]GF256 {
	return [4][4]GF256{
		{A[0][0], A[0][1], A[0][2], A[0][3]},
		{A[1][3], A[1][0], A[1][1], A[1][2]},
		{A[2][2], A[2][3], A[2][0], A[2][1]},
		{A[3][1], A[3][2], A[3][3], A[3][0]}}
}

func rotWord(i uint32) uint32 {
	return ((0xFF & i) << 8) ^ ((0xFF00 & i) << 8) ^ ((0xFF0000 & i) << 8) ^ (0xFF >> 24)
}

func tobytes(i uint32) [4]GF256 {
	return [4]GF256{GF256{uint8(0xFF & (i >> 24))}, GF256{uint8(0xFF & (i >> 16))}, GF256{uint8(0xFF & (i >> 8))}, GF256{uint8(0xFF & i)}}
}
func subWord(i uint32) uint32 {
	temp := tobytes(i)
	return uint32(temp[3].sbox().i) ^ uint32(temp[2].sbox().i)<<8 ^ uint32(temp[1].sbox().i)<<16 ^ uint32(temp[0].sbox().i)<<24
}

func keyschedule(key [8]uint32, Rin int) []uint32 {
	N := 8
	R := Rin + 2
	rcon := make([]uint32, R+1)
	rcon[1] = 1 << 24
	rc := GF256{1}
	for i := 2; i < R+1; i++ {
		rc = rc.mult(GF256{2})
		rcon[i] = uint32(rc.i) << 24
	}
	W := make([]uint32, (4*R)-1)
	for i := 0; i < (4*R)-1; i++ {
		if i < N {
			W[i] = key[i]
		} else if i%N == 0 {
			W[i] = W[i-N] ^ subWord(rotWord(W[i-1])) ^ rcon[i/N]
		} else if N > 6 && i%N == 4 {
			W[i] = W[i-N] ^ subWord(W[i-1])
		} else {
			W[i] = W[i-N] ^ W[i-1]
		}
	}
	return W
}

func addroundkey(S [4][4]GF256, W []uint32, R int) [4][4]GF256 {
	out := newGF256()
	for i := range out {
		keytemp := tobytes(W[4*R+i])
		for j := range out {
			out[i][j].i = S[i][j].add(keytemp[j]).i
		}
	}
	return out
}

func subBytes(S [4][4]GF256) [4][4]GF256 {
	out := newGF256()
	for i := range S {
		for j := range S {
			out[i][j] = S[i][j].sbox()
		}
	}
	return out
}

func unSubBytes(S [4][4]GF256) [4][4]GF256 {
	out := newGF256()
	for i := range S {
		for j := range S {
			out[i][j] = S[i][j].invsbox()
		}
	}
	return out
}

// AES256encrypt only encrypts
func AES256encrypt(key [8]uint32, message string) string {
	R := 13
	input := prepstring(message)
	W := keyschedule(key, R)
	var out [][4][4]GF256
	for _, s := range input {

		s = addroundkey(s, W, 0)
		for r := 1; r < R; r++ {
			s = subBytes(s)
			s = shiftRows(s)
			s = mixColumns(s)
			s = addroundkey(s, W, r)
		}
		s = subBytes(s)
		s = shiftRows(s)
		s = addroundkey(s, W, R)
		out = append(out, s)
	}
	return toString(out)
}

// AES256decrypt only decrypts
func AES256decrypt(key [8]uint32, message string) string {
	R := 13
	input := mesagesplit(message)
	W := keyschedule(key, R)
	var out [][4][4]GF256
	for _, s := range input {
		s = addroundkey(s, W, R)
		s = unShiftRows(s)
		s = unSubBytes(s)
		for r := R - 1; r > 0; r-- {
			s = addroundkey(s, W, r)
			s = unMixColumns(s)
			s = unShiftRows(s)
			s = unSubBytes(s)
		}
		s = addroundkey(s, W, 0)
		out = append(out, s)
	}
	return outstring(toString(out))
}

func prepstring(bytes string) [][4][4]GF256 {
	var random [4][4]GF256
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			random[i][j] = GF256{uint8(rand.Uint32())}
		}
	}
	length := len(bytes)
	if length%16 == 0 {
		length = length / 16
	} else {
		length = length/16 + 1
	}
	var out [][4][4]GF256
	out = append(out, random)
	for i := 0; i < length; i++ {
		temp := newGF256()
		offset := (i) * 16
		lim := 16
		if offset+16 > len(bytes) {
			lim = len(bytes) % 16
		}
		for j := 0; j < lim; j++ {
			temp[j/4][j%4] = GF256{bytes[offset+j]}
		}
		for i := 0; i < 4; i++ {
			for j := 0; j < 4; j++ {
				temp[i][j] = temp[i][j].add(random[i][j])
			}
		}
		random[0][0] = GF256{(random[0][0]).i + 1}
		for i := 0; random[i/4][i%4].i == 0; i++ {
			random[i/4][i%4] = GF256{random[(i+1)/4][(i+1)%4].i + 1}
			if i == 14 {
				break
			}
		}
		out = append(out, temp)
	}
	return out
}

func outstring(s string) string {
	bytes := []uint8(s[16:])
	offset := []uint8(s[:16])
	var out []uint8
	for i := 0; i < len(bytes)/16; i++ {
		for j := 0; j < 16; j++ {
			out = append(out, offset[j]^bytes[(i)*16+j])
		}
		offset[0] = offset[0] + 1
		for i := 0; offset[i] == 0; i++ {
			offset[i] = offset[i+1] + 1
			if i == 14 {
				break
			}
		}
	}
	return string(out)
}

func newGF256() [4][4]GF256 {
	return [4][4]GF256{[4]GF256{GF256{0}, GF256{0}, GF256{0}, GF256{0}}, [4]GF256{GF256{0}, GF256{0}, GF256{0}, GF256{0}}, [4]GF256{GF256{0}, GF256{0}, GF256{0}, GF256{0}}, [4]GF256{GF256{0}, GF256{0}, GF256{0}, GF256{0}}}
}

func mesagesplit(bytes string) [][4][4]GF256 {
	length := len(bytes)
	if length%16 == 0 {
		length = length / 16
	} else {
		length = length/16 + 1
	}
	var out [][4][4]GF256
	for i := 0; i < length; i++ {
		temp := newGF256()
		offset := (i) * 16
		lim := 16
		if offset+16 > len(bytes) {
			lim = len(bytes) % 16
		}
		for j := 0; j < lim; j++ {
			temp[j/4][j%4] = GF256{bytes[offset+j]}
		}
		out = append(out, temp)
	}
	return out
}

func toString(bytes [][4][4]GF256) string {
	out := make([]uint8, len(bytes)*16)
	for k := range out {
		out[k] = bytes[k/16][(k/4)%4][k%4].i
	}
	return string(out)
}

//AES256 is the main de/encryption function
// options is for choosing en or decrypt and in future possibly multicore support
func AES256(message string, key interface{}, options interface{}) string {
	var K [8]uint32
	switch T := key.(type) {
	case [8]uint32:
		K = T
	case [32]uint8:
		for i := 0; i < 8; i++ {
			K[i] = (uint32(T[(4*i)]) << 24) ^ (uint32(T[(4*i)+0]) << 16) ^ (uint32(T[(4*i)+0]) << 8) ^ uint32(T[(4*i)+0])
		}
	case string:
		for i := 0; i < 32; i++ {
			if len(T) < 64 {
				T = "  " + T
			} else {
				break
			}
		}
		if len(T) == 66 && string(T[0:1]) == "0x" {
			Temp, _ := hex.DecodeString(T[2:])
			T = string(Temp)
		}
		for i := 0; i < 8; i++ {
			K[i] = (uint32(T[(4*i)]) << 24) ^ (uint32(T[(4*i)+0]) << 16) ^ (uint32(T[(4*i)+0]) << 8) ^ uint32(T[(4*i)+0])
		}
	}
	var out string
	switch O := options.(type) {
	case string:
		if O == "decrypt" {
			out = AES256decrypt(K, message)
		} else {
			out = AES256encrypt(K, message)
		}
	case int:
		if O >= 1 {
			out = AES256decrypt(K, message)
		} else {
			out = AES256encrypt(K, message)
		}
	}
	return out

}
