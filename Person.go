package main

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// this is the main user instance
// all user accsesible functions are memberfunctions of person
type person struct {
	friends     map[string]friend
	iptofriend  map[string]map[string]string
	iptofriend2 map[string]string
	chatlog     map[string]string
	log         string
	Networks    map[string]Networkstore
	cookie      http.Cookie
	server      *http.Server
	servererror error
	password    string
	client      *http.Client
}

// stores all information tecnical information about befriended persons
type friend struct {
	name    string
	Network string
	IP      string
	status  uint32 // 0 done, 1 sending, 2 recieving, 3 blocked, 4 ignored 5 invalid
	keypriv publicKey
	keycom  key
	unread  bool
}

func newPerson() person {
	friends := make(map[string]friend)
	iptofriend := make(map[string]map[string]string)
	iptofriend2 := make(map[string]string)
	chatlog := make(map[string]string)
	log := ""
	Networks := make(map[string]Networkstore)
	value := randomInt(256)
	valuestr := (&value).Text(62)
	cookie := http.Cookie{
		Name:  "validate",
		Value: valuestr,
		Path:  "/",
	}
	server := &http.Server{Addr: ":8080", Handler: nil, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, MaxHeaderBytes: 1 << 20}
	client := &http.Client{}
	out := person{
		friends:     friends,
		iptofriend:  iptofriend,
		iptofriend2: iptofriend2,
		chatlog:     chatlog,
		log:         log,
		Networks:    Networks,
		cookie:      cookie,
		server:      server,
		servererror: nil,
		password:    "",
		client:      client}
	return out
}

func (P *person) saveFriends() {
	if P.password == "" {
		return
	}
	frlist := []friend{}
	for _, fr := range P.friends {
		frlist = append(frlist, fr)
	}
	savestr := printObject(frlist)
	key := SHA256(P.password + "key")
	savestr = AES256(savestr, key, 0)
	err := store("saves/friends.txt", savestr)
	if err != nil {
		P.Log(err.Error())
	}
}

func (P *person) saveChats() {
	if P.password == "" {
		return
	}
	chatstr := ""
	for ident, fr := range P.chatlog {
		chatstr += "<<<>>>" + ident + ">>><<<" + fr
	}
	key := SHA256(P.password + "key")
	chatstr = AES256(chatstr, key, 0)
	err := store("saves/chat.txt", chatstr)
	if err != nil {
		P.Log(err.Error())
	}
	logstr := AES256(P.log, key, 0)
	err = store("saves/log.txt", logstr)
	if err != nil {
		P.Log(err.Error())
	}
}

func (P *person) loadChats() {
	if P.password == "" {
		return
	}
	chatstr, err := load("saves/chat.txt")
	if err != nil {
		P.Log(err.Error())
		return
	}
	key := SHA256(P.password + "key")
	chatstr = AES256(chatstr, key, 1)
	chats := strings.Split(chatstr, "<<<>>>")
	for _, chat := range chats {
		if chat == "" {
			continue
		}
		chatarr := strings.Split(chat, ">>><<<")
		P.chatlog[chatarr[0]] = chatarr[1]
	}
	logstr, err := load("saves/log.txt")
	if err != nil {
		P.Log(err.Error())
		return
	}
	P.log = AES256(logstr, key, 1)
}

func (P *person) Shutdown() {
	P.saveFriends()
	P.saveChats()
	P.servererror = P.server.Close()
}

func (P *person) fromFile() {
	if P.password == "" {
		return
	}
	savestr, err := load("saves/friends.txt")
	if err != nil {
		P.Log(err.Error())
		fmt.Println(err.Error())
		return
	}
	key := SHA256(P.password + "key")
	savestr = AES256(savestr, key, 1)
	friendslice, err := read(savestr)
	if err != nil {
		P.Log(err.Error())
		fmt.Println(err.Error())
		return
	}
	for _, fr := range friendslice.([]friend) {
		P.friends[fr.name] = fr
		P.iptofriend2[fr.IP] = fr.name
		_, ok := P.friends[fr.Network]
		if ok {
			P.iptofriend[fr.Network][fr.IP] = fr.name
		} else {
			temp := make(map[string]string)
			P.iptofriend[fr.Network] = temp
			P.iptofriend[fr.Network][fr.IP] = fr.name
		}
	}
}

func (P *person) Log(m string) {
	P.log += (time.Now().Format(time.RFC850) + " : " + m)
}

func (P *person) ChatLog(name string, sender string, message string) {
	P.chatlog[name] = P.chatlog[name] + ("[" + sender + "]:" + message + "\n")
	temp := P.friends[name]
	temp.unread = false
	P.friends[name] = temp
}

func (P *person) chat(ident string, M string) {
	_, ok := P.friends[ident]
	if ok {
		P.msg(ident, M)
		P.ChatLog(ident, "me", M)
	} else {
		P.Log("invalid chat to : " + ident)
	}
}

func (P *person) msg(ident string, M interface{}) {
	F, ok := P.friends[ident]
	if ok {
		net := P.Networks[F.Network]
		net.send(F.IP, F.keycom.encrypt(newMessage(M), 0))
	} else {
		P.Log("invalid msg to : " + ident)
	}
}

func (P *person) leaveNetwork(ident string) {
	N, ok := P.Networks[ident]
	if ok {
		N.leave()
		delete(P.Networks, ident)
	} else {
		P.Log("failed to leave Network : " + ident)
	}
}

func (P *person) AddNetwork(ident string, joiner func() Networkstore) {
	_, ok := P.Networks[ident]
	if ok {
		P.Log("already joined Network : " + ident)
	} else {
		N := joiner()
		P.Networks[ident] = N
		P.iptofriend[ident] = make(map[string]string)
	}
}

//MyError is simple custum Error
type MyError struct {
	What string
}

func (e *MyError) Error() string {
	return (*e).What
}

func (P *person) RenameFriend(old string, new string) error {
	var validnames = regexp.MustCompile("^([a-zA-Z0-9:]+)$")
	valid := validnames.MatchString(new)
	if valid {
		F, knowno := P.friends[old]
		if knowno {
			_, knownn := P.friends[new]
			if !knownn {
				fmt.Println(F)
				F.name = new
				P.friends[new] = F
				P.iptofriend[F.Network][F.IP] = new
				P.iptofriend2[F.IP] = new
				P.chatlog[new] = P.chatlog[old]
				P.ChatLog(new, "system", "renamed "+old+" to "+new)
				delete(P.chatlog, old)
				delete(P.friends, old)
				P.saveFriends()
				return nil
			}
			P.Log("name to rename to already exist: " + new)
			return &MyError{"name to rename to already exist: " + new}
		}
		P.Log("name to rename does not exist: " + old)
		return &MyError{"name to rename does not exist: " + old}
	}
	P.Log("rename has invalid format: " + new)
	return &MyError{"rename has invalid format: " + new}
}

func (P *person) SendFriendrequest(network string, IP string, name string, options interface{}) {
	_, known := P.friends[name]
	if !known {
		namecolision, knownadr := P.iptofriend[network][IP]
		if !knownadr {
			priv, pub := newPair(options)
			P.friends[name] = friend{name, network, IP, 1, priv, AES256key{[8]uint32{0, 0, 0, 0, 0, 0, 0, 0}}, false}
			// P.iptofriend[network][IP] = name
			P.ChatLog(name, "system", "sending friendrequest")
			P.msg(name, newMessage(pub))
			P.saveFriends()
		} else {
			P.Log("connection already established under: " + namecolision)
		}

	} else {
		P.Log("name is already in use: " + name)
	}

}

func (P *person) recieve() {
	for net, N := range P.Networks {
		Msgs := N.recieve()
		for sender, Msgl := range Msgs {
			name, known := P.iptofriend[net][sender]
			if !known {
				namenew := net + sender
				switch m0 := Msgl[0].(type) {
				case keymessage:
					priv, pub := m0.pubkey.clone()
					comm, err := priv.mergewithpublic(m0.pubkey)
					if err != nil {
						fmt.Println(err)
						P.ChatLog(namenew, "system", "recieved key of invalid type")
						continue
					}
					P.friends[namenew] = friend{namenew, net, sender, 2, priv, comm, true}
					P.iptofriend[net][sender] = namenew
					P.msg(namenew, newMessage(pub))
					P.ChatLog(namenew, "system", "comonkey established from frendrequest")
				case textmessage:
					priv, pub := newPair(0)
					P.friends[namenew] = friend{namenew, net, sender, 5, priv, AES256key{[8]uint32{0, 0, 0, 0, 0, 0, 0, 0}}, true}
					P.iptofriend[net][sender] = namenew
					P.ChatLog(namenew, "system", "recieving unencrpted message:")
					P.ChatLog(namenew, namenew, m0.toString())
					P.msg(namenew, newMessage(pub))
				default:
					P.Log("recieved invalid message")
					Msgl = Msgl[1:]
				}
			} else if P.friends[name].status == 1 || P.friends[name].status == 5 {
				switch m0 := Msgl[0].(type) {
				case keymessage:
					f := P.friends[name]
					comkey, err := f.keypriv.mergewithpublic(m0.pubkey)
					if err != nil {
						P.ChatLog(name, "system", "recieving invalid key to merge")
						continue
					}
					P.friends[name] = friend{f.name, f.Network, f.IP, 0, f.keypriv, comkey, true}
					P.ChatLog(name, "system", "comonkey established")
				case textmessage:
					P.ChatLog(name, "system", "recieving unencrpted message:")
					P.ChatLog(name, name, m0.toString())
				default:
					P.Log("recieved invalid message")
				}
				Msgl = Msgl[1:]
			}
			name, _ = P.iptofriend[net][sender]
			for _, m := range Msgl {
				m := P.friends[name].keycom.decrypt(m, 0)
				switch m0 := m.(type) {
				case keymessage:
					P.ChatLog(name, "system", "recieving new public key")
				case textmessage:
					P.ChatLog(name, name, m0.toString())
				default:
					P.Log("recieved invalid message")
				}
			}
		}
	}
}

func (P *person) Showchat(ident string, options interface{}) string {
	_, ok := P.friends[ident]
	var out string
	if ok {
		temp := P.friends[ident]
		temp.unread = false
		P.friends[ident] = temp
		chat := strings.Split(P.chatlog[ident], "\n")
		switch op := options.(type) {
		case [2]int:
			start := op[0]
			if op[0] > 0 {
				if op[0] >= len(chat) {
					start = len(chat) - 1
				}
			} else { // op is negative here
				if op[0] >= -len(chat) {
					start = len(chat) + op[0]
				} else {
					start = 0
				}
			}
			end := op[1]
			if op[1] > 0 {
				if op[0] >= len(chat) {
					start = len(chat) - 1
				}
			} else { // op is negative here
				if op[0] >= -len(chat) {
					start = len(chat) + op[0]
				} else {
					start = 0
				}
			}
			if end < start {
				end = start
			}
			for _, message := range chat[start:end] {
				out = out + "\n" + message
			}
		default:
			for _, message := range chat {
				out = out + "\n" + message
			}
		}
	} else {
		P.Log("invalid request to chat with: " + ident)
	}
	return out
}

func (P *person) RecieveMsg(M message, sender string) {
	ident, known := P.iptofriend2[sender]
	// if known {
	// 	known = known &&  P.friends[ident].status<=2
	// }
	if known {
		fr := P.friends[ident]
		switch m := M.(type) {
		case textmessage:
			if P.friends[ident].status <= 2 {
				Mnew := fr.keycom.decrypt(M, 0)
				P.ChatLog(ident, ident, Mnew.toString())
			} else {
				P.ChatLog(ident, "system", "recieving unencrpted message:")
				P.ChatLog(ident, ident, m.toString())
			}
		case keymessage:
			if fr.status == 1 {
				comkey, err := fr.keypriv.mergewithpublic(m.pubkey)
				if err != nil {
					P.ChatLog(ident, "system", "recieving unmatching key type ")
					return
				}
				fr.keycom = comkey
				fr.status = 0
				P.ChatLog(ident, "system", "established key ")
			} else {
				P.ChatLog(ident, "system", "recieved another key message")
			}
		default:
			P.Log("recieved invalid message type")
		}
		fr.unread = true
		P.friends[ident] = fr
		P.saveFriends()
	} else {
		namenew := "newFriend"
		for {
			_, valid := P.friends[namenew]
			if valid {
				namenew += "0"
			} else {
				break
			}
		}
		fr := friend{name: namenew, Network: "LAN", IP: sender, status: 5, unread: true}
		P.chatlog[namenew] = ""
		P.iptofriend2[sender] = namenew
		P.friends[namenew] = fr
		//P.iptofriend["LAN"][sender]=namenew
		switch m := M.(type) {
		case textmessage:
			P.ChatLog(namenew, "system", "recieving unencrpted message:")
			P.ChatLog(namenew, namenew, m.toString())
		case keymessage:
			priv, pub := m.pubkey.clone()
			fr.keypriv = priv
			comkey, err := fr.keypriv.mergewithpublic(m.pubkey)
			if err != nil {
				P.ChatLog(ident, "system", "recieving unmatching key type ")
				return
			}
			fr.keycom = comkey
			fr.status = 2
			P.msg(namenew, pub)
			P.Log("established new connection with" + namenew)
		default:
			P.Log("recieved invalid message type")
		}
		P.iptofriend2[sender] = namenew
		P.friends[namenew] = fr
		P.saveFriends()
	}
}

func (P *person) MsgLAN(ident string, M message) {
	fr, known := P.friends[ident]
	if known {
		m := M
		if fr.status == 2 || fr.status == 0 {
			fr.keycom.encrypt(m, 0)
		}
		form := url.Values{}
		form.Add("IP", ":8080")
		switch m := m.(type) {
		case textmessage:
			m.toString()
			form.Add("Text", m.toString())
		case keymessage:
			switch key := m.pubkey.(type) {
			case mod:
				form.Add("Value", key.value.Text(0))
				form.Add("Root", key.root.Text(0))
				form.Add("Modu", key.mod.Text(0))
			case ECCkey:
				point := key.value.(n2)
				form.Add("X", (&point).x.Text(0))
				form.Add("Y", (&point).y.Text(0))
				form.Add("A", key.curve.a.Text(0))
				form.Add("B", key.curve.b.Text(0))
				form.Add("K", key.curve.k.Text(0))
				form.Add("RootX", key.curve.root.x.Text(0))
				form.Add("RootY", key.curve.root.y.Text(0))
			}
		default:
			P.Log("sending invalid message type")
		}
		urlf, err := url.Parse(fr.IP)
		if err != nil {
			fmt.Println("\n", err.Error())
			P.Log(err.Error())
			return
		}
		req, err := http.NewRequest("POST", urlf.String(), strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println("\n", err.Error())
			P.Log(err.Error())
			return
		}
		req.PostForm = form
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		resp, err := P.client.Do(req)
		defer resp.Body.Close()
		fmt.Println("\n resp: ", resp)

	} else {
		P.Log("unknown Friend to send to:" + ident)
	}
}
