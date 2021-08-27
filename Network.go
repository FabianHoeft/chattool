package main

import(
)

// Networks provide acess to other users through LAN etc
// currently omessages are just passed along
// Network requires a password/token for accsess to safly allow multiple persons use the same network
// without identitytheft and to avoid impersination

type Network interface{
  send(ip,ip,message,password)
  recieve(ip,password) map[ip][]message
  leave(ip,password)
  join() Networkstore   //netpointer of the return value should always be a pointer to the network
}


// stores Networks for the users
//provides send and recieve function without requiering passowrd
type Networkstore struct{
  myip ip
  mypasswort password
  netpointer Network
}



type ip struct{
  value [16]uint8
}


type password struct{
  value [8]uint32
}


func (N Networkstore)send(ipr ip,m message) {
  N.netpointer.send(N.myip,ipr,m,N.mypasswort)
}

func (N Networkstore)recieve() map[ip][]message {
  return (N.netpointer.recieve)(N.myip,N.mypasswort)
}

func (N Networkstore)leave() {
  (N.netpointer.leave)(N.myip,N.mypasswort)
}


type Intranet struct{
  msg map[ip]map[ip][]message
  member map[ip]password
  count uint32
}

func NewIntranet()Intranet  {
  msg:=make(map[ip]map[ip][]message)
  member:=make(map[ip]password)
  var count uint32
  return Intranet{msg,member,count}
}


func (I Intranet)join() Networkstore {
  I.count=I.count+1
  newip:=ip{[16]uint8{0,0,0,0,0,0,0,0,0,0,0,0,uint8((I.count>>24)&0xFF),uint8((I.count>>16)&0xFF),uint8((I.count>>8)&0xFF),uint8(I.count&0xFF)}}
  newpassword:=password{SHA256(string(newip.value[:]))}
  I.member[newip]=newpassword
  return Networkstore{newip,newpassword,&I}
}


func (I Intranet)send(ips ip, ipr ip, M message, P password)  {
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
      I.msg[ipr]=make(map[ip][]message)
      I.msg[ipr][ipr]=[]message{M}
    }
  }
}

func (I Intranet)recieve(ipr ip,P password) map[ip][]message {
  out:=make(map[ip][]message)
  Pchek,ok:=I.member[ipr]
  if ok && Pchek==P {
    out=I.msg[ipr]
    delete(I.msg,ipr)
  }
  return out
}

func (I Intranet)leave(ips ip, P password)  {
  Pchek,ok:=I.member[ips]
  if ok && Pchek==P {
    delete(I.msg,ips)
    delete(I.member,ips)
  }
}
