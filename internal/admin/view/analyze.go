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

var (
	analyzeMap = map[string]jamel.TaskType{
		Docker:        jamel.TaskType_DOCKER,
		DockerArchive: jamel.TaskType_DOCKER_ARCHIVE,
		Dir:           jamel.TaskType_DIR,
		File:          jamel.TaskType_DIR,
		Sbom:          jamel.TaskType_SBOM,
	}
	analyzeList = analyzeCommands()
)

func analyzeCommands() []string {
	var commands = []string{}
	for cmd := range analyzeMap {
		commands = append(commands, cmd)
	}
	return commands
}

func (v *View) analyzeAction(cmd string, filename string) {
	var (
		out string
		err error
	)
	switch analyzeMap[cmd] {
	case jamel.TaskType_DOCKER:
		out, err = v.admin.NewTaskFromImage(filename)
	default:
		out, err = v.admin.NewTaskFromFile(filename, analyzeMap[cmd])
	}
	if err != nil {
		ErrorFunc(err)
		return
	}
	fmt.Print(out)
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

	NotFoundFunc()
}

func (v *View) analyzeCompleter(d prompt.Document) []prompt.Suggest {
	var complete = []prompt.Suggest{
		{Text: DockerArchive, Description: "docker image from your local .tar archive"},
		{Text: Docker, Description: "docker image from public registry"},
		{Text: Dir, Description: "your local folder"},
		{Text: File, Description: ".tar or .zip archive or file with requirements (go.mod, requirement.txt, modules.json...)"},
		{Text: Sbom, Description: "sbom file in json"},
	}

	for _, c := range complete {
		if HasPrefix(d, c.Text) {
			complete = []prompt.Suggest{}
		}
	}

	if HasPrefix(d, DockerArchive) {
		complete = ListToSuggest(fs.MustFilesInDot("tar", "zip"))
	}

	if HasPrefix(d, Docker) {
		go v.dockerSuggestion(d.GetWordBeforeCursor())
		complete = ListToSuggest(v.dockerComplete)
	}

	if HasPrefix(d, File) {
		complete = ListToSuggest(fs.MustFilesInDot())
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
