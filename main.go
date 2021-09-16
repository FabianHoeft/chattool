package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func main() {

	intra := newIntranet()

	A := newPerson()
	B := newPerson()

	A.AddNetwork("intra", intra.join)
	B.AddNetwork("intra", intra.join)

	A.SendFriendrequest("intra", B.Networks["intra"].myip, "B", 0)
	B.recieve()
	B.RenameFriend("intra:::::::1", "A")
	A.recieve()
	A.chat("B", "hi")
	B.recieve()
	B.chat("A", "hello")
	A.recieve()
	fmt.Println(A.Showchat("B", 0))
	fmt.Println(B.Showchat("A", 0))

	go func() {
		time.Sleep(10 * time.Second)
		hc := http.Client{}
		form := url.Values{}
		form.Add("IP", "http://localhost:8080")
		form.Add("Text", "hi")
		req, err := http.NewRequest("POST", "http://localhost:8080/recieve/text/send", strings.NewReader(form.Encode()))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		req.PostForm = form
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		resp, err := hc.Do(req)
		defer resp.Body.Close()
	}()

	newserver(&A)

	// A.Add_Network("LAN",LAN.join)
	// B.Add_Network("LAN",LAN.join)
}
