package main

import(
  "math/big"
  "math/rand"
)




type key interface{
  encrypt(message,interface{}) message
  decrypt(message,interface{}) message
}


type AES256key struct{
  value [8]uint32
}

func (K AES256key)encrypt(M message, options interface{}) message  {
  var out message
  switch m:=M.(type) {
  case textmessage:
    out=Message( struct{m string ; length int}{AES256(m.text,K.value,0),m.len()})
  default:
    out=M
  }
  return out
}

func (K AES256key)decrypt(M message, options interface{}) message  {
  var out message
  switch m:=M.(type) {
  case textmessage:
    out=Message( struct{m string ; length int}{AES256(m.text,K.value,1),m.len()})
  default:
    out=M
  }
  return out
}





type mod struct{
  value,mod,root big.Int
}

func (priv mod)merge_with_public(pub mod) key {
  var shared big.Int
  _=shared.Exp(&pub.value,&priv.value,&priv.mod)
  return AES256key{SHA256(string(shared.Bytes()))}
}

func (priv mod)clone() (mod,mod) {
  return New_pair([2]big.Int{priv.mod,priv.root})
}

func New_pair(options interface{}) (mod,mod) {
  var modu,root big.Int
  switch  O:=options.(type) {
  case []int:
    modu=*big.NewInt(int64(O[0]))
    root=*big.NewInt(int64(O[1]))
  case [2]int:
    modu=*big.NewInt(int64(O[0]))
    root=*big.NewInt(int64(O[1]))
  case []big.Int:
    modu=O[0]
    root=O[1]
  case [2]big.Int:
    modu=O[0]
    root=O[1]
  default:
    modu=*big.NewInt(int64(23))
    root=*big.NewInt(int64(5))
  }
  var random,privkey,pubkey big.Int
  for i := 0; i < modu.BitLen()/64+1; i++ {
    rtemp1:=rand.Uint64()
    rtemp2:=*big.NewInt(int64(rtemp1))
    if rtemp1&(1<<63)==1 {
      _=rtemp2.Lsh(&rtemp2,1)
    }
    _=random.Lsh(&random,64)
    _=random.Add(&random,&rtemp2)
  }
  _,_=random.DivMod(&random,&modu,&privkey)
  _=pubkey.Exp(&root,&privkey,&modu)
  return mod{privkey,modu,root}, mod{pubkey,modu,root}
}
