package main

import(
	"time"
	"regexp"
	"net/http"
)



// this is the main user instance
// all user accsesible functions are memberfunctions of person
type person struct{
  friends map[string]friend
  iptofriend map[string]map[string]string
	iptofriend2 map[string]string
  chatlog map[string][]string
  log []string
  Networks map[string]Networkstore
	cookie http.Cookie
	server http.Server
	servererror error

}

// stores all information tecnical information about befriended persons
type friend struct{
  name string
  Network string
  IP string
  status uint32 // 0 done, 1 sending, 2 recieving, 3 blocked, 4 ignored 5 invalid
  keypriv mod
  keycom key
	unread bool
}


func NewPerson() person  {
  friends:=make(map[string]friend)
  iptofriend:=make(map[string]map[string]string)
	iptofriend2:=make(map[string]string)
  chatlog:=make(map[string][]string)
  var log []string
  Networks:=make(map[string]Networkstore)
	value:=Random_Int(256)
	valuestr:=(&value).String()
	cookie := http.Cookie{
					Name:   "validate",
					Value:  valuestr,
	        Path: "/",
			}
	server:=http.Server{Addr:":8080",Handler:nil}
  out:=person{friends:friends,
		iptofriend:iptofriend,
		iptofriend2:iptofriend2,
		chatlog:chatlog,log:log,
		Networks:Networks,
		cookie:cookie,
		server:server,
		servererror:nil}
	return out
}


func (P person)Log(m string)  {
  P.log=append(P.log,time.Now().Format(time.RFC850)+" : "+m)
}

func (P person)Chat_log(name string, sender string, message string)  {
  P.chatlog[name]=append(P.chatlog[name],"["+sender+"]:"+message)
	temp:=P.friends[name]
	temp.unread=false
	P.friends[name]=temp
}

func (P person)chat(ident string, M string)  {
  _,ok:=P.friends[ident]
  if ok {
    P.msg(ident,M)
    P.Chat_log(ident,"me",M)
  } else {
    P.Log("invalid chat to : "+ident)
  }
}

func (P person)msg(ident string, M interface{})  {
  F,ok:=P.friends[ident]
  if ok {
    net:=P.Networks[F.Network]
    net.send(F.IP,F.keycom.encrypt(Message(M),0))
  } else {
    P.Log("invalid msg to : "+ident)
  }
}

func (P person)leaveNetwork(ident string)  {
  N,ok:=P.Networks[ident]
  if ok {
    N.leave()
    delete(P.Networks,ident)
  } else {
    P.Log("failed to leave Network : "+ident)
  }
}

func (P person)Add_Network(ident string,joiner func()Networkstore )  {
  _,ok:=P.Networks[ident]
  if ok {
    P.Log("already joined Network : "+ident)
  } else {
    N:=joiner()
    P.Networks[ident]=N
    P.iptofriend[ident]=make(map[string]string)
  }
}

type MyError struct {
	What string
}

func (e *MyError) Error() string {
	return (*e).What
}

func (P person)rename_friend(old string, new string) error {
	var validnames = regexp.MustCompile("^([a-zA-Z0-9:]+)$")
	valid:=validnames.MatchString(new)
	if valid {
		F,knowno:=P.friends[old]
	  if knowno {
	    _,knownn:=P.friends[new]
	    if !knownn {
	      F.name=new
	      P.friends[new]=F
	      P.iptofriend[F.Network][F.IP]=new
				P.iptofriend2[F.IP]=new
	      P.chatlog[new]=P.chatlog[old]
	      P.Chat_log(new,"system","renamed "+ old + " to " +new)
	      delete(P.chatlog,old)
	      delete(P.friends,old)
				return nil
	    } else {
	      P.Log("name to rename to already exist: "+  new)
				return &MyError{"name to rename to already exist: "+  new}
	    }
	  } else {
	    P.Log("name to rename does not exist: "+  old)
			return &MyError{"name to rename does not exist: "+  old}
	  }
	} else {
		P.Log("rename has invalid format: "+  new)
		return &MyError{"rename has invalid format: "+  new}
	}

}



func (P person)send_friendrequest(network string, IP string, name string,options interface{}) {
  _,known:=P.friends[name]
  if !known {
    namecolision,knownadr:=P.iptofriend[network][IP]
    if !knownadr {
      priv,pub:=New_pair(options)
      P.friends[name]=friend{name,network,IP,1,priv,AES256key{[8]uint32{0,0,0,0,0,0,0,0}},false}
      P.iptofriend[network][IP]=name
      P.Chat_log(name,"system","sending friendrequest")
      P.msg(name,Message(pub))
    } else {
      P.Log("connection already established under: "+namecolision)
    }

  } else {
    P.Log("name is already in use: "+name)
  }


}

func (P person)recieve()  {
  for net,N := range P.Networks {
    Msgs:=N.recieve()
    for sender,Msgl := range Msgs {
      name,known:=P.iptofriend[net][sender]
      if !known {
        namenew:=net+sender
        switch m0:=Msgl[0].(type) {
        case keymessage:
          priv,pub:=m0.pubkey.clone()
          comm:=priv.merge_with_public(m0.pubkey)
          P.friends[namenew]=friend{namenew,net,sender,2,priv,comm,true}
          P.iptofriend[net][sender]=namenew
          P.msg(namenew,Message(pub))
          P.Chat_log(namenew,"system","comonkey established from frendrequest")
        case textmessage:
          priv,pub:=New_pair(0)
          P.friends[namenew]=friend{namenew,net,sender,5,priv,AES256key{[8]uint32{0,0,0,0,0,0,0,0}},true}
          P.iptofriend[net][sender]=namenew
          P.Chat_log(namenew,"system","recieving unencrpted message:")
          P.Chat_log(namenew,namenew,m0.to_string())
          P.msg(namenew,Message(pub))
        default:
          P.Log("recieved invalid message")
        Msgl=Msgl[1:]
        }
      } else if P.friends[name].status==1 || P.friends[name].status==5 {
        switch m0:=Msgl[0].(type) {
        case keymessage:
          f:=P.friends[name]
          P.friends[name]=friend{f.name,f.Network,f.IP,0,f.keypriv,f.keypriv.merge_with_public(m0.pubkey),true}
          P.Chat_log(name,"system","comonkey established")
        case textmessage:
          P.Chat_log(name,"system","recieving unencrpted message:")
          P.Chat_log(name,name,m0.to_string())
        default:
          P.Log("recieved invalid message")
        }
        Msgl=Msgl[1:]
      }
      name,_=P.iptofriend[net][sender]
      for _,m := range Msgl {
        m:=P.friends[name].keycom.decrypt(m,0)
        switch m0:=m.(type) {
        case keymessage:
          P.Chat_log(name,"system","recieving new public key")
        case textmessage:
          P.Chat_log(name,name,m0.to_string())
        default:
          P.Log("recieved invalid message")
        }
      }
    }
  }
}

func (P person)show_chat(ident string, options interface{}) string  {
  _,ok:=P.friends[ident]
	var out string
  if ok {
		temp:=P.friends[ident]
		temp.unread=false
		P.friends[ident]=temp
		switch op:=options.(type) {
		case [2]int :
			start:=op[0]
			if op[0]>0 {
				if op[0]>=len(P.chatlog[ident]) {
					start=len(P.chatlog[ident])-1
				}
			} else { // op is negative here
				if op[0]>=-len(P.chatlog[ident]) {
					start=len(P.chatlog[ident])+op[0]
				} else {
					start=0
				}
			}
			end:=op[1]
			if op[1]>0 {
				if op[0]>=len(P.chatlog[ident]) {
					start=len(P.chatlog[ident])-1
				}
			} else { // op is negative here
				if op[0]>=-len(P.chatlog[ident]) {
					start=len(P.chatlog[ident])+op[0]
				} else {
					start=0
				}
			}
			if end< start {
				end=start
			}
			for _,message := range P.chatlog[ident][start:end] {
				out=out+"\n"+message
			}
		default:
			for _,message := range P.chatlog[ident] {
				out=out+"\n"+message
			}
		}
  } else {
    P.Log("invalid request to chat with: "+ ident)
  }
	return out
}

func (P person)recieve_msg(M message,sender string)  {
	ident,known:=P.iptofriend2[sender]
	if known {
		known = known &&  P.friends[ident].status<=2
	}
	if known  {
		fr:=P.friends[ident]
		switch m:=M.(type) {
	  case textmessage:
			Mnew:=fr.keycom.decrypt(M,0)
			P.Chat_log(ident,ident,Mnew.to_string())
	  case keymessage:
			if fr.status==1 {
				fr.keycom=fr.keypriv.merge_with_public(m.pubkey)
				fr.status=0
				P.Log("established key with: "+ident)
			} else {
				P.Log("recieved another key message from"+ident)
			}
	  default:
			P.Log("recieved invalid message type")
		}
		fr.unread=true
		P.friends[ident]=fr
	} else {
		namenew:="IP"+sender
		for {
			_,valid:=P.friends[namenew]
			if valid {
				namenew+="0"
			}	else {
				break
			}
		}
		fr:=friend{name:namenew,Network:"LAN",IP:sender,status:5,unread:true}
		P.chatlog[namenew]=[]string{}
		P.iptofriend2[sender]=namenew
		P.friends[namenew]=fr
		//P.iptofriend["LAN"][sender]=namenew
		switch m:=M.(type) {
	  case textmessage:
			P.Chat_log(namenew,"system","recieving unencrpted message:")
			P.Chat_log(namenew,namenew,m.to_string())
	  case keymessage:
			priv,pub:=m.pubkey.clone()
			fr.keypriv=priv
			fr.keycom=priv.merge_with_public(m.pubkey)
			fr.status=2
			P.msg(namenew, pub)
			P.Log("established new connection with" + namenew)
	  default:
			P.Log("recieved invalid message type")
		}
		P.iptofriend2[sender]=namenew
		P.friends[namenew]=fr
	}
}


func (P person)Shutdown()  {
	P.servererror=P.server.Close()
}

// func (P person)msg_LAN(ident string, M message)  {
// 	fr,known:=P.friends[ident]
// 	if known {
// 		IP:=fr.IP
// 		m:=M
// 		if fr.keycom!=nil {
// 			m=fr.keycom.encrypt(M,0)
// 		}
// 		switch m:=m.(type) {
// 	  case textmessage:
// 			req, err := http.NewRequest("POST",IP+"/recieve/text/send/", nil)
// 			if err !=  nil {
// 				fmt.Println(err)
// 				P.Log("got send Error:"+err.Error())
// 				return
// 			}
// 			P.Chat_log(ident,"me",m.text)
// 			resp, err := P.client.Do(req)
// 	    if (err != nil) && (resp != nil) {
// 	        P.Log("conection error"+err.Error())
// 					P.Chat_log(ident,"system","last message not delivered")
// 	    }
// 			defer resp.Body.Close()
// 			fmt.Printf("StatusCode: %d\n", resp.StatusCode)
// 	  case keymessage:
// 			req, err := http.NewRequest("POST",IP+"/recieve/key/send/", nil)
// 			if err !=  nil {
// 				P.Log("got send Error:"+err.Error())
// 				return
// 			}
// 			resp, err := P.client.Do(req)
// 	    if err != nil {
// 	        P.Log("conection error"+err.Error())
// 					P.Chat_log(ident,"system","last message not delivered")
// 	    }
// 			defer resp.Body.Close()
// 			fmt.Printf("StatusCode: %d\n", resp.StatusCode)
// 	  default:
// 			P.Log("recieved invalid message type")
// 			return
// 		}
// 	} else {
// 		P.Log("unknown Friend :"+ ident)
// 	}
// }
