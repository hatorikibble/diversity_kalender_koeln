package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"text/template"
	"time"
)

type Configuration struct {
	Logfile                     string
	Twitter_access_token        string
	Twitter_access_token_secret string
	Twitter_consumer_key        string
	Twitter_consumer_secret     string
	Api_server                  string
	Debug                       int
}

type Feiertag struct {
	Date        string
	Name        string
	Type        string
	Description string
}

type Status struct {
	Status string `json:"status"`
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

	// check webservice
	check_webservice()

}

// check_webservice checks if webservice is up
func check_webservice() {
	query_url := fmt.Sprintf("%s/health", configuration.Api_server)

	logger.Printf("Querying %s", query_url)
	response, err := http.Get(query_url)
	check(err)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	check(err)
	status := new(Status)
	err = json.Unmarshal(body, status)
	check(err)

	logger.Printf("Webservice at %s is '%s'", configuration.Api_server, status.Status)
	if status.Status != "UP" {
		panic(fmt.Sprintf("Webservice at %s is %s", configuration.Api_server, status.Status))
	}

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
	if configuration.Debug == 1 {
		date_string = "14.04.2018"
		logger.Printf("DEBUG: set date_string to  %s", date_string)
	}
	logger.Printf("Today is %s", date_string)
	query_url := fmt.Sprintf("%s/holiday/%s", configuration.Api_server, date_string)

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
		tweet(create_tweet_message(*feiertag))
	}

}

func create_tweet_message(f Feiertag) string {
	var tpl_output bytes.Buffer
	tweet_templates := [3]string{"Übrigens: Heute ist \"{{.Name}}\"! {{.Description}}",
		"Kleine Erinnerung: heute ist \"{{.Name}}\"! {{.Description}}",
		"Na? Auch fast auf \"{{.Name}}\" vergessen? {{.Description}}"}

	// randomisierter name, damit template neu gebaut wird
	tmpl, err := template.New(fmt.Sprintf("tweet %d", rand.Intn(100))).Parse(tweet_templates[rand.Intn(len(tweet_templates))])
	check(err)

	err = tmpl.Execute(&tpl_output, f)
	check(err)

	return tpl_output.String()

}

func tweet(tweet_text string) {
	// twitter api
	anaconda.SetConsumerKey(configuration.Twitter_consumer_key)
	anaconda.SetConsumerSecret(configuration.Twitter_consumer_secret)
	// I don't know about any possible timeout, therefore
	// initialize new for every tweet
	api := anaconda.NewTwitterApi(configuration.Twitter_access_token, configuration.Twitter_access_token_secret)

	// is the tweet too long -> truncate
	if len(tweet_text) > 280 {
		tweet_text = tweet_text[0:275] + "..."

	}

	if configuration.Debug == 1 {
		logger.Printf("DEBUG-MODE! I am not posting '%s'!", tweet_text)
	} else {
		tweet, err := api.PostTweet(tweet_text, nil)
		if err != nil {
			logger.Printf("Problem posting '%s': %s", tweet_text, err)
		} else {
			logger.Printf("Tweet '%s' posted for user %s", tweet_text, tweet.User.ScreenName)
		}
	}
}

func main() {

	check_current_date()

}
