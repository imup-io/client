package util_test

import (
	"os"
	"testing"

	"github.com/imup-io/client/config"
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

	os.Setenv("BOOL_ENV", "true")
	ptr = false
	boolean = &ptr
	pointerPrecedence := util.BooleanValueOr(boolean, "BOOL_ENV", "true")
	is.Equal(pointerPrecedence, false)

	os.Setenv("BOOL_ENV", "true")
	ptr = false
	boolean = &ptr
	envPref := util.BooleanValueOr(boolean, "BOOL_ENV", "false")
	is.Equal(envPref, true)
}

func Test_IPMonitored(t *testing.T) {
	is := is.New(t)
	os.Setenv("EMAIL", "test@example.com")

	cfg, err := config.New()
	is.NoErr(err)
	is.Equal(true, util.IPMonitored("10.0.0.1", cfg.AllowedIPs(), cfg.BlockedIPs()))
	is.Equal(true, util.IPMonitored("127.0.0.1", cfg.AllowedIPs(), cfg.BlockedIPs()))

	os.Setenv("ALLOWLISTED_IPS", "192.168.1.1/32")
	cfg, err = config.New()
	is.NoErr(err)
	is.Equal(true, util.IPMonitored("192.168.1.1", cfg.AllowedIPs(), cfg.BlockedIPs()))
	is.Equal(false, util.IPMonitored("127.0.0.1", cfg.AllowedIPs(), cfg.BlockedIPs()))

	os.Setenv("ALLOWLISTED_IPS", "")
	os.Setenv("BLOCKLISTED_IPS", "127.0.0.1/32,1.1.1.1/32")
	cfg, err = config.New()
	is.NoErr(err)
	is.Equal(false, util.IPMonitored("127.0.0.1", cfg.AllowedIPs(), cfg.BlockedIPs()))
	is.Equal(true, util.IPMonitored("192.168.1.1", cfg.AllowedIPs(), cfg.BlockedIPs()))

}
