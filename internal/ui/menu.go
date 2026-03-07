package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

type MenuOption struct {
	Label string
	Value string
}

// NumberMenu shows a select menu. Returns the selected value or error on cancel.
func NumberMenu(message string, options []MenuOption) (string, error) {
	var result string

	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(fmt.Sprintf("%d) %s", i+1, opt.Label), opt.Value)
	}

	err := huh.NewSelect[string]().
		Title(message).
		Options(huhOptions...).
		Value(&result).
		Run()

	if err != nil {
		return "", err
	}
	return result, nil
}

// Confirm shows a yes/no prompt. Returns true for yes.
func Confirm(message string) (bool, error) {
	var result bool
	err := huh.NewConfirm().
		Title(message).
		Affirmative("Yes").
		Negative("No").
		Value(&result).
		Run()
	return result, err
}
