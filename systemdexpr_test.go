package systemdexpr

/******************************************************************************/

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

/******************************************************************************/

type systemdNormTest struct {
	denormExp string
	normExp   string
}

var systemdNormTests = []systemdNormTest{
	{"Sat,Thu,Mon..Wed,Sat..Sun", "Mon..Thu,Sat,Sun *-*-* 00:00:00"},
	{"Mon,Sun 12-*-* 2,1:23", "Mon,Sun 2012-*-* 01,02:23:00"},
	{"Wed *-1", "Wed *-*-01 00:00:00"},
	{"Wed..Wed,Wed *-1", "Wed *-*-01 00:00:00"},
	{"Wed, 17:48", "Wed *-*-* 17:48:00"},
	{"Wed..Sat,Tue 12-10-15 1:2:3", "Tue..Sat 2012-10-15 01:02:03"},
	{"*-*-7 0:0:0", "*-*-07 00:00:00"},
	{"10-15", "*-10-15 00:00:00"},
	{"monday *-12-* 17:00", "Mon *-12-* 17:00:00"},
	{"Mon,Fri *-*-3,1,2 *:30:45", "Mon,Fri *-*-01,02,03 *:30:45"},
	{"12,14,13,12:20,10,30", "*-*-* 12,13,14:10,20,30:00"},
	{"12..14:10,20,30", "*-*-* 12..14:10,20,30:00"},
	{"mon,fri *-1/2-1,3 *:30:45", "Mon,Fri *-01/2-01,03 *:30:45"},
	{"03-05 08:05:40", "*-03-05 08:05:40"},
	{"08:05:40", "*-*-* 08:05:40"},
	{"05:40", "*-*-* 05:40:00"},
	{"Sat,Sun 12-05 08:05:40", "Sat,Sun *-12-05 08:05:40"},
	{"Sat,Sun 08:05:40", "Sat,Sun *-*-* 08:05:40"},
	{"2003-03-05 05:40", "2003-03-05 05:40:00"},
	{"2003-02..04-05", "2003-02..04-05 00:00:00"},
	{"2003-03-05 05:40 UTC", "2003-03-05 05:40:00 UTC"},
	{"2003-03-05", "2003-03-05 00:00:00"},
	{"03-05", "*-03-05 00:00:00"},
	{"hourly", "*-*-* *:00:00"},
	{"daily UTC", "*-*-* 00:00:00 UTC"},
	{"monthly", "*-*-01 00:00:00"},
	{"weekly", "Mon *-*-* 00:00:00"},
	{"weekly Pacific/Auckland", "Mon *-*-* 00:00:00 Pacific/Auckland"},
	{"yearly", "*-01-01 00:00:00"},
	{"annually", "*-01-01 00:00:00"},
	{"*:2/3", "*-*-* *:02/3:00"},
}

/******************************************************************************/
func TestSystemdTimers(t *testing.T) {

	locName := "America/Los_Angeles"
	loc, err := time.LoadLocation(locName)

	cases := []struct {
		name     string
		pattern  string
		initTime time.Time
		expected []time.Time
	}{
		{
			"normal time, w/o seconds",
			"05:40",
			time.Date(2019, time.February, 7, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.February, 7, 5, 40, 0, 0, loc),
				time.Date(2019, time.February, 8, 5, 40, 0, 0, loc),
				time.Date(2019, time.February, 9, 5, 40, 0, 0, loc),
			},
		},
		{
			"normal time w seconds",
			"05:40:00",
			time.Date(2019, time.February, 7, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.February, 7, 5, 40, 0, 0, loc),
				time.Date(2019, time.February, 8, 5, 40, 0, 0, loc),
				time.Date(2019, time.February, 9, 5, 40, 0, 0, loc),
			},
		},
		{
			"normal time with seconds",
			"08:05:40",
			time.Date(2019, time.February, 7, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.February, 7, 8, 5, 40, 0, loc),
				time.Date(2019, time.February, 8, 8, 5, 40, 0, loc),
				time.Date(2019, time.February, 9, 8, 5, 40, 0, loc),
				time.Date(2019, time.February, 10, 8, 5, 40, 0, loc),
				time.Date(2019, time.February, 11, 8, 5, 40, 0, loc),
				time.Date(2019, time.February, 12, 8, 5, 40, 0, loc),
			},
		},
		{
			"Date",
			"2023-03-05",
			time.Date(2019, time.February, 7, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2023, time.March, 5, 0, 0, 0, 0, loc),
				//time.Date(2019, time.February, 8, 2, 0, 0, 0, loc),
				//time.Date(2019, time.February, 9, 2, 0, 0, 0, loc),
			},
		},
		{
			"Date with time in past",
			"2003-03-05 05:40",
			time.Date(2019, time.February, 7, 1, 0, 0, 0, loc),
			[]time.Time{
				// deep back date
				time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
				time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
				time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			"Date with time",
			"2020-06-05 05:40:00",
			time.Date(2019, time.February, 7, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2020, time.June, 5, 5, 40, 0, 0, loc),
				// one date and deep back date
				time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
				time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			"Daily",
			"daily",
			time.Date(2019, time.February, 7, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.February, 8, 0, 0, 0, 0, loc),
				time.Date(2019, time.February, 9, 0, 0, 0, 0, loc),
				time.Date(2019, time.February, 10, 0, 0, 0, 0, loc),
			},
		},
		{
			"Date with month range",
			"2019-02..04-05",
			time.Date(2019, time.January, 7, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.February, 5, 0, 0, 0, 0, loc),
				time.Date(2019, time.March, 5, 0, 0, 0, 0, loc),
				time.Date(2019, time.April, 5, 0, 0, 0, 0, loc),
				// end of Next()
				time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			"Date with dom range",
			"2019-02-05..08",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.February, 5, 0, 0, 0, 0, loc),
				time.Date(2019, time.February, 6, 0, 0, 0, 0, loc),
				time.Date(2019, time.February, 7, 0, 0, 0, 0, loc),
				time.Date(2019, time.February, 8, 0, 0, 0, 0, loc),
				// end of Next()
				time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			"Date with year range",
			"2019..2023-02-05",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.February, 5, 0, 0, 0, 0, loc),
				time.Date(2020, time.February, 5, 0, 0, 0, 0, loc),
				time.Date(2021, time.February, 5, 0, 0, 0, 0, loc),
				time.Date(2022, time.February, 5, 0, 0, 0, 0, loc),
				time.Date(2023, time.February, 5, 0, 0, 0, 0, loc),
				// end of Next()
				time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			"Date with day range and divide",
			"2023-02-05..15/3",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2023, time.February, 5, 0, 0, 0, 0, loc),
				time.Date(2023, time.February, 8, 0, 0, 0, 0, loc),
				time.Date(2023, time.February, 11, 0, 0, 0, 0, loc),
				time.Date(2023, time.February, 14, 0, 0, 0, 0, loc),
				// end of Next()
				time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			"Every minute",
			"*-*-* *:*:00",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 4, 1, 1, 0, 0, loc),
				time.Date(2019, time.January, 4, 1, 2, 0, 0, loc),
				time.Date(2019, time.January, 4, 1, 3, 0, 0, loc),
				time.Date(2019, time.January, 4, 1, 4, 0, 0, loc),
			},
		},
		{
			"Every second",
			"*-*-* *:*:*",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 4, 1, 0, 1, 0, loc),
				time.Date(2019, time.January, 4, 1, 0, 2, 0, loc),
				time.Date(2019, time.January, 4, 1, 0, 3, 0, loc),
				time.Date(2019, time.January, 4, 1, 0, 4, 0, loc),
			},
		},
		{
			"Every 5 second",
			"*-*-* *:*:0/5",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 4, 1, 0, 5, 0, loc),
				time.Date(2019, time.January, 4, 1, 0, 10, 0, loc),
				time.Date(2019, time.January, 4, 1, 0, 15, 0, loc),
				time.Date(2019, time.January, 4, 1, 0, 20, 0, loc),
			},
		},
		{
			"Minutest with interval",
			"00:17..43",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 5, 0, 17, 0, 0, loc),
				time.Date(2019, time.January, 5, 0, 18, 0, 0, loc),
				time.Date(2019, time.January, 5, 0, 19, 0, 0, loc),
				time.Date(2019, time.January, 5, 0, 20, 0, 0, loc),
			},
		},
		{
			"Minutest list",
			"00:17,43",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 5, 0, 17, 0, 0, loc),
				time.Date(2019, time.January, 5, 0, 43, 0, 0, loc),
				time.Date(2019, time.January, 6, 0, 17, 0, 0, loc),
				time.Date(2019, time.January, 6, 0, 43, 0, 0, loc),
			},
		},
		{
			"Days of week: MON",
			"MON 00:00",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 7, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 14, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 21, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 28, 0, 0, 0, 0, loc),
				time.Date(2019, time.February, 4, 0, 0, 0, 0, loc),
			},
		},
		{
			"Days of week: friday",
			"friday 00:00",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 11, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 18, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 25, 0, 0, 0, 0, loc),
				time.Date(2019, time.February, 1, 0, 0, 0, 0, loc),
				time.Date(2019, time.February, 8, 0, 0, 0, 0, loc),
			},
		},
		{
			"Days of week list",
			"SUN,SAT 00:00",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 5, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 6, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 12, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 13, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 19, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 20, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 26, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 27, 0, 0, 0, 0, loc),
				time.Date(2019, time.February, 2, 0, 0, 0, 0, loc),
				time.Date(2019, time.February, 3, 0, 0, 0, 0, loc),
			},
		},
		{
			"Days of week range",
			"FRI-SAT 00:00",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 5, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 11, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 12, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 18, 0, 0, 0, 0, loc),
				time.Date(2019, time.January, 19, 0, 0, 0, 0, loc),
			},
		},
		{
			"Every ten minutes",
			"*-*-* *:*/10:00",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 4, 1, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 1, 20, 0, 0, loc),
				time.Date(2019, time.January, 4, 1, 30, 0, 0, loc),
				time.Date(2019, time.January, 4, 1, 40, 0, 0, loc),
				time.Date(2019, time.January, 4, 1, 50, 0, 0, loc),
				time.Date(2019, time.January, 4, 1, 60, 0, 0, loc),
				time.Date(2019, time.January, 4, 2, 10, 0, 0, loc),
			},
		},
		{
			"Every ten minutes with zero",
			"*-*-* *:0/10:00",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 4, 1, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 1, 20, 0, 0, loc),
				time.Date(2019, time.January, 4, 1, 30, 0, 0, loc),
				time.Date(2019, time.January, 4, 1, 40, 0, 0, loc),
				time.Date(2019, time.January, 4, 1, 50, 0, 0, loc),
				time.Date(2019, time.January, 4, 1, 60, 0, 0, loc),
				time.Date(2019, time.January, 4, 2, 10, 0, 0, loc),
			},
		},
		{
			"Every day at one hour",
			"*-*-* 01:00:00",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 5, 1, 0, 0, 0, loc),
				time.Date(2019, time.January, 6, 1, 0, 0, 0, loc),
				time.Date(2019, time.January, 7, 1, 0, 0, 0, loc),
				time.Date(2019, time.January, 8, 1, 0, 0, 0, loc),
				time.Date(2019, time.January, 9, 1, 0, 0, 0, loc),
				time.Date(2019, time.January, 10, 1, 0, 0, 0, loc),
				time.Date(2019, time.January, 11, 1, 0, 0, 0, loc),
			},
		},
		{
			"Complex hour rule",
			"*-*-* 0..2,4..5,7..23:10:00",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 4, 1, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 2, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 4, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 5, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 7, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 8, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 9, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 10, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 11, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 12, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 13, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 14, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 15, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 16, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 17, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 18, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 19, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 20, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 21, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 22, 10, 0, 0, loc),
				time.Date(2019, time.January, 4, 23, 10, 0, 0, loc),
				time.Date(2019, time.January, 5, 0, 10, 0, 0, loc),
			},
		},
		{
			"Range days, list hours",
			"*-*-1..5 04,12:00:00",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 4, 4, 0, 0, 0, loc),
				time.Date(2019, time.January, 4, 12, 0, 0, 0, loc),
				time.Date(2019, time.January, 5, 4, 0, 0, 0, loc),
				time.Date(2019, time.January, 5, 12, 0, 0, 0, loc),
			},
		},
		{
			"Hour divider",
			"*-*-* 0/3:00:00",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.January, 4, 3, 0, 0, 0, loc),
				time.Date(2019, time.January, 4, 6, 0, 0, 0, loc),
				time.Date(2019, time.January, 4, 9, 0, 0, 0, loc),
				time.Date(2019, time.January, 4, 12, 0, 0, 0, loc),
			},
		},
		{
			"Leap years",
			"*-02-29 01:00:00",
			time.Date(2019, time.January, 4, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2020, time.February, 29, 1, 0, 0, 0, loc),
				time.Date(2024, time.February, 29, 1, 0, 0, 0, loc),
				time.Date(2028, time.February, 29, 1, 0, 0, 0, loc),
				time.Date(2032, time.February, 29, 1, 0, 0, 0, loc),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			starting := c.initTime
			nextTimes := MustParse(c.pattern)
			for _, next := range c.expected {
				n := nextTimes.Next(starting)
				assert.NoError(t, err)
				assert.Equalf(t, next, n, "next time of %v", starting)

				starting = next
			}
		})
	}
}

/******************************************************************************/

func TestParseSystemd(t *testing.T) {
	var err error
	initTime := time.Date(2001, time.January, 4, 1, 0, 0, 0, time.UTC)
	for _, test := range systemdNormTests {
		denorm := MustParse(test.denormExp)
		norm := MustParse(test.normExp)
		assert.NoError(t, err)
		assert.Equalf(t, denorm.Next(initTime), norm.Next(initTime), "next time of %v", initTime)

	}
}

/******************************************************************************/

func TestZero(t *testing.T) {
	from, _ := time.Parse("2006-01-02", "2013-08-31")
	next := MustParse("1980-*-* *:*").Next(from)
	if next.IsZero() == false {
		t.Error(`("1980-*-* *:*").Next("2013-08-31").IsZero() returned 'false', expected 'true'`)
	}

	next = MustParse("2050-*-* *:*").Next(from)
	if next.IsZero() == true {
		t.Error(`("2050-*-* *:*").Next("2013-08-31").IsZero() returned 'true', expected 'false'`)
	}

	next = MustParse("2099-*-* *:*").Next(time.Time{})
	if next.IsZero() == false {
		t.Error(`("2099-*-* *:*").Next(time.Time{}).IsZero() returned 'true', expected 'false'`)
	}
}

/******************************************************************************/

func TestNextN(t *testing.T) {
	expected := []string{
		"Sat, 7 Sep 2013 00:00:00",
		"Sat, 14 Sep 2013 00:00:00",
		"Sat, 21 Sep 2013 00:00:00",
		"Sat, 28 Sep 2013 00:00:00",
		"Sat, 5 Oct 2013 00:00:00",
	}
	from, _ := time.Parse("2006-01-02 15:04:05", "2013-09-02 08:44:30")
	result := MustParse("SAT 00:00").NextN(from, uint(len(expected)))
	if len(result) != len(expected) {
		t.Errorf(`MustParse("SAT 00:00").NextN("2013-09-02 08:44:30", 5):\n"`)
		t.Errorf(`  Expected %d returned time values but got %d instead`, len(expected), len(result))
	}
	for i, next := range result {
		nextStr := next.Format("Mon, 2 Jan 2006 15:04:15")
		if nextStr != expected[i] {
			t.Errorf(`MustParse("SAT 00:00").NextN("2013-09-02 08:44:30", 5):\n"`)
			t.Errorf(`  result[%d]: expected "%s" but got "%s"`, i, expected[i], nextStr)
		}
	}
}

func TestNextN_every5min(t *testing.T) {
	expected := []string{
		"Mon, 2 Sep 2013 08:45:00",
		"Mon, 2 Sep 2013 08:50:00",
		"Mon, 2 Sep 2013 08:55:00",
		"Mon, 2 Sep 2013 09:00:00",
		"Mon, 2 Sep 2013 09:05:00",
	}
	from, _ := time.Parse("2006-01-02 15:04:05", "2013-09-02 08:44:32")
	result := MustParse("*:0/5").NextN(from, uint(len(expected)))
	if len(result) != len(expected) {
		t.Errorf(`MustParse("*/5 * * * *").NextN("2013-09-02 08:44:30", 5):\n"`)
		t.Errorf(`  Expected %d returned time values but got %d instead`, len(expected), len(result))
	}
	for i, next := range result {
		nextStr := next.Format("Mon, 2 Jan 2006 15:04:05")
		if nextStr != expected[i] {
			t.Errorf(`MustParse("*/5 * * * *").NextN("2013-09-02 08:44:30", 5):\n"`)
			t.Errorf(`  result[%d]: expected "%s" but got "%s"`, i, expected[i], nextStr)
		}
	}
}

func TestPeriodicConfig_DSTChange_Transitions(t *testing.T) {
	locName := "America/Los_Angeles"
	loc, err := time.LoadLocation(locName)
	require.NoError(t, err)

	cases := []struct {
		name     string
		pattern  string
		initTime time.Time
		expected []time.Time
	}{
		{
			"normal time",
			"2019-*-* 02:00",
			time.Date(2019, time.February, 7, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.February, 7, 2, 0, 0, 0, loc),
				time.Date(2019, time.February, 8, 2, 0, 0, 0, loc),
				time.Date(2019, time.February, 9, 2, 0, 0, 0, loc),
			},
		},
		{
			"Spring forward but not in switch time",
			"2019-*-* 04:00",
			time.Date(2019, time.March, 9, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.March, 9, 4, 0, 0, 0, loc),
				time.Date(2019, time.March, 10, 4, 0, 0, 0, loc),
				time.Date(2019, time.March, 11, 4, 0, 0, 0, loc),
			},
		},
		{
			"Spring forward at a skipped time odd",
			"2019-*-* 02:02",
			time.Date(2019, time.March, 9, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.March, 9, 2, 2, 0, 0, loc),
				// no time in March 10!
				time.Date(2019, time.March, 11, 2, 2, 0, 0, loc),
				time.Date(2019, time.March, 12, 2, 2, 0, 0, loc),
			},
		},
		{
			"Spring forward at a skipped time",
			"2019-*-* 02:01",
			time.Date(2019, time.March, 9, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.March, 9, 2, 1, 0, 0, loc),
				// no time in March 8!
				time.Date(2019, time.March, 11, 2, 1, 0, 0, loc),
				time.Date(2019, time.March, 12, 2, 1, 0, 0, loc),
			},
		},
		{
			"Spring forward at a skipped time boundary",
			"2019-*-* 02:00",
			time.Date(2019, time.March, 9, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.March, 9, 2, 0, 0, 0, loc),
				// no time in March 8!
				time.Date(2019, time.March, 11, 2, 0, 0, 0, loc),
				time.Date(2019, time.March, 12, 2, 0, 0, 0, loc),
			},
		},
		{
			"Spring forward at a boundary of repeating time",
			"2019-*-* 01:00",
			time.Date(2019, time.March, 9, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.March, 9, 1, 0, 0, 0, loc),
				time.Date(2019, time.March, 10, 0, 0, 0, 0, loc).Add(1 * time.Hour),
				time.Date(2019, time.March, 11, 1, 0, 0, 0, loc),
				time.Date(2019, time.March, 12, 1, 0, 0, 0, loc),
			},
		},
		{
			"Fall back: before transition",
			"2019-*-* 00:30",
			time.Date(2019, time.November, 3, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.November, 3, 0, 30, 0, 0, loc),
				time.Date(2019, time.November, 4, 0, 30, 0, 0, loc),
				time.Date(2019, time.November, 5, 0, 30, 0, 0, loc),
				time.Date(2019, time.November, 6, 0, 30, 0, 0, loc),
			},
		},
		{
			"Fall back: after transition",
			"2019-*-* 03:30",
			time.Date(2019, time.November, 3, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.November, 3, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 4, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 5, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 6, 3, 30, 0, 0, loc),
			},
		},
		{
			"Fall back: after transition starting in repeated span before",
			"2019-*-* 03:30",
			time.Date(2019, time.November, 3, 0, 10, 0, 0, loc).Add(1 * time.Hour),
			[]time.Time{
				time.Date(2019, time.November, 3, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 4, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 5, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 6, 3, 30, 0, 0, loc),
			},
		},
		{
			"Fall back: after transition starting in repeated span after",
			"2019-*-* 03:30",
			time.Date(2019, time.November, 3, 0, 10, 0, 0, loc).Add(2 * time.Hour),
			[]time.Time{
				time.Date(2019, time.November, 3, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 4, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 5, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 6, 3, 30, 0, 0, loc),
			},
		},
		{
			"Fall back: in repeated region",
			"2019-*-* 01:30",
			time.Date(2019, time.November, 3, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.November, 3, 0, 30, 0, 0, loc).Add(1 * time.Hour),
				time.Date(2019, time.November, 3, 0, 30, 0, 0, loc).Add(2 * time.Hour),
				time.Date(2019, time.November, 4, 1, 30, 0, 0, loc),
				time.Date(2019, time.November, 5, 1, 30, 0, 0, loc),
				time.Date(2019, time.November, 6, 1, 30, 0, 0, loc),
			},
		},
		{
			"Fall back: in repeated region boundary",
			"2019-*-* 01:00",
			time.Date(2019, time.November, 3, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.November, 3, 0, 0, 0, 0, loc).Add(1 * time.Hour),
				time.Date(2019, time.November, 3, 0, 0, 0, 0, loc).Add(2 * time.Hour),
				time.Date(2019, time.November, 4, 1, 0, 0, 0, loc),
				time.Date(2019, time.November, 5, 1, 0, 0, 0, loc),
				time.Date(2019, time.November, 6, 1, 0, 0, 0, loc),
			},
		},
		{
			"Fall back: in repeated region boundary 2",
			"2019-*-* 02:00",
			time.Date(2019, time.November, 3, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.November, 3, 0, 0, 0, 0, loc).Add(3 * time.Hour),
				time.Date(2019, time.November, 4, 2, 0, 0, 0, loc),
				time.Date(2019, time.November, 5, 2, 0, 0, 0, loc),
				time.Date(2019, time.November, 6, 2, 0, 0, 0, loc),
			},
		},
		{
			"Fall back: in repeated region, starting from within region",
			"2019-*-* 01:30",
			time.Date(2019, time.November, 3, 0, 40, 0, 0, loc).Add(1 * time.Hour),
			[]time.Time{
				time.Date(2019, time.November, 3, 0, 30, 0, 0, loc).Add(2 * time.Hour),
				time.Date(2019, time.November, 4, 1, 30, 0, 0, loc),
				time.Date(2019, time.November, 5, 1, 30, 0, 0, loc),
				time.Date(2019, time.November, 6, 1, 30, 0, 0, loc),
			},
		},
		{
			"Fall back: in repeated region, starting from within region 2",
			"2019-*-* 01:30",
			time.Date(2019, time.November, 3, 0, 40, 0, 0, loc).Add(2 * time.Hour),
			[]time.Time{
				time.Date(2019, time.November, 4, 1, 30, 0, 0, loc),
				time.Date(2019, time.November, 5, 1, 30, 0, 0, loc),
				time.Date(2019, time.November, 6, 1, 30, 0, 0, loc),
			},
		},
		{
			"Fall back: wildcard",
			"2019-*-* *:30",
			time.Date(2019, time.November, 3, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.November, 3, 0, 30, 0, 0, loc),
				time.Date(2019, time.November, 3, 0, 30, 0, 0, loc).Add(1 * time.Hour),
				time.Date(2019, time.November, 3, 0, 30, 0, 0, loc).Add(2 * time.Hour),
				time.Date(2019, time.November, 3, 2, 30, 0, 0, loc),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			expr := MustParse(c.pattern)

			starting := c.initTime
			for _, next := range c.expected {
				n := expr.Next(starting)
				if next != n {
					t.Fatalf("next(%v) = %v not %v", starting, next, n)
				}

				starting = next
			}
		})
	}
}

func TestPeriodicConfig_DSTChange_Transitions_LordHowe(t *testing.T) {
	locName := "Australia/Lord_Howe"
	loc, err := time.LoadLocation(locName)
	require.NoError(t, err)

	cases := []struct {
		name     string
		pattern  string
		initTime time.Time
		expected []time.Time
	}{
		{
			"normal time",
			"2019-*-* 02:00",
			time.Date(2019, time.February, 7, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.February, 7, 2, 0, 0, 0, loc),
				time.Date(2019, time.February, 8, 2, 0, 0, 0, loc),
				time.Date(2019, time.February, 9, 2, 0, 0, 0, loc),
			},
		},
		{
			"backward: non repeated portion of the hour",
			"2019-*-* 01:03",
			time.Date(2019, time.April, 6, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.April, 6, 1, 3, 0, 0, loc),
				time.Date(2019, time.April, 7, 1, 3, 0, 0, loc),
				time.Date(2019, time.April, 8, 1, 3, 0, 0, loc),
				time.Date(2019, time.April, 9, 1, 3, 0, 0, loc),
			},
		},
		{
			"backward: repeated portion of the hour",
			"2019-*-* 01:31",
			time.Date(2019, time.April, 6, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.April, 6, 1, 31, 0, 0, loc),
				time.Date(2019, time.April, 7, 0, 31, 0, 0, loc).Add(60 * time.Minute),
				time.Date(2019, time.April, 7, 0, 31, 0, 0, loc).Add(90 * time.Minute),
				time.Date(2019, time.April, 8, 1, 31, 0, 0, loc),
				time.Date(2019, time.April, 9, 1, 31, 0, 0, loc),
			},
		},
		{
			"forward: skipped portion of the hour",
			"2019-*-* 02:03",
			time.Date(2019, time.October, 5, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.October, 5, 2, 3, 0, 0, loc),
				// no Oct 6
				time.Date(2019, time.October, 7, 2, 3, 0, 0, loc),
				time.Date(2019, time.October, 8, 2, 3, 0, 0, loc),
				time.Date(2019, time.October, 9, 2, 3, 0, 0, loc),
			},
		},
		{
			"forward: non-skipped portion of the hour",
			"2019-*-* 02:31",
			time.Date(2019, time.October, 5, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.October, 5, 2, 31, 0, 0, loc),
				// no Oct 6
				time.Date(2019, time.October, 7, 2, 31, 0, 0, loc),
				time.Date(2019, time.October, 8, 2, 31, 0, 0, loc),
				time.Date(2019, time.October, 9, 2, 31, 0, 0, loc),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			expr := MustParse(c.pattern)

			starting := c.initTime
			for _, next := range c.expected {
				n := expr.Next(starting)
				if next != n {
					t.Fatalf("next(%v) = %v not %v", starting, next, n)
				}

				starting = next
			}
		})
	}
}

func TestNext_DaylightSaving_Property(t *testing.T) {
	locName := "America/Los_Angeles"
	loc, err := time.LoadLocation(locName)
	if err != nil {
		t.Fatalf("failed to get location: %v", err)
	}

	cronExprs := []string{
		"*:*",
		"02:30",
		"01:*",
	}

	times := []time.Time{
		// spring forward
		time.Date(2019, time.March, 11, 0, 0, 0, 0, loc),
		time.Date(2019, time.March, 10, 0, 0, 0, 0, loc),
		time.Date(2019, time.March, 11, 0, 0, 0, 0, loc),

		// leap backwards
		time.Date(2019, time.November, 4, 0, 0, 0, 0, loc),
		time.Date(2019, time.November, 5, 0, 0, 0, 0, loc),
		time.Date(2019, time.November, 6, 0, 0, 0, 0, loc),
	}

	testSpan := 4 * time.Hour

	testCase := func(t *testing.T, cronExpr string, init time.Time) {
		cron := MustParse(cronExpr)

		prevNext := init
		for start := init; start.Before(init.Add(testSpan)); start = start.Add(1 * time.Minute) {
			next := cron.Next(start)
			if !next.After(start) {
				t.Fatalf("next(%v) = %v is not after start time", start, next)
			}

			if next.Before(prevNext) {
				t.Fatalf("next(%v) = %v reverted back in time from %v", start, next, prevNext)
			}

			if strings.HasPrefix(cronExpr, "* * ") {
				if next.Sub(start) != time.Minute {
					t.Fatalf("next(%v) = %v should be the next minute", start, next)
				}
			}

			prevNext = next
		}
	}

	for _, cron := range cronExprs {
		for _, startTime := range times {
			t.Run(fmt.Sprintf("%v: %v", cron, startTime), func(t *testing.T) {
				testCase(t, cron, startTime)
			})
		}
	}
}

func TestNext_DaylightSaving_Property_LordHowe(t *testing.T) {
	// Lord Howe, Australia is at GMT+1100 April-October and GMT+1030 otherwise.
	//
	// On April 7, 2019, at when clock approches 2am, the clock
	// transitions to 1.30am.
	//
	// On October 6, when the clock approaches 2am, the clock transitions
	// to 2.30am.
	locName := "Australia/Lord_Howe"
	loc, err := time.LoadLocation(locName)
	if err != nil {
		t.Fatalf("failed to get location: %v", err)
	}

	cronExprs := []string{
		"*:*",
		"02:00",
		"01:*",
		"1:53",
		"02:05",
	}

	times := []time.Time{
		// spring forward
		time.Date(2019, time.April, 6, 0, 0, 0, 0, loc),
		time.Date(2019, time.April, 7, 0, 0, 0, 0, loc),
		time.Date(2019, time.April, 8, 0, 0, 0, 0, loc),

		// leap backwards
		time.Date(2019, time.October, 5, 0, 0, 0, 0, loc),
		time.Date(2019, time.October, 6, 0, 0, 0, 0, loc),
		time.Date(2019, time.October, 7, 0, 0, 0, 0, loc),
	}

	testSpan := 4 * time.Hour

	testCase := func(t *testing.T, cronExpr string, init time.Time) {
		cron := MustParse(cronExpr)

		prevNext := init
		for start := init; start.Before(init.Add(testSpan)); start = start.Add(1 * time.Minute) {
			next := cron.Next(start)
			if !next.After(start) {
				t.Fatalf("next(%v) = %v is not after start time", start, next)
			}

			if next.Before(prevNext) {
				t.Fatalf("next(%v) = %v reverted back in time from %v", start, next, prevNext)
			}

			if strings.HasPrefix(cronExpr, "* * ") {
				if next.Sub(start) != time.Minute {
					t.Fatalf("next(%v) = %v should be the next minute", start, next)
				}
			}

			prevNext = next
		}
	}

	for _, cron := range cronExprs {
		for _, startTime := range times {
			t.Run(fmt.Sprintf("%v: %v", cron, startTime), func(t *testing.T) {
				testCase(t, cron, startTime)
			})
		}
	}
}

func TestNext_DaylightSaving_Property_Brazil(t *testing.T) {
	// Until 2018, Brazil/Sao Paulo and some South American countries used
	// to transition for daylight savings at midnight.
	//
	// When the clock approaches 2018-11-04 midnight, the clock transitions to 1am.
	locName := "America/Sao_Paulo"
	loc, err := time.LoadLocation(locName)
	if err != nil {
		t.Fatalf("failed to get location: %v", err)
	}

	cronExprs := []string{
		"*:*",
		"02:00",
		"01:*",
		"01:05",
		"23:05",
	}

	times := []time.Time{
		// spring forward
		time.Date(2018, time.February, 16, 22, 0, 0, 0, loc),
		time.Date(2018, time.February, 17, 22, 0, 0, 0, loc),
		time.Date(2018, time.February, 18, 22, 0, 0, 0, loc),

		// leap backwards
		time.Date(2018, time.November, 3, 23, 0, 0, 0, loc),
		time.Date(2018, time.November, 3, 23, 0, 0, 0, loc),
		time.Date(2018, time.November, 3, 23, 0, 0, 0, loc),
	}

	testSpan := 4 * time.Hour

	testCase := func(t *testing.T, cronExpr string, init time.Time) {
		cron := MustParse(cronExpr)

		prevNext := init
		for start := init; start.Before(init.Add(testSpan)); start = start.Add(1 * time.Minute) {
			next := cron.Next(start)
			if !next.After(start) {
				t.Fatalf("next(%v) = %v is not after start time", start, next)
			}

			if next.Before(prevNext) {
				t.Fatalf("next(%v) = %v reverted back in time from %v", start, next, prevNext)
			}

			if strings.HasPrefix(cronExpr, "* * ") {
				if next.Sub(start) != time.Minute {
					t.Fatalf("next(%v) = %v should be the next minute", start, next)
				}
			}

			prevNext = next
		}
	}

	for _, cron := range cronExprs {
		for _, startTime := range times {
			t.Run(fmt.Sprintf("%v: %v", cron, startTime), func(t *testing.T) {
				testCase(t, cron, startTime)
			})
		}
	}
}

// Issue: https://github.com/gorhill/cronexpr/issues/16
func TestInterval_Interval60Issue(t *testing.T) {
	_, err := Parse("*/60 * * * * *")
	if err == nil {
		t.Errorf("parsing with interval 60 should return err")
	}

	_, err = Parse("*:0/61")
	if err == nil {
		t.Errorf("parsing with interval 61 should return err")
	}

	_, err = Parse("*:2/60")
	if err == nil {
		t.Errorf("parsing with interval 60 should return err")
	}

	_, err = Parse("*:2..20/61")
	if err == nil {
		t.Errorf("parsing with interval 60 should return err")
	}
}

/******************************************************************************/

var benchmarkExpressions = []string{
	"*:*",
	"hourly",
	"weekly",
	"yearly",
	"Mon..Thu,Sat,Sun *-*-* 00:00:00",
	"Tue..Sat 2012-10-15 01:02:03",
	"2003-02..04-05 00:00:00",
}
var benchmarkExpressionsLen = len(benchmarkExpressions)

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MustParse(benchmarkExpressions[i%benchmarkExpressionsLen])
	}
}

func BenchmarkNext(b *testing.B) {
	exprs := make([]*Expression, benchmarkExpressionsLen)
	for i := 0; i < benchmarkExpressionsLen; i++ {
		exprs[i] = MustParse(benchmarkExpressions[i])
	}
	from := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expr := exprs[i%benchmarkExpressionsLen]
		next := expr.Next(from)
		next = expr.Next(next)
		next = expr.Next(next)
		next = expr.Next(next)
		next = expr.Next(next)
	}
}
