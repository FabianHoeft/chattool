package main

import(
	"time"
  "strconv"
)


type friend struct{
  name string
  Network string
  IP ip
  status uint32 // 0 done, 1 sending, 2 recieving, 3 blocked, 4 ignored 5 invalid
  keypriv mod
  keycom key
}


func NewPerson() person  {
  friends:=make(map[string]friend)
  iptofriend:=make(map[string]map[ip]string)
  chatlog:=make(map[string][]string)
  var log []string
  Networks:=make(map[string]Networkstore)
  return person{friends,iptofriend,chatlog,log,Networks}
}

type person struct{
  friends map[string]friend
  iptofriend map[string]map[ip]string
  chatlog map[string][]string
  log []string
  Networks map[string]Networkstore

}

func (P person)Log(m string)  {
  P.log=append(P.log,time.Now().Format(time.RFC850)+" : "+m)
}

func (P person)Chat_log(name string, sender string, message string)  {
  P.chatlog[name]=append(P.chatlog[name],"["+sender+"]:"+message)
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
    P.iptofriend[ident]=make(map[ip]string)
  }
}



func (P person)rename_friend(old string, new string)  {
  F,knowno:=P.friends[old]
  if knowno {
    _,knownn:=P.friends[new]
    if !knownn {
      F.name=new
      P.friends[new]=F
      P.iptofriend[F.Network][F.IP]=new
      P.chatlog[new]=P.chatlog[old]
      P.Chat_log(new,"system","renamed "+ old + " to " +new)
      delete(P.chatlog,old)
      delete(P.friends,old)
    } else {
      P.Log("name to rename to already exist: "+  new)
    }
  } else {
    P.Log("name to rename does not exist: "+  old)
  }
}



func (P person)send_friendrequest(network string, IP ip, name string,options interface{})  {
  _,known:=P.friends[name]
  if !known {
    namecolision,knownadr:=P.iptofriend[network][IP]
    if !knownadr {
      priv,pub:=New_pair(options)
      P.friends[name]=friend{name,network,IP,1,priv,AES256key{[8]uint32{0,0,0,0,0,0,0,0}}}
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
        namenew:=net
        for i := 0; i < 16; i++ {
          namenew=namenew+strconv.Itoa(int(sender.value[i]))
        }
        switch m0:=Msgl[0].(type) {
        case keymessage:
          priv,pub:=m0.pubkey.clone()
          comm:=priv.merge_with_public(m0.pubkey)
          // fmt.Println(comm)
          P.friends[namenew]=friend{namenew,net,sender,2,priv,comm}
          P.iptofriend[net][sender]=namenew
          P.msg(namenew,Message(pub))
          P.Chat_log(namenew,"system","comonkey established from frendrequest")
        case textmessage:
          priv,pub:=New_pair(0)
          P.friends[namenew]=friend{namenew,net,sender,5,priv,AES256key{[8]uint32{0,0,0,0,0,0,0,0}}}
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
          P.friends[name]=friend{f.name,f.Network,f.IP,0,f.keypriv,f.keypriv.merge_with_public(m0.pubkey)}
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
