package date

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const FormatDate string = "20060102"

func ApiNextDate(w http.ResponseWriter, r *http.Request) {
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeatStr := r.URL.Query().Get("repeat")

	now, err := time.Parse(FormatDate, nowStr)
	if err != nil {
		http.Error(w, "Некорректная дата now", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeatStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, nextDate)
}

func NextDate(now time.Time, dateStart, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("repeat value is empty")
	}

	date, err := time.Parse(FormatDate, dateStart)
	if err != nil {
		return "", err
	}

	repeatRule := strings.Split(repeat, " ")
	switch strings.ToLower(repeatRule[0]) {
	case "d":
		if len(repeatRule) < 2 {
			return "", fmt.Errorf("too few arguments passed")
		}
		days, err := strconv.Atoi(repeatRule[1])
		if days < 0 || days > 400 {
			return "", fmt.Errorf("incorrect count of days, must be between 0 and 400")
		}
		if err != nil {
			return "", err
		}
		date = addDayTask(now, date, days)
	case "y":
		date = addYearTask(now, date)
	case "w":
		if len(repeatRule) < 2 {
			return "", fmt.Errorf("too few arguments passed")
		}
		date, err = addWeekTask(now, date, repeatRule[1])
		if err != nil {
			return "", err
		}
	case "m":
		if len(repeatRule) < 2 {
			return "", fmt.Errorf("too few arguments passed")
		}
		date, err = addMonthTask(now, date, repeatRule)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("invalid repeat character")
	}

	return date.Format(FormatDate), nil
}

func addDayTask(now, date time.Time, days int) time.Time {

	if date.Equal(now) {
		return now
	}
	date = date.AddDate(0, 0, days)
	for date.Before(now) {
		date = date.AddDate(0, 0, days)
	}
	return date
}

func addYearTask(now, date time.Time) time.Time {
	date = date.AddDate(1, 0, 0)
	for date.Before(now) {
		date = date.AddDate(1, 0, 0)
	}
	return date
}

func addWeekTask(now, date time.Time, daysString string) (time.Time, error) {
	dayNumbers := strings.Split(daysString, ",")
	weekDays := make(map[int]bool)

	for _, day := range dayNumbers {
		dayInt, err := strconv.Atoi(day)
		if err != nil {
			return date, err
		}
		if dayInt < 1 || dayInt > 7 {
			return date, fmt.Errorf("invalid day of the week: %d", dayInt)
		}
		if dayInt == 7 {
			dayInt = 0
		}
		weekDays[dayInt] = true
	}

	for {
		if weekDays[int(date.Weekday())] && now.Before(date) {
			break
		}
		date = date.AddDate(0, 0, 1)
	}

	return date, nil
}

func addMonthTask(now, date time.Time, repeat []string) (time.Time, error) {
	days := strings.Split(repeat[1], ",")
	months := []string{}
	if len(repeat) > 2 {
		months = strings.Split(repeat[2], ",")
	}

	dayMap := make(map[int]bool)
	for _, day := range days {
		dayInt, err := strconv.Atoi(day)
		if err != nil {
			return date, err
		}
		if dayInt < -2 || dayInt > 31 || dayInt == 0 {
			return date, fmt.Errorf("invalid day of the month: %d", dayInt)
		}
		dayMap[dayInt] = true
	}

	monthMap := make(map[int]bool)
	for _, month := range months {
		if month == "" {
			continue
		}
		monthInt, err := strconv.Atoi(month)
		if err != nil {
			return date, err
		}
		if monthInt < 1 || monthInt > 12 {
			return date, fmt.Errorf("invalid month: %d", monthInt)
		}
		monthMap[monthInt] = true
	}

	for {
		if len(monthMap) == 0 || monthMap[int(date.Month())] {
			lastDay := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location()).Day()
			secondLastDay := lastDay - 1

			day := date.Day()
			switch {
			case day == lastDay && dayMap[-1]:
				day = -1
			case day == secondLastDay && dayMap[-2]:
				day = -2
			}

			if dayMap[day] && now.Before(date) {
				break
			}
		}
		date = date.AddDate(0, 0, 1)
	}

	return date, nil
}
