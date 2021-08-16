package routers

import (
	"github.com/maxmind/mmdbinspect/pkg/mmdbinspect"
	"testing"
)

func TestLookUp(t *testing.T) {
	ip := "51.38.112.54"
	r, e := LookUp(ip)
	t.Error(e)
	t.Error(mmdbinspect.RecordToString(r))
}
