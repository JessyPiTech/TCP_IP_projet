package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
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
	Nom  = "invite"
	IP   = "10.35.1.133" // IP serveur
	PORT = "3569"        // Port serveur
	clef = "example-key-1234"
)

func newMessage(Nom string, Destinataire string, Text string) Message {
	return Message{
		Nom:          Nom,
		Destinataire: Destinataire,
		Text:         Text,
	}
}

var key = []byte(clef)

// Fonction pour chiffrer un message
func encrypt(plainText string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Le texte doit être complété avec des zéros pour avoir une longueur multiple de la taille du bloc
	plainText = fmt.Sprintf("%-32s", plainText)

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], []byte(plainText))

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Fonction pour déchiffrer un message
func decrypt(cipherText string, secu string) (string, error) {
	block, err := aes.NewCipher([]byte(secu))
	if err != nil {
		return "", err
	}

	cipherTextBytes, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	if len(cipherTextBytes) < aes.BlockSize {
		return "", fmt.Errorf("cipherText trop court")
	}

	iv := cipherTextBytes[:aes.BlockSize]
	cipherTextBytes = cipherTextBytes[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherTextBytes, cipherTextBytes)

	// Supprimer les zéros ajoutés lors du chiffrement
	return strings.TrimRight(string(cipherTextBytes), "\x00"), nil
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

				text = Nom + ".000." + text

				text, err = encrypt(text)

				gestionErreur(err)
				text = clef + ".000." + text + "\n"
				conn.Write([]byte(text))
				fmt.Println("envoie-->")
			}
		}()
		go func() { // goroutine dédiée à la reception des messages du serveur
			defer wg.Done()
			for {
				fmt.Println("1")
				message, err := bufio.NewReader(conn).ReadString('\n')
				gestionErreur(err)
				fmt.Println("2")

				if message != "" {
					fmt.Println("<--recoit")
				}
				fmt.Println("3")
				if err != nil {
					fmt.Println("ressoit rien", err)
					continue
				}
				gestionErreur(err)
				con = message
				s := strings.SplitN(message, ".000.", 2)
				secu := string(s[0])
				message = string(s[1])
				message, err = decrypt(message, secu)
				gestionErreur(err)
				s = strings.SplitN(message, ".000.", 2)
				fmt.Println(s)
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
	fmt.Println("Server started on :5050")
	log.Fatal(http.ListenAndServe(":5050", nil))
}
