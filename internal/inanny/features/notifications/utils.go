package notifications

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func parseInterval(interval string) (time.Duration, error) {
	var total time.Duration

	re := regexp.MustCompile(`(\d+)([Mwdh])`)
	matches := re.FindAllStringSubmatch(interval, -1)

	for _, match := range matches {
		if len(match) != 3 {
			return 0, errors.New("invalid interval")
		}

		value, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, err
		}

		unit := strings.ToLower(match[2])

		switch unit {
		case "m":
			total += time.Duration(value) * 30 * 24 * time.Hour
		case "w":
			total += time.Duration(value) * 7 * 24 * time.Hour
		case "d":
			total += time.Duration(value) * 24 * time.Hour
		case "h":
			total += time.Duration(value) * time.Hour
		}
	}

	return total, nil
}

func mustParseInterval(interval string) time.Duration {
	duration, err := parseInterval(interval)
	if err != nil {
		panic(err)
	}
	return duration
}

func durationToInterval(duration time.Duration) pgtype.Interval {
	return pgtype.Interval{
		Microseconds: duration.Microseconds(),
		Valid:        true,
	}
}

func intervalToDuration(interval pgtype.Interval) time.Duration {
	if !interval.Valid {
		return 0
	}

	return time.Duration(interval.Microseconds)*time.Microsecond +
		time.Duration(interval.Days)*24*time.Hour +
		time.Duration(interval.Months)*30*24*time.Hour
}
