package view

import (
	"fmt"
	"slices"
	"strings"

	"jamel/gen/go/jamel"

	"jamel/pkg/fs"
	"jamel/pkg/hub"

	"github.com/c-bata/go-prompt"
)

const (
	docker        = "docker"
	dockerArchive = "archive-docker"
	file          = "file"
	sbom          = "sbom"
)

var (
	analyzeMap = map[string]jamel.TaskType{
		docker:        jamel.TaskType_DOCKER,
		dockerArchive: jamel.TaskType_DOCKER_ARCHIVE,
		sbom:          jamel.TaskType_SBOM,
		file:          jamel.TaskType_FILE,
	}
	analyzeList = analyzeCmds()
)

func analyzeCmds() []string {
	var commands = []string{}
	for cmd := range analyzeMap {
		commands = append(commands, cmd)
	}
	return commands
}

func (v *View) analyzeAction(cmd string, filename string) {
	var (
		resp *jamel.TaskResponse
		err  error
	)
	switch analyzeMap[cmd] {
	case jamel.TaskType_DOCKER:
		resp, err = v.admin.Client.TaskFromImage(filename)
	default:
		resp, err = v.admin.Client.TaskFromFile(
			filename, analyzeMap[cmd],
		)
	}
	if err != nil {
		ErrorFunc(err)
		return
	}
	fmt.Print(FormatTaskResponse(resp))
}

func (v *View) analyzeExecutor(in string) {
	args := strings.Fields(in)
	if len(args) == 0 {
		return
	}

	if slices.Contains(analyzeList, args[0]) {
		if len(args) < 2 {
			ErrorFunc(fmt.Errorf("you must set file, dir or docker image name"))
			return
		}
		go v.analyzeAction(
			args[0],
			args[1],
		)
		return
	}

	switch args[0] {
	case exit:
		return
	default:
		NotFoundFunc()
	}
}

func (v *View) analyzeCompleter(d prompt.Document) []prompt.Suggest {
	var complete = []prompt.Suggest{
		{Text: docker, Description: "image from docker.hub"},
		{Text: dockerArchive, Description: "image from local tar archive"},
		{Text: file, Description: "file or dir on disk"},
		{Text: sbom, Description: "json sbom file"},
		{Text: exit, Description: "close"},
	}

	for _, c := range complete {
		if HasPrefix(d, c.Text) {
			complete = []prompt.Suggest{}
		}
	}

	if HasPrefix(d, dockerArchive) {
		complete = ListToSuggest(fs.MustFilesInDot("tar", "zip"))
	}
	if HasPrefix(d, docker) {
		go v.dockerSuggestion(d.GetWordBeforeCursor())
		complete = ListToSuggest(v.dockerComplete)
	}
	if HasPrefix(d, file) {
		complete = ListToSuggest(fs.MustEntitiesInDot())
	}
	if HasPrefix(d, sbom) {
		complete = ListToSuggest(fs.MustFilesInDot("json"))
	}

	return prompt.FilterContains(
		complete,
		d.GetWordBeforeCursor(),
		true,
	)
}

func (v *View) dockerSuggestion(query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		return
	}

	var (
		parts = strings.SplitN(query, ":", 2)
		image = parts[0]
	)

	if len(parts) > 1 {
		tags, err := fetchTags(image)
		if err != nil {
			return
		}

		suggest := make([]string, len(tags))
		for i, t := range tags {
			suggest[i] = fmt.Sprintf("%s:%s", image, t)
		}
		v.dockerComplete = suggest
		return
	}

	// Query does not contain a colon
	images, err := hub.SearchDockerHubImages(image)
	if err != nil {
		return
	}
	v.dockerComplete = images
}

func fetchTags(image string) ([]string, error) {
	var parts = strings.Split(image, "/")

	switch len(parts) {
	case 1:
		return hub.SearchDockerHubImageTags(parts[0])
	case 2:
		return hub.SearchDockerHubImageTags(parts[1], parts[0])
	default:
		return nil, nil
	}
}
