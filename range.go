package icws

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
)

type Range struct {
	Unit  string
	First int
	Last  int
	Total int
}

// NewRange creates a new Range
func NewRange(unit string) Range {
	return Range{
		Unit:  unit,
		First: 0,
		Last:  0,
		Total: math.MaxInt32,
	}
}

// IsCollapsed tells if the range is collapsed
//
// A range is considered collapsed if its First and Last Indices are the same
func (r Range) IsCollapsed() bool {
	return r.First == r.Last
}

func (r Range) IsAtEnd() bool {
	return r.Last >= (r.Total - 1)
}

// ToMap fills the headers with the range if it is not collapsed
func (r Range) ToMap(data map[string]string) {
	if !r.IsCollapsed() {
		if len(r.Unit) == 0 {
			data["Range"] = fmt.Sprintf("bytes=%d-%d", r.First, r.Last)
		} else {
			data["Range"] = fmt.Sprintf("%s=%d-%d", r.Unit, r.First, r.Last)
		}
	}
}

// ToHeader fills the HTTP Header with the range if it is not collapsed
func (r Range) ToHeader(header *http.Header) {
	if !r.IsCollapsed() {
		if len(r.Unit) == 0 {
			header.Set("Range", fmt.Sprintf("bytes=%d-%d", r.First, r.Last))
		} else {
			header.Set("Range", fmt.Sprintf("%s=%d-%d", r.Unit, r.First, r.Last))
		}
	}
}

// Get a Range from an HTTP Header map
func GetRangeFromHeader(header http.Header) Range {
	var r Range
	if contentRange := header.Get("Content-Range"); len(contentRange) > 0 {
		components := regexp.MustCompile(`\s*(?P<unit>[[:alpha:]]+)\s+(?P<min>[[:digit:]]+)-(?P<max>[[:digit:]]+)\/(?P<total>[[:digit:]]+)`)

		if matches := components.FindStringSubmatch(contentRange); matches != nil {
			var err error
			r.Unit = matches[components.SubexpIndex("unit")]
			r.First, err = strconv.Atoi(matches[components.SubexpIndex("min")])
			if err != nil {
				r.First = 0
			}
			r.Last, err = strconv.Atoi(matches[components.SubexpIndex("max")])
			if err != nil {
				r.Last = 0
			}
			value := matches[components.SubexpIndex("total")]
			if value == "*" {
				r.Total = math.MaxInt32
			} else {
				r.Total, err = strconv.Atoi(value)
				if err != nil {
					r.Total = 0
				}
			}
		}
	}
	return r
}
