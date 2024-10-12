package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	masReap := strings.Split(repeat, " ")
	if repeat == "" {
		return "", fmt.Errorf("repeat: Lenth=0")
	}
	dateParse, err := time.Parse("20060102", date)
	if err != nil {
		return "", err
	}
	if len(masReap) > 0 {
		switch masReap[0] {
		case "d":
			if len(masReap) > 1 {
				if value, err := strconv.Atoi(masReap[1]); err == nil {
					if value <= 400 {
						dateParse = dateParse.AddDate(0, 0, value)
						for dateParse.Before(now) {
							dateParse = dateParse.AddDate(0, 0, value)
						}
					} else {
						return "", fmt.Errorf("masReap: The day is biggest 400")
					}
				} else {
					return "", fmt.Errorf("masReap: The day is not in int format")
				}
			} else {
				return "", fmt.Errorf("masReap: The number of days is missing")
			}
		case "y":
			dateParse = dateParse.AddDate(1, 0, 0)
			for dateParse.Before(now) {
				dateParse = dateParse.AddDate(1, 0, 0)
			}
		default:
			return "", fmt.Errorf("masReap: The wrong format")
		}
	} else {
		return "", fmt.Errorf("masReap: Length is less than 0")
	}
	return dateParse.Format("20060102"), nil
}
