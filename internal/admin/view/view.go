package view

import (
	"fmt"
	"jamel/gen/go/jamel"
	"jamel/internal/admin"
	"os"
	"text/tabwriter"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/charmbracelet/lipgloss"

	"strings"
)

const dateFormat = "Jan 02 15:04"

const (
	emoji    = "🐝"
	notFound = "⚠️ command not found"
)

const (
	analyze = "analyze"
	report  = "report"
	exit    = "exit"
)

func ErrorFunc(err error, pre ...string) {
	if len(pre) == 0 {
		pre = append(pre, "")
	}
	fmt.Printf("\r%s⚠️ %s\r\n", pre[0], err)
}

var NotFoundFunc func() = func() {
	fmt.Println(notFound)
}

var (
	title = fmt.Sprintf("\r%s # ", emoji)
)

type View struct {
	prompt         *prompt.Prompt
	admin          *admin.Admin
	dockerComplete []string
}

func New(_admin *admin.Admin) *View {
	_view := &View{
		admin: _admin,
	}
	_view.prompt = NewPrompt(
		_view.executor,
		_view.completer,
		title,
	)
	return _view
}

func (v *View) Run() {
	v.prompt.Run()
}

func (v *View) executor(in string) {
	args := strings.Fields(in)
	if len(args) == 0 {
		return
	}

	switch args[0] {
	case analyze:
		NewPrompt(
			v.analyzeExecutor,
			v.analyzeCompleter,
			fmt.Sprintf("%s%s %s # ", title, "⚙️", analyze),
		).Run()
	case report:
		NewPrompt(
			v.reportExecutor,
			v.reportCompleter,
			fmt.Sprintf("%s%s %s # ", title, "📒", report),
		).Run()
	case exit:
		os.Exit(0)
	default:
		NotFoundFunc()
	}
}

func (v *View) completer(d prompt.Document) []prompt.Suggest {
	complete := []prompt.Suggest{
		{Text: analyze, Description: "new task for analyze"},
		{Text: report, Description: "show or download reports"},
		{Text: exit, Description: "close"},
	}
	// Remove second postition complete
	for _, c := range complete {
		if HasPrefix(d, c.Text) {
			complete = []prompt.Suggest{}
		}
	}
	return prompt.FilterContains(complete, d.GetWordBeforeCursor(), true)
}

func NewPrompt(
	executor func(string),
	completer func(prompt.Document) []prompt.Suggest,
	title string,
) *prompt.Prompt {
	return prompt.New(
		executor,
		completer,
		PromptOptions(
			title,
			exit,
		)...,
	)
}

func HasPrefix(d prompt.Document, prefix string) bool {
	return strings.HasPrefix(strings.TrimSpace(d.TextBeforeCursor()), prefix)
}

func PromptOptions(prefix string, exit string) []prompt.Option {
	options := []prompt.Option{
		prompt.OptionPrefix(prefix),
		prompt.OptionTitle(prefix),
		prompt.OptionPrefixTextColor(0),
		prompt.OptionPrefixTextColor(prompt.DefaultColor),
		prompt.OptionPrefixBackgroundColor(prompt.DefaultColor),
		prompt.OptionInputTextColor(prompt.DefaultColor),
		prompt.OptionInputBGColor(prompt.DefaultColor),
		prompt.OptionPreviewSuggestionTextColor(prompt.DefaultColor),
		prompt.OptionPreviewSuggestionBGColor(prompt.DefaultColor),
		prompt.OptionSuggestionTextColor(prompt.DefaultColor),
		prompt.OptionSuggestionBGColor(prompt.DefaultColor),
		prompt.OptionSelectedSuggestionTextColor(prompt.Green),
		prompt.OptionSelectedSuggestionBGColor(prompt.DefaultColor),
		prompt.OptionDescriptionTextColor(prompt.DefaultColor),
		prompt.OptionDescriptionBGColor(prompt.DefaultColor),
		prompt.OptionSelectedDescriptionTextColor(prompt.Green),
		prompt.OptionSelectedDescriptionBGColor(prompt.DefaultColor),
		prompt.OptionScrollbarThumbColor(prompt.DefaultColor),
		prompt.OptionScrollbarBGColor(prompt.DefaultColor),
		prompt.OptionMaxSuggestion(3),
		prompt.OptionSetExitCheckerOnInput(prompt.ExitChecker(func(in string, breakline bool) bool {
			return strings.TrimSpace(in) == exit
		})),
	}
	return options
}

func ListToSuggest(list []string) []prompt.Suggest {
	suggest := []prompt.Suggest{}
	for _, l := range list {
		suggest = append(suggest, prompt.Suggest{Text: l, Description: ""})
	}
	return suggest
}

func FormatTaskResponse(resp *jamel.TaskResponse) string {
	var (
		sb strings.Builder
		w  = tabwriter.NewWriter(&sb, 1, 1, 1, ' ', 0)
	)
	sb.WriteString("\r\n")
	fmt.Fprintf(w,
		"%s\t%s\t%s\t%s",
		"Id", "Created", "Filename", "Type",
	)
	fmt.Fprintf(w,
		"\n%s\t%s\t%s\t%s",
		resp.TaskId,
		respTime(resp.CreatedAt),
		resp.Name,
		resp.TaskType,
	)
	w.Flush()
	sb.WriteString("\n")
	sb.WriteString("\n" + resp.Report)
	return lipgloss.NewStyle().
		Padding(0, 1).
		Render(sb.String())
}

func FormatTable(tasks []*jamel.TaskResponse) string {
	var (
		sb strings.Builder
		w  = tabwriter.NewWriter(&sb, 1, 1, 1, ' ', 0)
	)
	sb.WriteString("\r\n")
	fmt.Fprintf(w,
		"#\t%s\t%s\t%s\t%s",
		"Id", "Created", "Filename", "Type",
	)
	for i, task := range tasks {
		fmt.Fprintf(w,
			"\n%d\t%s\t%s\t%s\t%s",
			i,
			task.TaskId,
			respTime(task.CreatedAt),
			task.Name,
			task.TaskType,
		)
	}
	w.Flush()
	sb.WriteString("\n")
	return lipgloss.NewStyle().
		Padding(0, 1).
		Render(sb.String())
}

func respTime(t int64) string {
	return time.Unix(t, 0).Format(dateFormat)
}
