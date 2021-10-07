package main

// message interfaces are send to networking
// can be en/decrypted with key.encrypt(message)
// validate is reserved for future temper checking
type message interface {
	len() int
	toString() string
	validate() bool
}

func newMessage(M interface{}) message {
	var out message
	switch m := M.(type) {
	case string:
		out = textmessage{len(m), m, false}
	case struct {
		m      string
		length int
	}:
		out = textmessage{m.length, m.m, false}
	case message:
		out = m
	case publicKey:
		out = keymessage{m}
	}
	return out
}

type textmessage struct {
	bytecount int
	text      string
	encrypted bool
}

func (T textmessage) validate() bool {
	return true
}

func (T textmessage) len() int {
	return T.bytecount
}

func (T textmessage) toString() string {
	if T.encrypted {
		return T.text[:T.bytecount+16]
	}
	return T.text[:T.bytecount]
}

type keymessage struct {
	pubkey publicKey
}

func (T keymessage) validate() bool {
	return true
}

func (T keymessage) len() int {
	return 0
}

func (T keymessage) toString() string {
	return "pubkeyobject"
}
