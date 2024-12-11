package view

import (
	"fmt"
	"jamel/internal/admin"
	"os"

	"github.com/c-bata/go-prompt"

	"strings"
)

const (
	Emoji    = "🪣"
	NotFound = "⚠️ command not found"
)

const (
	Analyze = "analyze"
	Report  = "report"
	Exit    = "exit"
)

const (
	Docker        = "docker"
	DockerArchive = "docker-archive"
	Dir           = "dir"
	File          = "file"
	Sbom          = "sbom"
)

var ErrorFunc func(err error) = func(err error) {
	fmt.Print("⚠️ " + err.Error() + "\r\n")
}

var NotFoundFunc func() = func() {
	fmt.Println(NotFound)
}

var (
	Title = fmt.Sprintf("\r%s >> ", Emoji)
)

type View struct {
	prompt *prompt.Prompt
	admin  *admin.Admin
}

func New(_admin *admin.Admin) *View {
	_view := &View{
		admin: _admin,
	}
	_view.prompt = NewPrompt(
		_view.executor,
		_view.completer,
		Title,
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
	case Analyze:
		NewPrompt(
			v.analyzeExecutor,
			v.analyzeCompleter,
			fmt.Sprintf("%s%s >> ", Title, "🗄️ "+Analyze),
		).Run()
	// case Report:
	// 	NewPrompt(
	// 		v.settingsExecutor,
	// 		v.settingsCompleter,
	// 		fmt.Sprintf("%s%s >> ", Title, "⚙️ "+Report),
	// 	).Run()
	case Exit:
		os.Exit(0)
	default:
		NotFoundFunc()
	}
}

func (v *View) completer(d prompt.Document) []prompt.Suggest {
	complete := []prompt.Suggest{
		{Text: Analyze, Description: "new task for analyze"},
		{Text: Report, Description: "show or download reports"},
		{Text: Exit, Description: "close"},
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
			Exit,
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
