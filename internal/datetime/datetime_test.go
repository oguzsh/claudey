package datetime

import (
	"regexp"
	"testing"
)

func TestDateString(t *testing.T) {
	s := DateString()
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, s)
	if !matched {
		t.Errorf("DateString() = %q, want YYYY-MM-DD format", s)
	}
}

func TestTimeString(t *testing.T) {
	s := TimeString()
	matched, _ := regexp.MatchString(`^\d{2}:\d{2}$`, s)
	if !matched {
		t.Errorf("TimeString() = %q, want HH:MM format", s)
	}
}

func TestDateTimeString(t *testing.T) {
	s := DateTimeString()
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`, s)
	if !matched {
		t.Errorf("DateTimeString() = %q, want YYYY-MM-DD HH:MM:SS format", s)
	}
}




