package util_test

import (
	"os"
	"testing"

	"github.com/imup-io/client/util"
	"github.com/matryer/is"
)

func Test_ValueOr(t *testing.T) {
	is := is.New(t)
	var str *string
	fsv := util.ValueOr(str, "STRING_ENV", "fallback")
	is.Equal(fsv, "fallback")

	os.Setenv("STRING_ENV", "foo")
	esv := util.ValueOr(str, "STRING_ENV", "fallback")
	is.Equal(esv, "foo")

	os.Clearenv()

	ptr := "pointer"
	str = &ptr
	psv := util.ValueOr(str, "STRING_ENV", "fallback")
	is.Equal(psv, "pointer")
}

func Test_BooleanValueOr(t *testing.T) {
	is := is.New(t)
	var boolean *bool

	fbv := util.BooleanValueOr(boolean, "BOOL_ENV", "false")
	is.Equal(fbv, false)

	os.Setenv("BOOL_ENV", "true")
	ebv := util.BooleanValueOr(boolean, "BOOL_ENV", "false")
	is.Equal(ebv, true)

	os.Clearenv()

	ptr := true
	boolean = &ptr
	pbv := util.BooleanValueOr(boolean, "BOOL_ENV", "false")
	is.Equal(pbv, true)
}
