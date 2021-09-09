package main

import(
  "fmt"
)

func main() {

  intra:=NewIntranet()

  A:=NewPerson()
  B:=NewPerson()

  A.Add_Network("intra",intra.join)
  B.Add_Network("intra",intra.join)

  A.send_friendrequest("intra",B.Networks["intra"].myip,"B",0)
  B.recieve()
  B.rename_friend("intra:::::::1","A")
  A.recieve()
  A.chat("B","hi")
  B.recieve()
  B.chat("A","hello")
  A.recieve()
  fmt.Println(A.show_chat("B",0))
  fmt.Println(B.show_chat("A",0))


  Newserver(&A)

  // A.Add_Network("LAN",LAN.join)
  // B.Add_Network("LAN",LAN.join)
}
