package main

import(
)


// message interfaces are send to networking
// can be en/decrypted with key.encrypt(message)
// validate is reserved for future temper checking
type message interface{
  len() int
  to_string() string
  validate() bool
}

func Message(M interface{}) message{
  var out message
  switch m:=M.(type) {
  case string:
    out=textmessage{len(m),m}
  case struct{m string ; length int}:
    out=textmessage{m.length,m.m}
  case mod:
    out=keymessage{m}
  case message:
    out=m
  }
  return out
}

type textmessage struct{
  bytecount int
  text string
}

func (T textmessage)validate() bool  {
  return true
}

func (T textmessage)len() int  {
  return T.bytecount
}

func (T textmessage)to_string() string  {
  return T.text[:T.bytecount]
}




type keymessage struct{
  pubkey mod
}

func (T keymessage)validate() bool  {
  return true
}

func (T keymessage)len() int  {
  return 0
}

func (T keymessage)to_string() string  {
  return "pubkeyobject"
}
