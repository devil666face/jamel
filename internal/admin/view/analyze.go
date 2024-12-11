package view

import (
	"jamel/gen/go/jamel"
	"slices"
	"strings"

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
	for cmd, _ := range analyzeMap {
		commands = append(commands, cmd)
	}
	return commands
}

func (v *View) analyzeAction(cmd string, filename string) {
	v.admin.NewTaskFromFile(filename, analyzeMap[cmd])
}

func (v *View) analyzeExecutor(in string) {
	args := strings.Fields(in)
	if len(args) == 0 {
		return
	}

	if slices.Contains(analyzeList, args[0]) {
		v.analyzeAction(args[0], args[1])
		return
	}

	NotFoundFunc()
}

func (v *View) analyzeCompleter(d prompt.Document) []prompt.Suggest {
	var complete = []prompt.Suggest{
		{Text: DockerArchive, Description: "docker image from your local .tar archive"},
		{Text: Docker, Description: "docker image from public registry"},
		{Text: Dir, Description: "your local folder"},
		{Text: File, Description: ".tar or .zip archive or file with requirements (go.mod,requirement.txt,modules.json...)"},
		{Text: Sbom, Description: "sbom file in json"},
	}
	// remove second postition complete
	for _, c := range complete {
		if HasPrefix(d, c.Text) {
			complete = []prompt.Suggest{}
		}
	}

	return prompt.FilterContains(complete, d.GetWordBeforeCursor(), true)
}
