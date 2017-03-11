package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
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
	Api_url string
	Debug   int
}

type Feiertag struct {
	Date     string
	Name     string
	Religion string
}

var configuration Configuration
var logger *log.Logger
var logfile *os.File

func init() {

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

func check_current_date() {

	current_time := time.Now()
	date_string := current_time.Format("02.01.2006")
	// debug
	date_string = "15.02.2017"
	logger.Printf("Today is %s", date_string)
	query_url := fmt.Sprintf("%s/%s", configuration.Api_url, date_string)

	logger.Printf("Querying %s", query_url)
	response, err := http.Get(query_url)
	check(err)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	check(err)
	feiertag := new(Feiertag)
	err = json.Unmarshal(body, feiertag)
	check(err)
	logger.Printf("Feiertag '%s'", feiertag.Name)

	if feiertag.Name != "" {
		logger.Printf("Found holiday: '%s'", feiertag.Name)
		fmt.Println(create_tweet_message(*feiertag))
	}

}

func create_tweet_message(f Feiertag) string {
	var tpl_output bytes.Buffer
	wikisearch_url := "https://de.wikipedia.org/w/index.php?search=" + url.QueryEscape(f.Name)
	tweet_templates := [3]string{"Übrigens: Heute ist \"{{.Name}}\"!",
		"Kleine Erinnerung: heute ist \"{{.Name}}\"!",
		"Na? Auch fast auf \"{{.Name}}\" vergessen?"}
	wiki_templates := [2]string{"(Mehr Infos dazu in der Wikipedia " + wikisearch_url+")", "(Wikipedia weiß mehr: " + wikisearch_url+")"}

	// randomisierter name, damit template neu gebaut wird
	tmpl, err := template.New(fmt.Sprintf("tweet %d", rand.Intn(100))).Parse(tweet_templates[rand.Intn(len(tweet_templates))] + " " + wiki_templates[rand.Intn(len(wiki_templates))])
	check(err)

	err = tmpl.Execute(&tpl_output, f)
	check(err)

	return tpl_output.String()

}

func main() {

	check_current_date()

}
