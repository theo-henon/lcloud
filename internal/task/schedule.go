package task

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

const minInterval = time.Minute

func ValidateSchedule(scheduleType, schedule string) error {
	schedule = strings.TrimSpace(schedule)
	if schedule == "" {
		return fmt.Errorf("%w: schedule is required", ErrInvalidSchedule)
	}
	switch scheduleType {
	case ScheduleTypeCron:
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		if _, err := parser.Parse(schedule); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidSchedule, err)
		}
		return nil
	case ScheduleTypeInterval:
		d, err := time.ParseDuration(schedule)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidSchedule, err)
		}
		if d < minInterval {
			return fmt.Errorf("%w: minimum interval is 1m", ErrInvalidSchedule)
		}
		return nil
	default:
		return fmt.Errorf("%w: unknown schedule type %s", ErrInvalidSchedule, scheduleType)
	}
}

func DescribeSchedule(scheduleType, schedule string) string {
	switch scheduleType {
	case ScheduleTypeCron:
		switch schedule {
		case "0 0 * * *":
			return "Daily at 00:00 UTC"
		case "0 2 * * 0":
			return "Every Sunday at 02:00 UTC"
		case "0 * * * *":
			return "Every hour at :00 UTC"
		default:
			return fmt.Sprintf("Cron: %s UTC", schedule)
		}
	case ScheduleTypeInterval:
		d, err := time.ParseDuration(schedule)
		if err != nil {
			return schedule
		}
		if d == time.Hour {
			return "Every hour"
		}
		if d == 24*time.Hour {
			return "Every 24 hours"
		}
		if d == 168*time.Hour {
			return "Every 7 days"
		}
		return fmt.Sprintf("Every %s", d.String())
	default:
		return schedule
	}
}

func NextRunAt(scheduleType, schedule string, from time.Time) (*time.Time, error) {
	if err := ValidateSchedule(scheduleType, schedule); err != nil {
		return nil, err
	}
	from = from.UTC()
	switch scheduleType {
	case ScheduleTypeCron:
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		sched, err := parser.Parse(schedule)
		if err != nil {
			return nil, err
		}
		next := sched.Next(from)
		return &next, nil
	case ScheduleTypeInterval:
		d, _ := time.ParseDuration(schedule)
		next := from.Add(d)
		return &next, nil
	default:
		return nil, ErrInvalidSchedule
	}
}
