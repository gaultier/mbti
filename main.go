package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
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

func pickShow(apiKey string, client *http.Client) *ShowSummary {
	var search string
	// UI input
	{
		search, _ = pterm.DefaultInteractiveTextInput.WithDefaultText("I want to watch").WithMultiLine(false).Show()
		pterm.Println() // Blank line
	}

	// API request
	response := SearchSeriesResponse{Results: make([]ShowSummary, 0, 500)}
	{
		// No pagination
		url := fmt.Sprintf("%s/search/tv?page=1&api_key=%s&query=%s", ApiUrl, apiKey, url.QueryEscape(search))
		res, err := client.Get(url)
		if err != nil {
			panic(err) // FIXME
		}

		if res.StatusCode != 200 {
			body, _ := io.ReadAll(res.Body)
			log.Panicf("Non 200 response: url=%s status=%s body=%s", url, res.Status, body)
		}

		j := json.NewDecoder(res.Body)
		err = j.Decode(&response)
		if err != nil {
			log.Fatalf("Failed to decode response: %v", err)
		}

		if len(response.Results) == 0 {
			pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Italic)).Println("Nothing found, try something else?")
			return nil
		}
	}

	// UI select
	{
		options := make([]string, len(response.Results))
		for i, show := range response.Results {
			options[i] = fmt.Sprintf("%s (%.1f/10)", show.Name, show.VoteAverage)
		}
		selectedOption, _ := pterm.DefaultInteractiveSelect.WithOptions(options).WithMaxHeight(pterm.GetTerminalHeight() / 2).WithDefaultText("Show").Show()

		for i, option := range options {
			if option == selectedOption {
				pickedShow := &response.Results[i]
				pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Italic)).Println(strings.TrimSpace(pickedShow.Overview))
				return pickedShow
			}
		}
	}

	return nil
}

func pickSeason(showSummary *ShowSummary, apiKey string, client *http.Client) *SeasonSummary {
	show := ShowFull{}
	// API request
	{
		url := fmt.Sprintf("%s/tv/%d?api_key=%s", ApiUrl, showSummary.Id, apiKey)
		res, err := client.Get(url)
		if err != nil {
			panic(err) // FIXME
		}

		if res.StatusCode != 200 {
			body, _ := io.ReadAll(res.Body)
			log.Panicf("Non 200 response: url=%s status=%s body=%s", url, res.Status, body)
		}

		j := json.NewDecoder(res.Body)
		err = j.Decode(&show)
		if err != nil {
			log.Fatalf("Failed to decode response: %v", err)
		}
	}
	// We assume there is at least one season

	// UI select
	{
		options := make([]string, len(show.Seasons))
		for i, season := range show.Seasons {
			options[i] = season.Name
		}
		selectedOption, _ := pterm.DefaultInteractiveSelect.WithMaxHeight(pterm.GetTerminalHeight()).WithOptions(options).WithDefaultText("Season").Show()

		for i, option := range options {
			if option == selectedOption {
				pickedSeason := &show.Seasons[i]
				pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Italic)).Println(strings.TrimSpace(pickedSeason.Overview))
				return pickedSeason
			}
		}
	}
	return nil
}

func pickEpisode(pickedShow *ShowSummary, pickedSeason *SeasonSummary, apiKey string, client *http.Client) *EpisodeSummary {
	// Get season's episodes
	pterm.Println()
	{
		season := SeasonFull{}
		// API request
		{
			url := fmt.Sprintf("%s/tv/%d/season/%d?api_key=%s", ApiUrl, pickedShow.Id, pickedSeason.SeasonNumber, apiKey)
			res, err := client.Get(url)
			if err != nil {
				panic(err) // FIXME
			}

			if res.StatusCode != 200 {
				body, _ := io.ReadAll(res.Body)
				log.Fatalf("Non 200 response: url=%s status=%s body=%s", url, res.Status, body)
			}

			j := json.NewDecoder(res.Body)
			err = j.Decode(&season)
			if err != nil {
				log.Fatalf("Failed to decode response: %v", err)
			}
		}
		// Some seasons actually have no episodes (e.g. House of the Dragon season 2)
		if len(season.Episodes) == 0 {
			pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Italic)).Println("This season does not have episodes yet, check again later!")
			return nil
		}

		// UI select
		{
			options := make([]string, len(season.Episodes))
			for i, episode := range season.Episodes {
				options[i] = fmt.Sprintf("%s (%.1f/10)", episode.Name, episode.VoteAverage)
			}
			selectedOption, _ := pterm.DefaultInteractiveSelect.WithMaxHeight(pterm.GetTerminalHeight()).WithOptions(options).WithDefaultText("Episode").Show()

			for i, option := range options {
				if option == selectedOption {
					return &season.Episodes[i]
				}
			}
		}
	}

	return nil
}

func main() {
	client := http.Client{Timeout: 10 * time.Second}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatalf("No API_KEY environment variable set")
	}

	pterm.DefaultHeader.WithFullWidth(true).WithBackgroundStyle(pterm.NewStyle(pterm.BgLightMagenta)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Println("Let's watch a TV show!")
	pterm.Println()

	// Search shows
	var pickedShow *ShowSummary
	for pickedShow == nil {
		pickedShow = pickShow(apiKey, &client)
	}

	pterm.Println()
	var pickedEpisode *EpisodeSummary
	// Need this loop since some seasons do not have any episodes,
	// and we only noticde when the season has been picked already,
	// so we ask then the user to pick a different season
	for pickedEpisode == nil {
		pickedSeason := pickSeason(pickedShow, apiKey, &client)
		pickedEpisode = pickEpisode(pickedShow, pickedSeason, apiKey, &client)
	}

	// UI niceties for the end
	pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Italic)).Println(strings.TrimSpace(pickedEpisode.Overview))
	pterm.DefaultBasicText.Println()

	pterm.DefaultCenter.Println("Now watching:\n")
	s, _ := pterm.DefaultBigText.WithLetters(pterm.NewLettersFromStringWithStyle(pickedEpisode.Name, pterm.NewStyle(pterm.FgLightMagenta))).Srender()
	pterm.DefaultCenter.Println(s) // Print BigLetters with the default CenterPrinter
}
