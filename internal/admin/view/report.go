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

func (v *View) handleIdTask(args []string, taskFunc func(string) error) error {
	if len(args) < 2 {
		return fmt.Errorf("set task id")
	}
	id, ok := v.indexIdMap[args[1]]
	if !ok {
		return fmt.Errorf("invalid task index")
	}
	return taskFunc(id)
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
		go v.setTaskComplete(tasks.Tasks)
		fmt.Print(FormatTable(tasks.Tasks))
	case show:
		if err := v.handleIdTask(args, func(id string) error {
			resp, err := v.admin.Client.GetReport(id)
			if err != nil {
				return err
			}
			fmt.Print(FormatTaskResponse(resp))
			return nil
		}); err != nil {
			ErrorFunc(err)
		}
	case json, sbom:
		reportType := map[string]jamel.ReportType{
			json: jamel.ReportType_JSON,
			sbom: jamel.ReportType_SBOM_R,
		}[args[0]]

		if err := v.handleIdTask(args, func(id string) error {
			file, err := v.admin.Client.GetFile(id, reportType)
			if err != nil {
				return err
			}
			fmt.Printf("\r⬅️ %s - downloaded\n", file)
			return nil
		}); err != nil {
			ErrorFunc(err)
		}
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
