package main

import (
	"fmt"
	"math/big"
	"math/rand"
)

// used for message encryption with different algortihms
// AES256key implements this but other encryption methods would be added here
type key interface {
	encrypt(message, interface{}) message
	decrypt(message, interface{}) message
}

//AES256key is default key for AES encryption
type AES256key struct {
	value [8]uint32
}

func (K AES256key) encrypt(M message, options interface{}) message {
	var out message
	switch m := M.(type) {
	case textmessage:
		if m.encrypted {
			out = textmessage{text: AES256(m.text, K.value, 0), bytecount: m.bytecount, encrypted: true}
		} else {
			out = M
		}
	default:
		out = M
	}
	return out
}

func (K AES256key) decrypt(M message, options interface{}) message {
	var out message
	switch m := M.(type) {
	case textmessage:
		if m.encrypted {
			out = textmessage{text: AES256(m.text, K.value, 0), bytecount: m.bytecount, encrypted: false}
		} else {
			out = M
		}
	default:
		out = M
	}
	return out
}

type publicKey interface {
	mergewithpublic(publicKey) (key, error)
	clone() (publicKey, publicKey)
}

// mod  stores private and public keys for Diffie Hellman keyexchange
type mod struct {
	value, mod, root big.Int
}

func (priv mod) mergewithpublic(pub publicKey) (key, error) {
	switch p := pub.(type) {
	case mod:
		var shared big.Int
		_ = shared.Exp(&p.value, &priv.value, &priv.mod)
		return AES256key{SHA256(string(shared.Bytes()))}, nil
	default:
		return *new(AES256key), &MyError{"cant merge with this keytype"}
	}
}

func (priv mod) clone() (publicKey, publicKey) {
	var privkey, pubkey big.Int
	random := randomInt(2056)
	_, _ = random.DivMod(&random, &priv.mod, &privkey)
	_ = pubkey.Exp(&priv.root, &privkey, &priv.mod)
	return mod{privkey, priv.mod, priv.root}, mod{pubkey, priv.mod, priv.root}
}

// generate public private key pair
func newPair(options interface{}) (publicKey, publicKey) {
	switch O := options.(type) {
	case publicKey:
		return O.clone()
	default:
		// curve from https://neuromancer.sk/std/x962/c2tnb431r1#
		a, b, k, x, y := new(big.Int), new(big.Int), new(big.Int), new(big.Int), new(big.Int)
		a, _ = a.SetString("0x1a827ef00dd6fc0e234caf046c6a5d8a85395b236cc4ad2cf32a0cadbdc9ddf620b0eb9906d0957f6c6feacd615468df104de296cd8f", 0)
		b, _ = b.SetString("0x10d9b4a3d9047d8b154359abfb1b7f5485b04ceb868237ddc9deda982a679a5a919b626d4e50a8dd731b107a9962381fb5d807bf2618", 0)
		k, _ = k.SetString("0x0340340340340340340340340340340340340340340340340340340323c313fab50589703b5ec68d3587fec60d161cc149c1ad4a91", 0)
		x, _ = x.SetString("0x120fc05d3c67a99de161d2f4092622feca701be4f50f4758714e8a87bbf2a658ef8c21e7c5efe965361f6c2999c0c247b0dbd70ce6b7", 0)
		y, _ = y.SetString("0x20d0af8903a96f8d5fa2c255745d3c451b302c9346d9b7e485e7bce41f6b591f3e8f6addcbb0bc4c2f947a7de1a89b625d6a598b3760", 0)
		key := ECCkey{curve: elipticcurve{a: *a, b: *b, k: *k, root: n2{x: *x, y: *y}}}
		return key.clone()
	}
}

func randomInt(n int) big.Int {
	var random big.Int
	for i := 0; i < n/64+1; i++ {
		rtemp1 := rand.Uint64()
		rtemp2 := *big.NewInt(int64(rtemp1))
		if rtemp1&(1<<63) == 1 {
			_ = rtemp2.Lsh(&rtemp2, 1)
		}
		_ = random.Lsh(&random, 64)
		_ = random.Add(&random, &rtemp2)
	}
	_ = random.Rsh(&random, uint(64-n%64))
	return random
}

// generate random primes for the key exchange
func bigPrimes(size int) big.Int {
	seed := randomInt(size)
	seed.SetBit(&seed, 0, 1)
	bigTwo := big.NewInt(int64(2))
	for {
		if seed.ProbablyPrime(5) {
			if seed.ProbablyPrime(32) {
				if seed.ProbablyPrime(128) {
					break
				}
			}
		}
		_ = seed.Add(&seed, bigTwo)
	}
	return seed
}

func egcd(a *big.Int, b *big.Int) (*big.Int, *big.Int, *big.Int) {
	if a.BitLen() == 0 {
		return b, big.NewInt(0), big.NewInt(1)
	}
	z, m := new(big.Int), new(big.Int)
	_, _ = z.DivMod(b, a, m) // b=a*z+m
	gcd, x, y := egcd(m, a)
	return gcd, y.Sub(y, z.Mul(z, x)), x
}

func invert(a *big.Int, modu *big.Int) *big.Int {
	_, inv, _ := egcd(a, modu)
	return inv.Mod(inv, modu)
}

func sqrt(a *big.Int, modu *big.Int) (*big.Int, error) {
	if a == nil || modu == nil {
		return nil, &MyError{"recieving nil pointer"}
	}
	pm1d2 := new(big.Int)
	pm1d2 = pm1d2.Sub(modu, big.NewInt(1)).Rsh(pm1d2, 1)
	pow := func(n *big.Int) bool {
		out := new(big.Int)
		out = out.Exp(n, pm1d2, modu)
		return 1 == out.BitLen()
	}
	if !pow(a) {
		return nil, &MyError{"Has no sqrt"}
	}
	z0, Q := new(big.Int), new(big.Int)
	S := z0.Sub(modu, big.NewInt(1)).TrailingZeroBits()
	Q = Q.Rsh(modu, S)
	zint := int64(2)
	for ; pow(big.NewInt(zint)); zint++ {
	}
	z := big.NewInt(zint)
	acopy0, Qp1d2 := new(big.Int), new(big.Int)
	*acopy0, *Qp1d2 = *a, *Q
	M, c, t, R := big.NewInt(int64(S)), z.Exp(z, Q, modu), acopy0.Exp(a, Q, modu), Qp1d2.Exp(a, Qp1d2.Add(Qp1d2, big.NewInt(1)).Rsh(Qp1d2, 1), modu)
	for {
		switch t.BitLen() {
		case 0:
			return big.NewInt(0), nil
		case 1:
			return R, nil
		default:
			temp := new(big.Int)
			temp.Exp(t, big.NewInt(2), modu)
			i := 1
			for {
				if temp.BitLen() == 1 {
					break
				}
				temp.Exp(temp, big.NewInt(2), modu)
				i = i + 1
			}
			bigi := big.NewInt(int64(i))
			b := new(big.Int)
			b.Exp(c, b.Exp(big.NewInt(2), M.Sub(M, bigi).Sub(M, big.NewInt(1)), nil), modu)
			bsq := new(big.Int)
			bsq.Mul(b, b).Mod(bsq, modu)
			M, R, c, t = bigi, R.Mul(R, b).Mod(R, modu), bsq, t.Mul(t, bsq).Mod(t, modu)
		}
	}
}

type n2 struct {
	x, y big.Int
	o    bool
}

type elipticcurve struct {
	a, b, k big.Int
	root    n2
}

// ECCkey implements publicKey
type ECCkey struct {
	value interface{}
	curve elipticcurve
}

func (K ECCkey) clone() (publicKey, publicKey) {
	rand := randomInt(K.curve.k.BitLen() + 8)
	priv := rand.Mod(&rand, &K.curve.k)
	pub := K.curve.Mul(K.curve.root, priv)
	return ECCkey{value: *priv, curve: K.curve}, ECCkey{value: pub, curve: K.curve}
}

func (K ECCkey) mergewithpublic(pub publicKey) (out key, err error) {
	recoverfunc := func() {
		if r := recover(); r != nil {
			err = &MyError{fmt.Sprint(r)}
		}
	}
	defer recoverfunc()
	switch p := pub.(type) {
	case ECCkey:
		temp := K.value.(big.Int)
		pcom := K.curve.Mul((p.value).(n2), &temp)
		shared := new(big.Int)
		shared = shared.Add(&pcom.x, &pcom.y)
		return AES256key{SHA256(string(shared.Bytes()))}, nil
	default:
		return *new(AES256key), &MyError{"cant merge with this keytype"}
	}
}

func (C *elipticcurve) valid() bool {
	z0, z1 := big.NewInt(4), big.NewInt(27)
	return 0 == z0.Mul(z0, &C.a).Mul(z0, &C.a).Mul(z0, &C.a).Add(z0, z1.Mul(z1, &C.b).Mul(z1, &C.b)).Mod(z0, &C.k).Sign()
}

func (C *elipticcurve) getPoint(x *big.Int) (n2, error) {
	z0, z1 := new(big.Int), new(big.Int)
	*z0 = *x
	z0.Mul(z0, x).Mul(z0, x).Mul(z0, x).Add(z0, z1.Mul(&C.a, x)).Add(z0, &C.b).Mod(z0, &C.k)
	z0, err := sqrt(z0, &C.k)
	if err != nil {
		return n2{o: true}, &MyError{"no y value for this x"}
	}
	return n2{*x, *z0, false}, nil

}

func (C *elipticcurve) isonCurve(p n2) bool {
	z0, z1, z2 := new(big.Int), new(big.Int), new(big.Int)
	*z0, *z1, *z2 = p.x, C.a, p.y
	return 0 == z0.Mul(z0, &p.x).Mul(z0, &p.x).Mul(z0, &p.x).Add(z0, z1.Mul(z1, &p.x)).Add(z0, &C.b).Sub(z0, z2.Mul(z2, &p.y)).Mod(z0, &C.k).Sign()
}

func (C *elipticcurve) Add(p0 n2, p1 n2) n2 {
	if p0.o == true {
		return p1
	}
	if p1.o == true {
		return p0
	}
	if (p0.x).Cmp(&(p1.x)) == 1 {
		if (p0.y).Cmp(&(p1.y)) == 1 {
			s, inv := new(big.Int), new(big.Int)
			inv = invert(inv.Mul(&(p0.y), big.NewInt(2)), &C.k)
			s.Mul(big.NewInt(3), &(p0.x)).Mul(s, &(p0.x)).Mul(s, &(p0.x)).Add(s, &(C.a))
			xr, yr, temp := new(big.Int), new(big.Int), new(big.Int)
			xr.Mul(s, s).Sub(xr, temp.Mul(big.NewInt(2), &(p0.x)))
			yr.Sub(&(p1.x), &(p0.x)).Mul(yr, s).Add(yr, &(p0.y))
			return n2{*xr, *yr, false}
		}
		return n2{o: true}
	}
	s, inv := new(big.Int), new(big.Int)
	s = s.Sub(&(p0.y), &(p1.y)).Mul(s, invert(inv.Sub(&(p0.x), &(p1.x)), &(C.k)))
	xr, yr := new(big.Int), new(big.Int)
	xr.Mul(s, s).Sub(xr, &(p0.x)).Sub(xr, &(p1.x))
	yr.Sub(&(p1.x), &(p0.x)).Mul(yr, s).Add(yr, &(p0.y))
	return p0
}

func (C *elipticcurve) Mul(p n2, n *big.Int) n2 {
	length := n.BitLen()
	if length == 0 {
		return n2{o: true}
	}
	res := p
	for i := length - 1; i >= 0; i-- {
		res = C.Add(res, res)
		if n.Bit(i) == 1 {
			res = C.Add(res, p)
		}
	}
	return res
}
