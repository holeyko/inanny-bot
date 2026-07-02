package poll

import (
	"errors"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

type Flag string

const (
	Anonymous Flag = "ano"
	Multipoll Flag = "mul"
	Pin       Flag = "pin"
	Remove    Flag = "rm"
	Cron      Flag = "cron"

	titleDelimeter  = "\n"
	flagsDelimeter  = ","
	optionDelimeter = "\n"
	openFlagSymb    = "["
	closeFlagSymb   = "]"
)

type Poll struct {
	ID      int64
	ChatID  int64
	Command string
	Title   string
	Options []string
	Flags   []Flag
}

type PollCommand struct {
	Poll     Poll
	CronExpr string
}

type StoredPoll struct {
	Poll
	CronExpr  string
	CreatedAt time.Time
}

func ParsePoll(input string) (Poll, error) {
	command, err := ParsePollCommand(input)
	return command.Poll, err
}

func ParsePollCommand(input string) (PollCommand, error) {
	input = strings.TrimSpace(input)

	flags, remain, _ := parseFlags(input)
	cronExpr, remain, err := parseCronExpression(strings.TrimSpace(remain))
	if err != nil {
		return PollCommand{}, err
	}

	title, remain, err := parseTitle(remain)
	if err != nil {
		return PollCommand{}, err
	}

	options, _, _ := parseOptions(remain)

	return PollCommand{
		Poll: Poll{
			Title:   title,
			Options: options,
			Flags:   flags,
		},
		CronExpr: cronExpr,
	}, nil
}

func parseCronExpression(input string) (string, string, error) {
	if len(input) == 0 || string(input[0]) != "{" {
		return "", input, nil
	}

	closeIndex := strings.Index(input, "}")
	if closeIndex == -1 {
		return "", input, errors.New("Cron expression should be closed with }")
	}

	cronExpr := strings.TrimSpace(input[1:closeIndex])
	if cronExpr == "" {
		return "", input, errors.New("Cron expression can't be empty")
	}

	if err := ValidateCronExpr(cronExpr); err != nil {
		return "", input, err
	}

	return cronExpr, strings.TrimSpace(input[closeIndex+1:]), nil
}

func ValidateCronExpr(cronExpr string) error {
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return err
	}

	now := time.Now()
	next := schedule.Next(now)
	afterNext := schedule.Next(next)
	if afterNext.Sub(next) < time.Minute {
		return errors.New("Min cron timer is one minute")
	}

	return nil
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
		flag = strings.TrimSpace(flag)
		flags = append(flags, Flag(flag))
	}

	return flags, input[closeFlagIndex+1:], nil
}

func parseTitle(input string) (string, string, error) {
	titleDelimeterIndex := strings.Index(input, titleDelimeter)
	if titleDelimeterIndex == -1 {
		return input, "", nil
	}

	title := input[:titleDelimeterIndex]
	if title == "" {
		return "", "", errors.New("Title can't be empty")
	}

	return title, input[titleDelimeterIndex+1:], nil
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

func FlagsToStrings(flags []Flag) []string {
	result := make([]string, len(flags))
	for i, flag := range flags {
		result[i] = string(flag)
	}
	return result
}

func StringsToFlags(flags []string) []Flag {
	result := make([]Flag, len(flags))
	for i, flag := range flags {
		result[i] = Flag(flag)
	}
	return result
}
