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
	Poll Poll
}

type StoredPoll struct {
	Poll
	CronExpr  string
	CreatedAt time.Time
}

func ParsePoll(input string) (Poll, error) {
	command, err := ParsePollCommand(input, nil)
	return command.Poll, err
}

func ParsePollCommand(input string, fixedOptions []string) (PollCommand, error) {
	input = strings.TrimSpace(input)

	flags, remain, _ := parseFlags(input)
	title, options, err := parseTitleAndOptions(strings.TrimSpace(remain), fixedOptions)
	if err != nil {
		return PollCommand{}, err
	}

	return PollCommand{
		Poll: Poll{
			Title:   title,
			Options: options,
			Flags:   flags,
		},
	}, nil
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

func parseTitleAndOptions(input string, fixedOptions []string) (string, []string, error) {
	lines := strings.Split(input, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		return "", nil, errors.New("Title can't be empty")
	}

	title := strings.TrimSpace(lines[0])
	if len(fixedOptions) > 0 {
		return title, fixedOptions, nil
	}

	options := make([]string, 0, len(lines)-1)
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		options = append(options, line)
	}

	return title, options, nil
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
