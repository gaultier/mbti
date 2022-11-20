package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	ApiUrl = "https://api.themoviedb.org/3"
)

type Series struct {
	Id          uint64
	Name        string
	Overview    string
	VoteAverage float32 `json:"vote_average"`
}

type Response struct {
	Page    uint64
	Results []Series
}

func main() {
	fmt.Println("vim-go")
	client := http.Client{Timeout: 10 * time.Second}

	apiKey := os.Getenv("API_KEY")
	page := 1
	query := "Friends" // FIXME
	res, err := client.Get(fmt.Sprintf("%s/search/tv?page=%d&api_key=%s&query=%s", ApiUrl, page, apiKey, url.QueryEscape(query)))
	if err != nil {
		panic(err) // TODO: improve
	}

	if res.StatusCode != 200 {
		log.Panicf("Non 200 response: %s %v", res.Status, res.Body)
	}

	response := Response{Results: make([]Series, 0, 500)}
	j := json.NewDecoder(res.Body)
	err = j.Decode(&response)
	if err != nil {
		panic(err) // TODO: improve
	}

	log.Println(response.Results[0:5])

}
