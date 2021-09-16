package main

import (
	"fmt"
	"html/template"
	"math/big"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

func getIP(r *http.Request) string {
	regex, _ := regexp.Compile("([a-zA-Z0-9.:\\]\\[]+):([^:]+)$")
	IPstr := r.Header.Get("X-FORWARDED-FOR")
	if IPstr == "" {
		IPstr = r.RemoteAddr
	}
	m := regex.FindStringSubmatch(IPstr)
	return m[1]
}

type chat struct {
	Name    string
	IP      string
	Content []todo
}

type todo struct {
	Title string
	Done  bool
}

func newChat(ident string, P *person) chat {
	temp := (*P).friends[ident]
	temp.unread = false
	(*P).friends[ident] = temp
	a := strings.Split((*P).Showchat(ident, 0), "\n")
	out := make([]todo, len(a))
	for i, ai := range a {
		out[i] = todo{ai, true}
	}
	return chat{ident, (*P).friends[ident].IP, out}
}

func chatHandler(w http.ResponseWriter, r *http.Request, title string, P *person) {
	loc := strings.Split(title, "/")
	ident := loc[0]
	_, ok := (*P).friends[ident]
	if ok {
		if r.Method != http.MethodPost {
			p := newChat(ident, P)
			renderTemplate(w, "chat", p)
		} else {
			if len(loc) <= 1 {
				http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
				return
			}
			switch loc[1] {
			case "rename":
				body := r.FormValue("newname")
				err := (*P).RenameFriend(ident, body)
				if err == nil {
					http.Redirect(w, r, "/chat/"+body, http.StatusFound)
				} else {
					http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
				}
			case "send":
				text := r.FormValue("MSG")
				(*P).MsgLAN(ident, newMessage(text))
				http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
			case "add":
				http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
			case "remove":
				confirm := r.FormValue("Confirm")
				if confirm == "true" {
					http.Redirect(w, r, "/main/", http.StatusFound)
				} else {
					http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
				}
			default:
				http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
			}
		}
	} else {
		ident = "addfriend"
		if len(loc) == 1 {
			p := struct{ Title string }{Title: ident}
			renderTemplate(w, "addfriend", p)
		} else {
			switch loc[1] {
			case "add":
				Name := r.FormValue("Name")
				Net := r.FormValue("Net")
				IP := r.FormValue("IP")
				(*P).SendFriendrequest(Net, IP, Name, 0)
				http.Redirect(w, r, "/chat/"+Name, http.StatusFound)
			default:
				http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
			}
		}
	}
}

func mainHandler(w http.ResponseWriter, r *http.Request, title string, P *person) {
	loc := strings.Split(title, "/")
	if len(loc) == 1 {
		switch loc[0] {
		case "me":
			var friendlist []struct {
				Unread bool
				Name   string
			}
			for _, fr := range (*P).friends {
				friendlist = append(friendlist, struct {
					Unread bool
					Name   string
				}{fr.unread, fr.name})
			}
			renderTemplate(w, "main", struct {
				LocalIP string
				Friends []struct {
					Unread bool
					Name   string
				}
			}{getLocalIP(), friendlist})
		case "log":
			var logs []struct{ Line string }
			for _, line := range strings.Split((*P).log, "\n") {
				logs = append(logs, struct{ Line string }{line})
			}
			renderTemplate(w, "log", struct{ Log []struct{ Line string } }{logs})
		case "shutdown":
			confirm := r.FormValue("Confirm")
			if confirm == "true" {
				http.NotFound(w, r)
				(*P).Shutdown()
			} else {
				http.Redirect(w, r, "/main/me", http.StatusFound)
			}
		default:
			http.Redirect(w, r, "/main/me", http.StatusFound)
		}
	} else {
		http.Redirect(w, r, "/main/me", http.StatusFound)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request, title string, P *person) {
	loc := strings.Split(title, "/")
	if loc[0] == (*P).cookie.Value {
		if r.Method == http.MethodPost && len(loc) >= 2 {
			switch loc[1] {
			case "login":
				psw := r.FormValue("Password")
				valid := checkpassword(psw)
				if valid {
					http.SetCookie(w, &((*P).cookie))
					http.Redirect(w, r, "/main/me", http.StatusFound)
					(*P).password = psw
					(*P).fromFile()
					(*P).loadChats()
				} else {
					renderTemplate(w, "login", struct {
						Site  string
						Error string
					}{loc[0], "wrong password!"})
				}
			case "register":
				if haspassword() {
					http.Redirect(w, r, "/login/"+loc[0], http.StatusFound)
				} else {
					psw0 := r.FormValue("Password0")
					psw1 := r.FormValue("Password1")
					if psw0 == psw1 {
						good := strongpassword(psw0)
						if good {
							writepassword(psw0)
							(*P).password = psw0
							http.SetCookie(w, &((*P).cookie))
							http.Redirect(w, r, "/main/me", http.StatusFound)
						} else {
							renderTemplate(w, "register", struct {
								Site  string
								Error string
							}{loc[0], "weak Password"})
						}
					} else {
						renderTemplate(w, "register", struct {
							Site  string
							Error string
						}{loc[0], "Passwords do not match"})
					}
				}
			case "remove":
				confirm := r.FormValue("Confirm")
				if confirm == "true" {
					deletefiles()
					http.Redirect(w, r, "/login/"+loc[0], http.StatusFound)
				} else {
					http.Redirect(w, r, "/login/"+loc[0], http.StatusFound)
				}
			default:
				http.Redirect(w, r, "/login/"+loc[0], http.StatusFound)
			}
		} else {
			if haspassword() {
				renderTemplate(w, "login", struct {
					Site  string
					Error string
				}{loc[0], ""})
			} else {
				renderTemplate(w, "register", struct {
					Site  string
					Error string
				}{loc[0], ""})
			}
		}
	} else {
		time.Sleep(time.Second)
		http.NotFound(w, r)
	}
}

func recieveHandler(w http.ResponseWriter, r *http.Request, title string, P *person) {
	loc := strings.Split(title, "/")
	if r.Method == http.MethodPost && len(loc) >= 1 {
		switch loc[0] {
		case "key":
			var value, root, modu big.Int
			IP := getIP(r)
			// IP := r.FormValue("IP")
			_, err0 := (&value).SetString(r.FormValue("Value"), 0)
			_, err1 := (&root).SetString(r.FormValue("Root"), 0)
			_, err2 := (&modu).SetString(r.FormValue("Modu"), 0)
			if err0 && err1 && err2 {
				(*P).Log("keyconversion error")
				pubkey := mod{value, modu, root}
				(*P).RecieveMsg(newMessage(pubkey), IP)
			}
			renderTemplate(w, "recieve_key", struct{ Error string }{Error: "send success"})
		case "text":
			IP := getIP(r)
			// IP := r.FormValue("IP")
			Text := r.FormValue("Text")
			(*P).RecieveMsg(newMessage(Text), IP)
			renderTemplate(w, "recieve_text", struct{ Error string }{Error: "send success"})
		case "notice":
			http.Redirect(w, r, "/recieve/text", http.StatusFound)
		default:
			http.Redirect(w, r, "/recieve/text", http.StatusFound)
		}
	} else {
		switch loc[0] {
		case "key":
			renderTemplate(w, "recieve_key", struct{ Error string }{Error: ""})
		case "text":
			renderTemplate(w, "recieve_text", struct{ Error string }{Error: ""})
		case "notice":
			http.Redirect(w, r, "/recieve/text", http.StatusFound)
		default:
			http.Redirect(w, r, "/recieve/text", http.StatusFound)
		}
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func strongpassword(psw string) bool {
	good, err := regexp.MatchString("^[a-zA-Z0-9]{5,}$", psw)
	if err != nil {
		return false
	}
	return good
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string, *person), P *person, checkcookie bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if checkcookie {
			valid := validatecookie(r, P)
			if !valid {
				time.Sleep(time.Second)
				http.NotFound(w, r)
				return
			}
		}
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			fmt.Println(r.URL.Path)
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2], P)
	}
}

var templates = template.Must(template.ParseFiles("HTML/addfriend.html",
	"HTML/chat.html", "HTML/main.html", "HTML/recieve_key.html", "HTML/recieve_text.html",
	"HTML/log.html", "HTML/login.html", "HTML/register.html"))
var validPath = regexp.MustCompile("^/(chat|main|recieve|login)/([a-zA-Z0-9/:]+)$")

func netserver(P *person) {
	http.HandleFunc("/recieve/", makeHandler(recieveHandler, P, false))
	http.HandleFunc("/main/", makeHandler(mainHandler, P, true))
	http.HandleFunc("/chat/", makeHandler(chatHandler, P, true))
	http.HandleFunc("/login/", makeHandler(loginHandler, P, false))
	err := (*P).server.ListenAndServe()
	if err == http.ErrServerClosed {
		fmt.Println("starting shutdown")
		time.Sleep(time.Second)
		if (*P).servererror != nil {
			fmt.Println((*P).servererror)
		}
		fmt.Println("shutdown complete")
	}
	fmt.Println("unexpected shutdown")
	fmt.Println(err)
}

func newserver(P *person) {
	fmt.Println("for login visit:")
	fmt.Println("http://localhost:8080/login/" + (*P).cookie.Value)

	fmt.Println("for login external access:")
	fmt.Println(getLocalIP())

	go func() {
		time.Sleep(300 * time.Second)
		(*P).saveChats()
	}()

	openbrowser("http://localhost:8080/login/" + (*P).cookie.Value)
	netserver(P)
}

func validatecookie(r *http.Request, P *person) bool {
	C, err := r.Cookie("validate")
	if err != nil {
		return false
	}
	if (*C).Value == (*P).cookie.Value {
		return true
	}
	return false
}

func openbrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Println(err.Error())
	}
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
