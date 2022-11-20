package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	ApiUrl = "https://api.themoviedb.org/3"
)

type ShowSummary struct {
	Id          uint64
	Name        string
	Overview    string
	VoteAverage float32 `json:"vote_average"`
}

type SearchSeriesResponse struct {
	Page    uint64
	Results []ShowSummary
}

type SeasonSummary struct {
	Id           uint64
	Name         string
	Overview     string
	SeasonNumber uint64 `json:"season_number"`
}

type ShowFull struct {
	Seasons []SeasonSummary
}

type EpisodeSummary struct {
	Id            uint64
	Name          string
	Overview      string
	EpisodeNumber uint64  `json:"episode_number"`
	VoteAverage   float32 `json:"vote_average"`
}

type SeasonFull struct {
	Episodes []EpisodeSummary
}

func main() {
	fmt.Println("vim-go")
	client := http.Client{Timeout: 10 * time.Second}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatalf("No API_KEY environment variable set")
	}

	// Search shows.
	var pickedShow *ShowSummary
	{
		page := 1
		query := "Friends" // FIXME
		url := fmt.Sprintf("%s/search/tv?page=%d&api_key=%s&query=%s", ApiUrl, page, apiKey, url.QueryEscape(query))
		res, err := client.Get(url)
		if err != nil {
			panic(err) // FIXME
		}

		if res.StatusCode != 200 {
			body, _ := io.ReadAll(res.Body)
			log.Panicf("Non 200 response: url=%s status=%s body=%s", url, res.Status, body)
		}

		response := SearchSeriesResponse{Results: make([]ShowSummary, 0, 500)}
		j := json.NewDecoder(res.Body)
		err = j.Decode(&response)
		if err != nil {
			panic(err) // FIXME
		}

		log.Println(response.Results[0:5])
		pickedShow = &response.Results[0] // FIXME
	}

	// Get show
	var pickedSeason *SeasonSummary
	{
		url := fmt.Sprintf("%s/tv/%d?api_key=%s", ApiUrl, pickedShow.Id, apiKey)
		res, err := client.Get(url)
		if err != nil {
			panic(err) // FIXME
		}

		if res.StatusCode != 200 {
			body, _ := io.ReadAll(res.Body)
			log.Panicf("Non 200 response: url=%s status=%s body=%s", url, res.Status, body)
		}

		show := ShowFull{}
		j := json.NewDecoder(res.Body)
		err = j.Decode(&show)
		if err != nil {
			panic(err) // FIXME
		}

		log.Println(show)
		pickedSeason = &show.Seasons[1] // FIXME
	}

	// Get season
	{
		url := fmt.Sprintf("%s/tv/%d/season/%d?api_key=%s", ApiUrl, pickedShow.Id, pickedSeason.SeasonNumber, apiKey)
		res, err := client.Get(url)
		if err != nil {
			panic(err) // FIXME
		}

		if res.StatusCode != 200 {
			body, _ := io.ReadAll(res.Body)
			log.Panicf("Non 200 response: url=%s status=%s body=%s", url, res.Status, body)
		}

		season := SeasonFull{}
		j := json.NewDecoder(res.Body)
		err = j.Decode(&season)
		if err != nil {
			panic(err) // FIXME
		}

		log.Println(season)
	}
}
