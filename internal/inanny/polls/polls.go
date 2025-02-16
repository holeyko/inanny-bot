package poll

import (
	"strings"
)

type Flag string

const (
	Anonymous Flag = "ano"
	Multipoll Flag = "mul"

	titleDelimeter  = ":"
	flagsDelimeter  = ","
	optionDelimeter = "\n"
	openFlagSymb    = "["
	closeFlagSymb   = "]"
)

type Poll struct {
	Title   string
	Options []string
	Flags   []Flag
}

func ParsePoll(input string) (Poll, error) {
	input = strings.TrimSpace(input)

	flags, remain, _ := parseFlags(input)
	title, remain, _ := parseTitle(remain)
	options, _, _ := parseOptions(remain)

	return Poll{
		Title:   title,
		Options: options,
		Flags:   flags,
	}, nil
}

func parseFlags(input string) ([]Flag, string, error) {
	flags := []Flag{}

	if len(input) == 0 {
		return flags, input, nil
	}

	if string(input[0]) != openFlagSymb {
		return flags, input, nil
	}

	closeFlagIndex := strings.Index(input, closeFlagSymb)
	titleEnd := strings.Index(input, titleDelimeter)
	if titleEnd == -1 {
		titleEnd = len(input)
	}

	if closeFlagIndex == -1 || titleEnd < closeFlagIndex {
		return flags, input, nil
	}

	flagsString := strings.Split(input[1:closeFlagIndex], flagsDelimeter)
	for _, flag := range flagsString {
		flags = append(flags, Flag(flag))
	}

	return flags, input[closeFlagIndex+1:], nil
}

func parseTitle(input string) (string, string, error) {
	titleDelimeterIndex := strings.Index(input, titleDelimeter)
	if titleDelimeterIndex == -1 {
		return input, "", nil
	}

	return input[:titleDelimeterIndex], input[titleDelimeterIndex+1:], nil
}

func parseOptions(input string) ([]string, string, error) {
	options := []string{}
	for _, option := range strings.Split(input, optionDelimeter) {
		option = strings.TrimSpace(option)
		if len(option) == 0 {
			continue
		}

		options = append(options, option)
	}

	return options, "", nil
}
