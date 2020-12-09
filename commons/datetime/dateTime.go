package datetime

import (
	"github.com/vjeantet/jodaTime"
	"time"
)

const (
	TimeFormat = "YYYY-MM-ddTHH:mm:ss.SSSZ"
)

type DateTime struct {
	timeObject time.Time
}

func Now() DateTime {
	return DateTime{time.Now()}
}

func FromString(timestamp string) (DateTime, error) {
	var dateTime DateTime
	timeObject, err := jodaTime.Parse(TimeFormat, timestamp)
	if nil == err {
		dateTime = DateTime{timeObject}
	}

	return dateTime, err
}

func FromUnixTime(unixTime int64) DateTime {
	var dateTime DateTime
	timeObject := time.Unix(0, unixTime*int64(time.Millisecond))
	dateTime = DateTime{timeObject}

	return dateTime
}

func (dateTime *DateTime) String() string {
	return jodaTime.Format(TimeFormat, dateTime.timeObject)
}

func (dateTime *DateTime) UnixTimestamp() int64 {
	return dateTime.timeObject.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func (dateTime *DateTime) WeekDay() int {
	return (int)(dateTime.timeObject.Weekday())
}

func (dateTime *DateTime) AddSecond(value int) *DateTime {
	dateTime.timeObject = dateTime.timeObject.Add(time.Second * time.Duration(value))
	return dateTime
}

func (dateTime *DateTime) AddMinute(value int) *DateTime {
	dateTime.timeObject = dateTime.timeObject.Add(time.Minute * time.Duration(value))
	return dateTime
}

func (dateTime *DateTime) AddHour(value int) *DateTime {
	dateTime.timeObject = dateTime.timeObject.Add(time.Hour * time.Duration(value))
	return dateTime
}

func (dateTime *DateTime) AddDay(value int) *DateTime {
	dateTime.timeObject = dateTime.timeObject.AddDate(0, 0, value)
	return dateTime
}

func (dateTime *DateTime) AddMonth(value int) *DateTime {
	dateTime.timeObject = dateTime.timeObject.AddDate(0, value, 0)
	return dateTime
}

func (dateTime *DateTime) AddYear(value int) *DateTime {
	dateTime.timeObject = dateTime.timeObject.AddDate(value, 0, 0)
	return dateTime
}
