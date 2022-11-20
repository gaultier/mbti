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

	"github.com/pterm/pterm"
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

	var search string
	{
		search, _ = pterm.DefaultInteractiveTextInput.WithDefaultText("What do you want to watch?").WithMultiLine(false).Show()
		pterm.Println() // Blank line
	}

	// Search shows.
	var pickedShow *ShowSummary
	{
		page := 1
		url := fmt.Sprintf("%s/search/tv?page=%d&api_key=%s&query=%s", ApiUrl, page, apiKey, url.QueryEscape(search))
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

		{
			options := make([]string, len(response.Results))
			for i, show := range response.Results {
				options[i] = show.Name
			}
			selectedOption, _ := pterm.DefaultInteractiveSelect.WithOptions(options).Show()
			pterm.Info.Printfln("Selected show: %s", pterm.Green(selectedOption))

			for i, show := range response.Results {
				if show.Name == selectedOption {
					pickedShow = &response.Results[i]
				}
			}
		}
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

		{
			options := make([]string, len(show.Seasons))
			for i, season := range show.Seasons {
				options[i] = season.Name
			}
			selectedOption, _ := pterm.DefaultInteractiveSelect.WithOptions(options).Show()
			pterm.Info.Printfln("Selected season: %s", pterm.Green(selectedOption))

			for i, season := range show.Seasons {
				if season.Name == selectedOption {
					pickedSeason = &show.Seasons[i]
				}
			}
		}
	}

	// Get season's episodes
	var pickedEpisode *EpisodeSummary
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

		{
			options := make([]string, len(season.Episodes))
			for i, episode := range season.Episodes {
				options[i] = episode.Name
			}
			selectedOption, _ := pterm.DefaultInteractiveSelect.WithOptions(options).Show()
			pterm.Info.Printfln("Selected show: %s", pterm.Green(selectedOption))

			for i, episode := range season.Episodes {
				if episode.Name == selectedOption {
					pickedEpisode = &season.Episodes[i]
				}
			}
		}
	}
	pterm.DefaultCenter.Println("Now watching:\n")
	s, _ := pterm.DefaultBigText.WithLetters(pterm.NewLettersFromString(pickedEpisode.Name)).Srender()
	pterm.DefaultCenter.Println(s) // Print BigLetters with the default CenterPrinter
}
