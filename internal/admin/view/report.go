package view

import (
	"fmt"
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

func reportCmds() []string {
	return []string{
		list,
		pdf,
		json,
		show,
		sbom,
	}
}

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
		fmt.Print(FormatTable(tasks.Tasks))
	case show:
		if len(args) < 2 {
			ErrorFunc(fmt.Errorf("set task id"))
			return
		}
		resp, err := v.admin.Client.GetReport(args[1])
		if err != nil {
			ErrorFunc(err)
			return
		}
		fmt.Print(FormatTaskResponse(resp))
	case json:
	case sbom:
	case pdf:
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
		{Text: pdf, Description: "download report for task in pdf"},
		{Text: json, Description: "download report for task in json"},
		{Text: sbom, Description: "download sbom file for task"},
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
