package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"time"
)

// Event blah blah blah
type Event struct {
	EventName string `json:"event_name"`
	EventID   uint64 `json:"event_id"`
	Time      int64  `json:"time"`
	EventURL  string `json:"event_url"`
}

// Group blah blah blah
type Group struct {
	GroupCity    string `json:"group_city"`
	GroupCountry string `json:"group_country"`
	GroupID      uint64 `json:"group_id"`
}

// Meetup blah blah blah
type Meetup struct {
	ResvpID  uint64 `json:"rsvp_id"`
	Mtime    int64  `json:"mtime"`
	TheEvent Event  `json:"event"`
	TheGroup Group  `json:"group"`
}

var eventList = make([]Meetup, 1)
var futureDate int64 = 0
var futureURL string = ""
var countryMap map[string]int = make(map[string]int)
var ticker *time.Ticker

func handleJSON(jsonStr <-chan string, doneChannel chan<- bool) {
	looping := true

	for looping {
		select {
		case msg := <-jsonStr:

			// fmt.Printf(msg)

			var res Meetup

			json.Unmarshal([]byte(msg), &res)

			fmt.Printf("\n+++\nJson: %+v\n", res)

			// t := time.Unix(0, res.Mtime*1000000)
			// fmt.Printf("Mtime: %s\n", t.Local().UTC())

			// t = time.Unix(0, res.TheEvent.Time*1000000)
			// fmt.Printf("Event time: %s\n", t.Local().UTC())

			if res.TheEvent.Time > futureDate {
				futureDate = res.TheEvent.Time
				futureURL = res.TheEvent.EventURL

			}

			countryMap[res.TheGroup.GroupCountry]++

			// fmt.Printf("Country: %s. Count: %d\n", res.TheGroup.GroupCountry, countryMap[res.TheGroup.GroupCountry])
			// fmt.Printf("Map size: %d\n", len(countryMap))

			eventList = append(eventList, res)
		case t := <-ticker.C:
			fmt.Printf("Timer fired, so end of looping: %+v\n", t)
			ticker.Stop()
			looping = false
		}
	}

	doneChannel <- true
}

// Entry blah blah blah
type Entry struct {
	key   string
	value int
}

type byValue []Entry

func (v byValue) Len() int {
	return len(v)
}

func (v byValue) Less(i, j int) bool {
	return v[i].value < v[j].value
}

func (v byValue) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func main() {

	jsonChannel := make(chan string, 1000)
	doneChannel := make(chan bool)

	ticker = time.NewTicker(120 * time.Second)

	// thread for handling json
	go handleJSON(jsonChannel, doneChannel)

	timeout := time.Duration(10 * time.Second)
	client := &http.Client{
		// CheckRedirect: redirectPolicyFunc,
		Timeout: timeout,
	}

	req, _ := http.NewRequest("GET", "http://stream.meetup.com/2/rsvps", nil)

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept-Encoding", "utf-8")

	// res, err := client.Get("http://stream.meetup.com/2/rsvps")
	res, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	bytesRead := 0

	buffer := make([]byte, 16384)

	myLooping := true

	for myLooping {
		n, err := res.Body.Read(buffer)

		bytesRead += n

		if err == io.EOF {
			myLooping = false
		}

		if n != 0 {
			str := string(buffer[:n])

			jsonChannel <- str
		}

		select {
		case <-doneChannel:
			myLooping = false
		default:
			// fmt.Println("Default...")
		}

		time.Sleep(450 * time.Millisecond)
	}
	defer close(jsonChannel)
	defer close(doneChannel)

	// fmt.Printf("The list: %+v\n", eventList)

	fmt.Printf("\n\nNumber of RSVPs %d\n", len(eventList))

	theFutureDate := time.Unix(0, futureDate*1000000)
	fmt.Printf("The future date: %s\n", theFutureDate.Local().UTC())
	fmt.Printf("The future URL: %s\n", futureURL)
	fmt.Printf("The map length: %d\n", len(countryMap))

	for k, v := range countryMap {
		fmt.Printf("Country name: '%s'. Count: %d\n", k, v)
	}

	goober := make(byValue, 0, len(countryMap))

	for key, value := range countryMap {
		goober = append(goober, Entry{key, value})
	}

	sort.Sort(goober)

	for _, entry := range goober {
		fmt.Printf("%s : %v\n", entry.key, entry.value)
	}

}
