package proc

import (
	"fmt"
	"testing"
)

func TestStatParse(t *testing.T) {
	pf := "/proc/3579/stat"
	s := NewStat(pf)
	fields, err := s.Parse()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(*fields)

}
