package view

import (
	"fmt"
	"jamel/gen/go/jamel"
	"jamel/pkg/fs"
	"strings"

	"github.com/c-bata/go-prompt"
)

const (
	list = "list"
	json = "json"
	show = "show"
	pdf  = "pdf"
)

func (v *View) reportExecutor(in string) {
	args := strings.Fields(in)
	if len(args) == 0 {
		return
	}

	switch args[0] {
	case list:
		tasks, err := v.admin.Client.TaskList()
		if err != nil {
			ErrorFunc(err)
			return
		}
		go v.setTaskComplete(tasks.Tasks)
		fmt.Print(FormatTable(tasks.Tasks))
	case show:
		if len(args) < 2 {
			ErrorFunc(fmt.Errorf("set task id"))
			return
		}
		id, ok := v.indexIdMap[args[1]]
		if !ok {
			ErrorFunc(fmt.Errorf("invalid task index"))
			return
		}
		resp, err := v.admin.Client.GetReport(id)
		if err != nil {
			ErrorFunc(err)
			return
		}
		fmt.Print(FormatTaskResponse(resp))
	case json:
		if len(args) < 2 {
			ErrorFunc(fmt.Errorf("set task id"))
			return
		}
		id, ok := v.indexIdMap[args[1]]
		if !ok {
			ErrorFunc(fmt.Errorf("invalid task index"))
			return
		}
		file, err := v.admin.Client.GetFile(id, jamel.ReportType_JSON)
		if err != nil {
			ErrorFunc(err)
			return
		}
		fmt.Printf("\r⬅️ %s - downloaded\n", file)
	case sbom:
		if len(args) < 2 {
			ErrorFunc(fmt.Errorf("set task id"))
			return
		}
		id, ok := v.indexIdMap[args[1]]
		if !ok {
			ErrorFunc(fmt.Errorf("invalid task index"))
			return
		}
		file, err := v.admin.Client.GetFile(id, jamel.ReportType_SBOM_R)
		if err != nil {
			ErrorFunc(err)
			return
		}
		fmt.Printf("\r⬅️ %s - downloaded\n", file)
	// case pdf:
	case exit:
		return
	default:
		NotFoundFunc()
	}
}

func (v *View) reportCompleter(d prompt.Document) []prompt.Suggest {
	var complete = []prompt.Suggest{
		{Text: list, Description: "show all results"},
		{Text: show, Description: "show report for task"},
		{Text: json, Description: "download report for task in json"},
		{Text: sbom, Description: "download sbom file for task"},
		// {Text: pdf, Description: "download report for task in pdf"},
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

	if HasPrefix(d, pdf) ||
		HasPrefix(d, json) ||
		HasPrefix(d, sbom) ||
		HasPrefix(d, show) {
		complete = v.taskComplete
	}

	return prompt.FilterContains(
		complete,
		d.GetWordBeforeCursor(),
		true,
	)
}
