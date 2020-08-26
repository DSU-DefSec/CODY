package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

// Retrieve playlistItems in the specified playlist
func playlistItemsList(service *youtube.Service, part string, playlistId string, pageToken string) *youtube.PlaylistItemListResponse {
	call := service.PlaylistItems.List([]string{part})
	call = call.PlaylistId(playlistId)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	response, err := call.Do()
	if err != nil {
		panic(err)
	}
	return response
}

func retrievePlaylist(playlistId string) []lessonData {
	fmt.Println("[YT] Retrieving playlist", playlistId)

	client := &http.Client{
		Transport: &transport.APIKey{Key: codyConf.DeveloperKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("error creating new YouTube client: %v", err)
	}

	playlistData := []lessonData{}

	nextPageToken := ""
	pageCount := 1
	for {
		fmt.Println("[YT] Retrieving playlist page", pageCount)
		// Retrieve next set of items in the playlist.
		playlistResponse := playlistItemsList(service, "snippet", playlistId, nextPageToken)

		for _, playlistItem := range playlistResponse.Items {
			// videoTitle := playlistItem.Snippet.Title
			videoId := playlistItem.Snippet.ResourceId.VideoId
			fmt.Println("[YT] Processing playlist item", videoId)
			videoDesc := playlistItem.Snippet.Description
			playlistDatum := lessonData{
				Video: videoId,
			}
			parseDescription(&playlistDatum, videoDesc)
			playlistData = append(playlistData, playlistDatum)
		}

		// Set the token to retrieve the next page of results
		// or exit the loop if all results have been retrieved.
		nextPageToken = playlistResponse.NextPageToken
		if nextPageToken == "" {
			fmt.Println("[YT] Done processing playlist")
			break
		}
	}
	return playlistData
}

func parseDescription(ld *lessonData, description string) {
	fmt.Println("[YT] Input to parseDescription:", description)
	newlineSplit := strings.Split(description, "\n")
	if len(newlineSplit) < 5 {
		fmt.Println("[YT] Description is malformed.")
		return
	}
	ld.Id = newlineSplit[0]
	ld.Title = newlineSplit[1]
	ld.Description = newlineSplit[2]
	ld.VApp = ld.Id
	ld.Slides = newlineSplit[4]
	ld.PDF = newlineSplit[5]
}
