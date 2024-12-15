package hub

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type DockerHubSearchResponse struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		RepoName         string `json:"repo_name"`
		ShortDescription string `json:"short_description"`
		StarCount        int    `json:"star_count"`
		PullCount        int    `json:"pull_count"`
		RepoOwner        string `json:"repo_owner"`
		IsAutomated      bool   `json:"is_automated"`
		IsOfficial       bool   `json:"is_official"`
	} `json:"results"`
}

type Image struct {
	Architecture string  `json:"architecture"`
	Features     string  `json:"features"`
	Variant      *string `json:"variant"`
	Digest       string  `json:"digest"`
	OS           string  `json:"os"`
	OSFeatures   string  `json:"os_features"`
	OSVersion    *string `json:"os_version"`
	Size         int     `json:"size"`
	Status       string  `json:"status"`
	LastPulled   string  `json:"last_pulled"`
	LastPushed   string  `json:"last_pushed"`
}

type Tag struct {
	Creator             int     `json:"creator"`
	ID                  int     `json:"id"`
	Images              []Image `json:"images"`
	LastUpdated         string  `json:"last_updated"`
	LastUpdater         int     `json:"last_updater"`
	LastUpdaterUsername string  `json:"last_updater_username"`
	Name                string  `json:"name"`
	Repository          int     `json:"repository"`
	FullSize            int     `json:"full_size"`
	V2                  bool    `json:"v2"`
	TagStatus           string  `json:"tag_status"`
	TagLastPulled       string  `json:"tag_last_pulled"`
	TagLastPushed       string  `json:"tag_last_pushed"`
	MediaType           string  `json:"media_type"`
	ContentType         string  `json:"content_type"`
	Digest              string  `json:"digest"`
}

type DockerHubTagsResponse struct {
	Count    int     `json:"count"`
	Next     string  `json:"next"`
	Previous *string `json:"previous"`
	Results  []Tag   `json:"results"`
}

var ErrTooManyRequests = errors.New("error to many requests")

func SearchDockerHubImages(query string) ([]string, error) {
	var (
		baseURL        = "https://hub.docker.com/v2/search/repositories/"
		params         = url.Values{}
		names          = []string{}
		searchResponse DockerHubSearchResponse
	)
	params.Add("query", query)

	resp, err := http.Get(fmt.Sprintf("%s?%s", baseURL, params.Encode()))
	if err != nil {
		return []string{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return []string{}, ErrTooManyRequests
	}

	if resp.StatusCode != http.StatusOK {
		return []string{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return []string{}, fmt.Errorf("failed to decode response: %w", err)
	}

	for _, result := range searchResponse.Results {
		names = append(names, result.RepoName)
	}

	return names, nil
}

func SearchDockerHubImageTags(name string, owner ...string) ([]string, error) {
	if len(owner) == 0 {
		owner = append(owner, "library")
	}
	var (
		tags     []string
		tagsResp = &DockerHubTagsResponse{}
	)
	tagsResp.Next = fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags/", owner[0], name)

	for range 3 {
		_tagsResp, err := searchTagRequest(tagsResp.Next)
		if err != nil {
			return []string{}, err
		}

		for _, tag := range _tagsResp.Results {
			tags = append(tags, tag.Name)
		}
		if _tagsResp.Next == "" {
			break
		}
		tagsResp = _tagsResp
	}

	return tags, nil
}

func searchTagRequest(url string) (*DockerHubTagsResponse, error) {
	var tagsResp = &DockerHubTagsResponse{}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return tagsResp, nil
}
