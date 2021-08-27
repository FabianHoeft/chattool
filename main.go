package main

import(
  "fmt"
)

func main() {

  LAN:=NewIntranet()
  A:=NewPerson()
  B:=NewPerson()
  A.Add_Network("LAN",LAN.join)
  B.Add_Network("LAN",LAN.join)
  A.send_friendrequest("LAN",B.Networks["LAN"].myip,"B",0)
  B.recieve()
  B.rename_friend("LAN0000000000000001","A")
  A.recieve()
  A.chat("B","hi")
  B.recieve()


  fmt.Println(A.show_chat("B",0))
  fmt.Println(B.show_chat("A",0))

}
