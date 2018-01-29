package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

type pass struct {
	Title    string
	URL      string
	Username string
	Email    string
	Password string
	Notes    string
}

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	r := csv.NewReader(file)
	r.FieldsPerRecord = -1

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, line := range records {
		p := pass{
			Title: line[0],
		}
		for i, cell := range line {
			switch cell {
			case "Email":
				if p.Username == "" {
					p.Username = line[i+1]
				} else {
					p.Email = line[i+1]
				}
			case "Username":
				p.Username = line[i+1]
			case "Password":
				p.Password = line[i+1]
			case "URL":
				p.URL = strings.SplitN(line[i+1], "?", 2)[0]
			}
		}

		if p.Password == "" && len(line) == 4 {
			p.Username = line[1]
			p.Password = line[2]
		}

		p.Notes = strings.Join(line, "\n")

		id := p.URL
		if p.URL == "" {
			id = p.Title
		}

		id = strings.Replace(id, "https://", "", -1)
		id = strings.Replace(id, "http://", "", -1)
		id = strings.Replace(id, "www.", "", -1)

		if p.Username == "" {
			log.Printf("No username %+v", p)
			continue
		}

		subProcess := exec.Command("pass", "insert", "--multiline", fmt.Sprintf("%s/%s", id, p.Username))

		stdin, err := subProcess.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}
		defer stdin.Close()

		subProcess.Stdout = os.Stdout
		subProcess.Stderr = os.Stderr

		if err = subProcess.Start(); err != nil {
			log.Fatal(err)
		}

		io.WriteString(stdin, fmt.Sprintf("%s\n", p.Password))
		io.WriteString(stdin, fmt.Sprintf("login: %s\n", p.Username))
		io.WriteString(stdin, fmt.Sprintf("email: %s\n", p.Email))
		io.WriteString(stdin, fmt.Sprintf("title: %s\n", p.Title))
		io.WriteString(stdin, fmt.Sprintf("notes: %s\n", p.Notes))
		stdin.Close()
		subProcess.Wait()
	}
}
