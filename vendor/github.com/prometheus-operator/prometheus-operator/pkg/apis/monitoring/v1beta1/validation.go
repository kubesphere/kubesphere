// Copyright 2021 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1beta1

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func (hc *HTTPConfig) Validate() error {
	if hc == nil {
		return nil
	}

	if (hc.BasicAuth != nil || hc.OAuth2 != nil) && (hc.BearerTokenSecret != nil) {
		return fmt.Errorf("at most one of basicAuth, oauth2, bearerTokenSecret must be configured")
	}

	if hc.Authorization != nil {
		if hc.BearerTokenSecret != nil {
			return fmt.Errorf("authorization is not compatible with bearerTokenSecret")
		}

		if hc.BasicAuth != nil || hc.OAuth2 != nil {
			return fmt.Errorf("at most one of basicAuth, oauth2 & authorization must be configured")
		}

		if err := hc.Authorization.Validate(); err != nil {
			return err
		}
	}

	if hc.OAuth2 != nil {
		if hc.BasicAuth != nil {
			return fmt.Errorf("at most one of basicAuth, oauth2 & authorization must be configured")
		}

		if err := hc.OAuth2.Validate(); err != nil {
			return err
		}
	}

	if hc.TLSConfig != nil {
		if err := hc.TLSConfig.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate the TimeInterval
func (ti TimeInterval) Validate() error {
	if ti.Name == "" {
		return errors.New("empty name field for time interval")
	}

	for i, ti := range ti.TimeIntervals {
		for _, time := range ti.Times {
			if err := time.Validate(); err != nil {
				return fmt.Errorf("time range at %d is invalid: %w", i, err)
			}
		}
		for _, weekday := range ti.Weekdays {
			if err := weekday.Validate(); err != nil {
				return fmt.Errorf("weekday range at %d is invalid: %w", i, err)
			}
		}
		for _, dom := range ti.DaysOfMonth {
			if err := dom.Validate(); err != nil {
				return fmt.Errorf("day of month range at %d is invalid: %w", i, err)
			}
		}
		for _, month := range ti.Months {
			if err := month.Validate(); err != nil {
				return fmt.Errorf("month range at %d is invalid: %w", i, err)
			}
		}
		for _, year := range ti.Years {
			if err := year.Validate(); err != nil {
				return fmt.Errorf("year range at %d is invalid: %w", i, err)
			}
		}
	}
	return nil
}

// Validate the TimeRange
func (tr TimeRange) Validate() error {
	_, err := tr.Parse()
	return err
}

// Parse returns a ParsedRange on valid input or an error if the fields cannot be parsed
// End of the day is represented as 1440.
func (tr TimeRange) Parse() (*ParsedRange, error) {
	if tr.StartTime == "" || tr.EndTime == "" {
		return nil, fmt.Errorf("start and end are required")
	}

	start, err := parseTime(string(tr.StartTime))
	if err != nil {
		return nil, fmt.Errorf("start time invalid: %w", err)
	}

	end, err := parseTime(string(tr.EndTime))
	if err != nil {
		return nil, fmt.Errorf("end time invalid: %w", err)
	}

	if start >= end {
		return nil, fmt.Errorf("start time %d cannot be equal or greater than end time %d", start, end)
	}
	return &ParsedRange{
		Start: start,
		End:   end,
	}, nil
}

// Validate the WeekdayRange
func (wr WeekdayRange) Validate() error {
	_, err := wr.Parse()
	return err
}

// Parse returns a ParsedRange on valid input or an error if the fields cannot be parsed
// The week starts on Sunday -> 0
func (wr WeekdayRange) Parse() (*ParsedRange, error) {
	startStr, endStr, err := parseRange(string(wr))
	if err != nil {
		return nil, err
	}

	start, err := Weekday(startStr).Int()
	if err != nil {
		return nil, fmt.Errorf("failed to parse start day from weekday range: %w", err)
	}

	end, err := Weekday(endStr).Int()
	if err != nil {
		return nil, fmt.Errorf("failed to parse end day from weekday range: %w", err)
	}

	if start > end {
		return nil, errors.New("start day cannot be before end day")
	}
	if start < 0 || start > 6 {
		return nil, fmt.Errorf("%s is not a valid day of the week: out of range", startStr)
	}
	if end < 0 || end > 6 {
		return nil, fmt.Errorf("%s is not a valid day of the week: out of range", endStr)
	}
	return &ParsedRange{Start: start, End: end}, nil
}

// Validate the YearRange
func (yr YearRange) Validate() error {
	_, err := yr.Parse()
	return err
}

// Parse returns a ParsedRange on valid input or an error if the fields cannot be parsed
func (yr YearRange) Parse() (*ParsedRange, error) {
	startStr, endStr, err := parseRange(string(yr))
	if err != nil {
		return nil, err
	}

	start, err := strconv.Atoi(startStr)
	if err != nil {
		fmt.Errorf("start year cannot be %s parsed: %w", startStr, err)
	}

	end, err := strconv.Atoi(endStr)
	if err != nil {
		fmt.Errorf("end year cannot be %s parsed: %w", endStr, err)
	}

	if start > end {
		return nil, fmt.Errorf("end year %d is before start year %d", end, start)
	}
	return &ParsedRange{Start: start, End: end}, nil
}

// Int returns an integer, which is the canonical representation
// of the Weekday in upstream types.
// Returns an error if the Weekday is invalid
func (w Weekday) Int() (int, error) {
	normaliseWeekday := Weekday(strings.ToLower(string(w)))

	day, found := daysOfWeek[normaliseWeekday]
	if !found {
		i, err := strconv.Atoi(string(normaliseWeekday))
		if err != nil {
			return day, fmt.Errorf("%s is an invalid weekday", w)
		}
		day = i
	}

	return day, nil
}

// Int validates the Month and returns an integer, which is the canonical representation
// of the Month in upstream types.
// Returns an error if the Month is invalid
func (m Month) Int() (int, error) {
	normaliseMonth := Month(strings.ToLower(string(m)))

	day, found := months[normaliseMonth]
	if !found {
		i, err := strconv.Atoi(string(normaliseMonth))
		if err != nil {
			return day, fmt.Errorf("%s is an invalid month", m)
		}
		day = i
	}

	return day, nil
}

// Validate the DayOfMonthRange
func (r DayOfMonthRange) Validate() error {
	// Note: Validation is copied from UnmarshalYAML for DayOfMonthRange in alertmanager repo

	// Check beginning <= end accounting for negatives day of month indices as well.
	// Months != 31 days can't be addressed here and are clamped, but at least we can catch blatant errors.
	if r.Start == 0 || r.Start < -31 || r.Start > 31 {
		return fmt.Errorf("%d is not a valid day of the month: out of range", r.Start)
	}
	if r.End == 0 || r.End < -31 || r.End > 31 {
		return fmt.Errorf("%d is not a valid day of the month: out of range", r.End)
	}
	// Restricting here prevents errors where begin > end in longer months but not shorter months.
	if r.Start < 0 && r.End > 0 {
		return fmt.Errorf("end day must be negative if start day is negative")
	}
	// Check begin <= end. We can't know this for sure when using negative indices,
	// but we can prevent cases where its always invalid (using 28 day minimum length).
	checkBegin := r.Start
	checkEnd := r.End
	if r.Start < 0 {
		checkBegin = 28 + r.Start
	}
	if r.End < 0 {
		checkEnd = 28 + r.End
	}
	if checkBegin > checkEnd {
		return fmt.Errorf("end day %d is always before start day %d", r.End, r.Start)
	}
	return nil
}

// Validate the month range
func (mr MonthRange) Validate() error {
	_, err := mr.Parse()
	return err
}

// Parse returns a ParsedMonthRange or error on invalid input
func (mr MonthRange) Parse() (*ParsedRange, error) {
	startStr, endStr, err := parseRange(string(mr))
	if err != nil {
		return nil, err
	}

	start, err := Month(startStr).Int()
	if err != nil {
		return nil, fmt.Errorf("failed to parse start month from month range: %w", err)
	}

	end, err := Month(endStr).Int()
	if err != nil {
		return nil, fmt.Errorf("failed to parse start month from month range: %w", err)
	}

	if start > end {
		return nil, fmt.Errorf("end month %s is before start month %s", endStr, startStr)
	}
	return &ParsedRange{
		Start: start,
		End:   end,
	}, nil
}

// ParsedRange is an integer representation of a range
// +kubebuilder:object:generate:=false
type ParsedRange struct {
	// Start is the beginning of the range
	Start int `json:"start,omitempty"`
	// End of the range
	End int `json:"end,omitempty"`
}

var validTime = "^((([01][0-9])|(2[0-3])):[0-5][0-9])$|(^24:00$)"
var validTimeRE = regexp.MustCompile(validTime)

// Converts a string of the form "HH:MM" into the number of minutes elapsed in the day.
func parseTime(in string) (mins int, err error) {
	if !validTimeRE.MatchString(in) {
		return 0, fmt.Errorf("couldn't parse timestamp %s, invalid format", in)
	}
	timestampComponents := strings.Split(in, ":")
	if len(timestampComponents) != 2 {
		return 0, fmt.Errorf("invalid timestamp format: %s", in)
	}
	timeStampHours, err := strconv.Atoi(timestampComponents[0])
	if err != nil {
		return 0, err
	}
	timeStampMinutes, err := strconv.Atoi(timestampComponents[1])
	if err != nil {
		return 0, err
	}
	if timeStampHours < 0 || timeStampHours > 24 || timeStampMinutes < 0 || timeStampMinutes > 60 {
		return 0, fmt.Errorf("timestamp %s out of range", in)
	}
	// Timestamps are stored as minutes elapsed in the day, so multiply hours by 60.
	mins = timeStampHours*60 + timeStampMinutes
	return mins, nil
}

// parseRange parses a valid range string into parts
func parseRange(in string) (start, end string, err error) {
	if !strings.ContainsRune(in, ':') {
		return in, in, nil
	}

	parts := strings.Split(string(in), ":")
	if len(parts) != 2 {
		return start, end, fmt.Errorf("invalid range provided %s", in)
	}
	return parts[0], parts[1], nil
}
