package main

import (
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
			out = newMessage(struct {
				m      string
				length int
			}{AES256(m.text, K.value, 0), m.len()})
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
			out = newMessage(struct {
				m      string
				length int
			}{AES256(m.text, K.value, 1), m.len()})
		} else {
			out = M
		}
	default:
		out = M
	}
	return out
}

// mod  stores private and public keys for Diffie Hellman keyexchange
type mod struct {
	value, mod, root big.Int
}

func (priv mod) mergewithpublic(pub mod) key {
	var shared big.Int
	_ = shared.Exp(&pub.value, &priv.value, &priv.mod)
	return AES256key{SHA256(string(shared.Bytes()))}
}

func (priv mod) clone() (mod, mod) {
	return newPair([2]big.Int{priv.mod, priv.root})
}

// generate public private key pair
func newPair(options interface{}) (mod, mod) {
	var modu, root big.Int
	switch O := options.(type) {
	case [2]int:
		modu = *big.NewInt(int64(O[0]))
		root = *big.NewInt(int64(O[1]))
	case [2]big.Int:
		modu = O[0]
		root = O[1]
	case int:
		modu = *big.NewInt(int64(23))
		root = *big.NewInt(int64(5))
	default:
		modu = bigPrimes(2048)
		root = bigPrimes(2040)
	}
	var privkey, pubkey big.Int
	random := randomInt(2056)
	_, _ = random.DivMod(&random, &modu, &privkey)
	_ = pubkey.Exp(&root, &privkey, &modu)
	return mod{privkey, modu, root}, mod{pubkey, modu, root}
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
