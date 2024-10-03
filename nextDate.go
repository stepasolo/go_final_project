package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	dateFormat := "20060102"
	//debagdata := time.Now()
	debagdata := now

	if repeat == "y" {
		parseDate, err := time.Parse(dateFormat, date)
		if err != nil {
			fmt.Println("Ошибка парсинга времени!", err)
			return "Ошибка парсинга времени!", err
		}

		retDate := parseDate.AddDate(1, 0, 0)
		for debagdata.After(retDate) {
			retDate = retDate.AddDate(1, 0, 0)
		}
		return retDate.Format(dateFormat), err
	}

	if strings.HasPrefix(repeat, "d") {
		parseRepeat := strings.Split(repeat, " ")
		if len(parseRepeat) < 2 {
			return "Нехватает параметров репита для дней", nil
		}
		getRepeatNumber := strings.Split(repeat, " ")[1]

		days, err := strconv.Atoi(getRepeatNumber)
		if err != nil {
			return "Ошибка парсинга строки:", err
		}

		if days > 400 {
			return "Слишком большой срок", nil
		}

		startDate, err := time.Parse(dateFormat, date)
		if err != nil {
			fmt.Println("Ошибка парсинга времени!", err)
			return "Ошибка парсинга времени!", err
		}

		retDate := startDate.AddDate(0, 0, days)

		if (debagdata.Year() == startDate.Year()) && (int(debagdata.Month()) == int(startDate.Month())) && (debagdata.Day() == startDate.Day()) {
			return debagdata.Format(dateFormat), err

		}
		for debagdata.After(retDate) {
			fmt.Printf("Дата возвращаемая:%v дата текущая:%v", retDate, debagdata)
			retDate = retDate.AddDate(0, 0, days)
		}
		return retDate.Format(dateFormat), err

	}

	if strings.HasPrefix(repeat, "w") {
		getDays := strings.TrimSpace(repeat[1:])
		getDaysNumbers := strings.Split(getDays, ",")

		var weekdays []int

		for _, getDayNumber := range getDaysNumbers {
			day, err := strconv.Atoi(strings.TrimSpace(getDayNumber))
			if err != nil || day < 1 || day > 7 {
				return "", fmt.Errorf("некоретный день недели: %v", getDayNumber)
			}
			weekdays = append(weekdays, day)
		}

		nextDate := nextWeekday(debagdata, weekdays) // тут дата неправильная
		return nextDate.Format(dateFormat), nil

	}

	if strings.HasPrefix(repeat, "m") {
		parts := strings.Split(repeat, " ")
		if len(parts) < 2 {
			return "", fmt.Errorf("несоблюдение правил шаблона месяца")
		}

		daysArr := strings.Split(parts[1], ",")
		var days []int
		for _, dayStr := range daysArr {
			day, err := strconv.Atoi(strings.TrimSpace(dayStr))
			if err != nil || day < -31 || day > 31 {
				return "", fmt.Errorf("некорректный день месяца: %v", dayStr)
			}
			days = append(days, day)
		}

		var monts []int
		if len(parts) > 2 {
			monthsArr := strings.Split(parts[2], ",")
			for _, monmonthStr := range monthsArr {
				month, err := strconv.Atoi(strings.TrimSpace(monmonthStr))
				if err != nil || month < 1 || month > 12 {
					return "", fmt.Errorf("неправильный месяц: %v", monmonthStr)
				}
				monts = append(monts, month)
			}
		}

		nextDate := nextMonthDay(now, days, monts)
		return nextDate.Format(dateFormat), nil
	}

	if repeat != "" {
		return "", fmt.Errorf("некорректный параметр репита: %v", repeat)
	}

	return "Opa", nil

}

func nextWeekday(now time.Time, weekdays []int) time.Time {
	nowWeekday := int(now.Weekday())
	if nowWeekday == 0 {
		nowWeekday = 7
	}

	for i := 0; i < 7; i++ {
		nextWeekday := (nowWeekday + i) % 7
		if nextWeekday == 0 {
			nextWeekday = 7
		}
		for _, wd := range weekdays {
			if wd == nextWeekday {
				return now.AddDate(0, 0, i)
			}

		}
	}

	return now
}

func getLastDayMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day() //moget tut sdelat time.local?
}
func nextMonthDay(now time.Time, days []int, months []int) time.Time {
	year := now.Year()
	month := now.Month()

	for {
		if len(months) > 0 {
			for _, m := range months {
				if int(month) < m {
					month = time.Month(m)
					break
				}
			}
		}

		for _, d := range days {
			var targetDay int
			if d < 0 {
				targetDay = getLastDayMonth(year, month) + d + 1
			} else {
				targetDay = d
			}

			if now.Day() <= targetDay {
				return time.Date(year, month, targetDay, 0, 0, 0, 0, time.UTC) //moget tut sdelat time.local?

			}

		}
		if month == 12 {
			year++
			month = 1
		} else {
			month++
		}
	}
}
