package proc

import (
	"testing"
)

func TestStatParse(t *testing.T) {
	pf := "/proc/2051/stat"
	s := NewStat(pf)
	_, err := s.Parse()
	if err != nil {
		t.Error(err)
		return
	}
	// fmt.Println(fields.State)

}
