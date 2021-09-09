package main

import(
  "strconv"
)

// Networks provide acess to other users through LAN etc
// currently omessages are just passed along
// Network requires a password/token for accsess to safly allow multiple persons use the same network
// without identitytheft and to avoid impersination

type Network interface{
  send(string,string,message,password)
  recieve(string,password) map[string][]message
  leave(string,password)
  join() Networkstore   //netpointer of the return value should always be a pointer to the network
}


// stores Networks for the users
//provides send and recieve function without requiering passowrd
type Networkstore struct{
  myip string
  mypasswort password
  netpointer Network
}


type password struct{
  value [8]uint32
}


func (N Networkstore)send(ipr string,m message) {
  N.netpointer.send(N.myip,ipr,m,N.mypasswort)
}

func (N Networkstore)recieve() map[string][]message {
  return (N.netpointer.recieve)(N.myip,N.mypasswort)
}

func (N Networkstore)leave() {
  (N.netpointer.leave)(N.myip,N.mypasswort)
}


type Intranet struct{
  msg map[string]map[string][]message
  member map[string]password
  count int
}

func NewIntranet()Intranet  {
  msg:=make(map[string]map[string][]message)
  member:=make(map[string]password)
  var count int
  return Intranet{msg,member,count}
}


func (I Intranet)join() Networkstore {
  I.count=I.count+1
  newip:=":::::::"+strconv.Itoa(I.count)
  newpassword:=password{SHA256(newip)}
  I.member[newip]=newpassword
  return Networkstore{newip,newpassword,&I}
}


func (I Intranet)send(ips string, ipr string, M message, P password)  {
  Pchek,ok:=I.member[ips]
  if ok && Pchek==P {
    _,ipr_has_message:=I.msg[ipr]
    if ipr_has_message {
      _,ipr_has_message_from_ips:=I.msg[ipr][ips]
      if ipr_has_message_from_ips {
        I.msg[ipr][ips]=append(I.msg[ipr][ips],M)
      } else {
        I.msg[ipr][ips]=[]message{M}
      }
    } else {
      I.msg[ipr]=make(map[string][]message)
      I.msg[ipr][ipr]=[]message{M}
    }
  }
}

func (I Intranet)recieve(ipr string,P password) map[string][]message {
  out:=make(map[string][]message)
  Pchek,ok:=I.member[ipr]
  if ok && Pchek==P {
    out=I.msg[ipr]
    delete(I.msg,ipr)
  }
  return out
}

func (I Intranet)leave(ips string, P password)  {
  Pchek,ok:=I.member[ips]
  if ok && Pchek==P {
    delete(I.msg,ips)
    delete(I.member,ips)
  }
}



type LAN struct{
  connection string
  myip string
  in,out chan message
}


func (L LAN)send(ips string, ipr string, M message, P password)  {
}

func (L LAN)join() Networkstore {
  newip:=":::::::"
  return Networkstore{newip,password{},&L}
}

func (L LAN)leave(ips string, P password)  {
}

func (L LAN)recieve(ipr string,P password) map[string][]message {
  out:=make(map[string][]message)
  return out
}
