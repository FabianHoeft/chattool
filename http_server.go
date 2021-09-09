package main

import(
  "fmt"
  "net/http"
  "html/template"
  "regexp"
  "strings"
  "math/big"
  "time"
  "runtime"
  "os/exec"
)


type Chat struct{
  Name string
  IP string
  Content []Todo
}

type Todo struct {
    Title string
    Done  bool
}



func NewChat(ident string, P *person) Chat{
  temp:=(*P).friends[ident]
  temp.unread=false
  (*P).friends[ident]=temp
  a:=(*P).chatlog[ident]
  out:=make([]Todo,len(a))
  for i,ai := range a {
    out[i]=Todo{ai,true}
  }
  return Chat{ident ,(*P).friends[ident].IP, out }
}


func chatHandler(w http.ResponseWriter, r *http.Request, title string, P *person) {
  loc:=strings.Split(title,"/")
  ident:=loc[0]
  _,ok:=(*P).friends[ident]
	if ok {
    if r.Method != http.MethodPost {
      p:=NewChat(ident,P)
      renderTemplate(w,"chat",p)
    } else {
      if len(loc)<=1{
        http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
        return
      }
      switch loc[1] {
      case "rename":
        body := r.FormValue("newname")
        err:=(*P).rename_friend(ident,body)
        if err==nil {
          http.Redirect(w, r, "/chat/"+body, http.StatusFound)
        } else {
          http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
        }
      case "send":
        body := r.FormValue("body")
        (*P).chat(ident,body)
        http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
      case "add":
        http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
      case "remove":
        confirm := r.FormValue("Confirm")
        if confirm == "true"{
          http.Redirect(w, r, "/main/", http.StatusFound)
        } else {
          http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
        }
      default:
        http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
      }
    }
    // if len(loc)==1{
    //   p:=NewChat(ident,P)
    //   renderTemplate(w, "chat", p)
    // } else {
    //   switch loc[1] {
    //   case "rename":
    //     body := r.FormValue("newname")
    //     err:=(*P).rename_friend(ident,body)
    //     if err!=nil {
    //       http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
    //     } else {
    //       http.Redirect(w, r, "/chat/"+body, http.StatusFound)
    //     }
    //   case "send":
    //     msg := r.FormValue("msg")
    //     (*P).msg_LAN(ident,Message(msg))
    //     http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
    //   case "add":
    //     http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
    //   case "remove":
    //     confirm := r.FormValue("Confirm")
    //     if confirm == "true"{
    //       http.Redirect(w, r, "/main/", http.StatusFound)
    //     } else {
    //       http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
    //     }
    //   default:
    //   	http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
    //   }
    // }
	} else {
    ident="addfriend"
    if len(loc)==1{
      p:=struct{Title string}{Title: ident}
      renderTemplate(w, "addfriend", p)
    } else {
      switch loc[1] {
      case "add":
        Name := r.FormValue("Name")
        Net := r.FormValue("Net")
        IP := r.FormValue("IP")
        (*P).send_friendrequest(Net, IP, Name ,0)
        http.Redirect(w, r, "/chat/"+Name, http.StatusFound)
      default:
      	http.Redirect(w, r, "/chat/"+ident, http.StatusFound)
      }
    }
  }
}


func mainHandler(w http.ResponseWriter, r *http.Request, title string, P *person) {
  loc:=strings.Split(title,"/")
  if len(loc)==1 {
    switch loc[0] {
    case "me":
      var friendlist []struct{Unread bool; Name string}
      for _,fr := range (*P).friends {
        friendlist=append(friendlist,struct{Unread bool; Name string}{fr.unread,fr.name})
      }
      renderTemplate(w, "main", struct{Friends []struct{Unread bool; Name string}}{friendlist})
    case "log":
      var logs []struct{Line string}
      for _,line := range (*P).log {
        logs=append(logs,struct{Line string}{line})
      }
      renderTemplate(w, "log", struct{Log []struct{Line string}}{logs})
    case "shutdown":
      confirm := r.FormValue("Confirm")
      if confirm == "true"{
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
  time.Sleep(time.Second)
  if title==(*P).cookie.Value {
    http.SetCookie(w,&((*P).cookie))
    http.Redirect(w, r, "/main/me", http.StatusFound)
  } else {
    http.NotFound(w, r)
  }
}




func recieveHandler(w http.ResponseWriter, r *http.Request, title string, P *person) {
  loc:=strings.Split(title,"/")
  if len(loc)>=0{
    switch loc[0] {
    case "key":
      if len(loc)==2 {
        var value,root,modu big.Int
        IP := r.FormValue("IP")
        _,err0:=(&value).SetString(r.FormValue("Value"),0)
        _,err1:=(&root).SetString(r.FormValue("Root"),0)
        _,err2:=(&modu).SetString(r.FormValue("Modu"),0)
        if err0 && err1 && err2 {
          (*P).Log("keyconversion error")
          pubkey:=mod{value,modu,root}
          (*P).recieve_msg(Message(pubkey),IP)
        }
        http.Redirect(w, r, "/recieve/text", http.StatusFound)
      } else {
        renderTemplate(w, "recieve_key", 0)
      }
    case "text":
      if len(loc)==2 {
        IP := r.FormValue("IP")
        Text := r.FormValue("Text")
        (*P).recieve_msg(Message(Text),IP)
        http.Redirect(w, r, "/recieve/text", http.StatusFound)
      } else {
        renderTemplate(w, "recieve_text", 0)
      }
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




func makeHandler(fn func(http.ResponseWriter, *http.Request, string, *person), P *person, checkcookie bool) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if checkcookie {
      valid:=validate_cookie(r, P)
      if !valid {
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
   "HTML/chat.html", "HTML/main.html", "HTML/recieve_key.html", "HTML/recieve_text.html","HTML/log.html"))
var validPath = regexp.MustCompile("^/(chat|main|recieve|login)/([a-zA-Z0-9/]+)$")

func Netserver( P *person)  {
  http.HandleFunc("/recieve/", makeHandler(recieveHandler,P,false))
  http.HandleFunc("/main/", makeHandler(mainHandler,P,true))
  http.HandleFunc("/chat/", makeHandler(chatHandler,P,true))
  http.HandleFunc("/login/", makeHandler(loginHandler,P,false))
  err:=(*P).server.ListenAndServe()
  if err==http.ErrServerClosed {
    fmt.Println("starting shutdown")
    time.Sleep(time.Second)
    if (*P).servererror!=nil {
      fmt.Println((*P).servererror)
    }
    fmt.Println("shutdown complete")
  }
  fmt.Println("unexpected shutdown")
  fmt.Println(err)
}


func Newserver(P *person)  {
  fmt.Println("for login visit:")
  fmt.Println("http://localhost:8080/login/"+(*P).cookie.Value)

  openbrowser("http://localhost:8080/login/"+(*P).cookie.Value)
  Netserver(P)
}

func validate_cookie( r *http.Request, P *person) bool{
  C, err:=r.Cookie("validate")
  if err!= nil {
    return false
  }
  if (*C).Value==(*P).cookie.Value {
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
