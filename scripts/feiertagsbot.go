package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/template"
	"time"
)

type Configuration struct {
	Sourcefile string
	Logfile    string
	// Twitter_access_token        string
	// Twitter_access_token_secret string
	// Twitter_consumer_key        string
	// Twitter_consumer_secret     string
	Sleep_time_in_hours         int
	Sleep_time_margin_in_hours  int
	Debug int
}

type Feiertag struct {
	Datum       string
	Bezeichnung string
	Religion    string
}

var configuration Configuration
var logger *log.Logger
var logfile *os.File

func init_bot() {

	// config
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	configuration = Configuration{}
	err := decoder.Decode(&configuration)
	check(err)

	// logging
	logfile, err = os.OpenFile(configuration.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	check(err)
	logger = log.New(logfile, "", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Print("Started...")

	// init random generator
	rand.Seed(time.Now().UnixNano())

}

// check panics if an error is detected
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func read_sourcefile() {
	f, err := os.Open(configuration.Sourcefile)
	check(err)
	defer f.Close()

	current_time := time.Now()
	time_string := current_time.Format("02.01.2006")
	// debug
	time_string = "15.02.2017"
	logger.Printf("Today is %s", time_string)
	lineCount := 0

	r := csv.NewReader(f)
	r.Comma = ';'

	for {
		record, err := r.Read()
		if err == io.EOF {
			logger.Printf("Found %d records", lineCount)
			break
		} else if err != nil {
			log.Fatal(err)
		}
		lineCount += 1
		s := Feiertag{Datum: record[0], Bezeichnung: strings.TrimSpace(record[1]), Religion: record[2]}
		if s.Datum == time_string {
			logger.Printf("Found holiday: '%s'", s.Bezeichnung)
			fmt.Println(create_tweet_message(s))
		}
		//fmt.Println(s)
	}

}

func create_tweet_message(f Feiertag) string {
	var tpl_output bytes.Buffer
	wikisearch_url := "https://de.wikipedia.org/w/index.php?search=" + url.QueryEscape(f.Bezeichnung)
	tweet_templates := [3]string{"Übrigens: Heute ist \"{{.Bezeichnung}}\"!",
		"Kleine Erinnerung: heute ist \"{{.Bezeichnung}}\"!",
		"Na? Auch fast auf \"{{.Bezeichnung}}\" vergessen?"}
	wiki_templates := [2]string{"(Mehr Infos dazu in der Wikipedia " + wikisearch_url, "(Wikipedia weiß mehr: " + wikisearch_url}

	// randomisierter name, damit template neu gebaut wird
	tmpl, err := template.New(fmt.Sprintf("tweet %d", rand.Intn(100))).Parse(tweet_templates[rand.Intn(len(tweet_templates))] + " " + wiki_templates[rand.Intn(len(wiki_templates))])
	check(err)

	err = tmpl.Execute(&tpl_output, f)
	check(err)

	return tpl_output.String()

}

func main() {

	// catch interrupts
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		logger.Print("ended...")
		os.Exit(1)
	}()

	init_bot()
	// infinite loop
	for {

		read_sourcefile()
		sleep_hours := configuration.Sleep_time_in_hours + (configuration.Sleep_time_margin_in_hours - 2*rand.Intn(configuration.Sleep_time_margin_in_hours))
		logger.Printf("Will go to sleep for %d hours..", sleep_hours)
		time.Sleep(time.Duration(sleep_hours) * time.Hour)
	}

}
