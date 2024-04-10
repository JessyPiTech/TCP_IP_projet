package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

var httpServer *http.Server
var con = "a"

var text = ""

type Message struct {
	Nom          string
	Destinataire string
	Text         string
}

func gestionErreur(err error) {
	if err != nil {
		fmt.Println("erreur")
		panic(err)
	}
}

const (
	Nom  = "Boug"
	IP   = "10.35.1.210" // IP serveur
	PORT = "3569"        // Port serveur
)

func newMessage(Nom string, Destinataire string, Text string) Message {
	return Message{
		Nom:          Nom,
		Destinataire: Destinataire,
		Text:         Text,
	}
}
func main() {
	fmt.Println("0")
	var wg sync.WaitGroup
	Messages := map[string][]Message{
		"Messages": {},
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", IP, PORT))
	gestionErreur(err)
	h1 := func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("indextest.html"))

		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				text := strings.ToLower(r.FormValue("Ms"))
				text = strings.ToLower(strings.TrimSpace(text))
				if con == text || text == "" {
					return
				}
				con = text
				fmt.Println("compris :", text)
				fmt.Println(Messages["Messages"])
				text = text + "\n"
				text = Nom + ".000." + text
				conn.Write([]byte(text))
				fmt.Println("envoie-->")
			}
		}()
		go func() { // goroutine dédiée à la reception des messages du serveur
			defer wg.Done()
			for {
				message, err := bufio.NewReader(conn).ReadString('\n')
				gestionErreur(err)
				if message != "" {
					fmt.Println("<--recoit")
				}
				if err != nil {
					fmt.Println("ressoit rien", err)
					continue
				}
				gestionErreur(err)
				con = message
				s := strings.SplitN(message, ".000.", 2)
				name := string(s[0])
				txms := string(s[1])
				Messages["Messages"] = append(Messages["Messages"], newMessage(name, Nom, txms))
				fmt.Println(Messages["Messages"])
			}
		}()
		tmpl.Execute(w, Messages)
	}
	http.HandleFunc("/", h1)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
