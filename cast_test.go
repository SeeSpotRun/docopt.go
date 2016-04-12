package docopt

import (
	"testing"
)

func TestRead(t *testing.T) {


	const (
		intstr = "-10"
		intval = -10

		uintstr = "1000000000000000000"
		uintval = 1000000000000000000

		realval = 1.234
		realstr = "1.234"

		byteval = 512 * 7
		bytestr = "3.5k"

		strval = "hello"
	)

        usage := `Usage:
  cast_test [options]

Options:
  --intval=N       Sets intval
  --uintval=N      Sets uintval
  --byteval=N      Sets byteval, optionally add suffix K/M/G/T for kilobytes etc
  --realval=X      Sets realval
  --boolval        Sets boolval to true
  --strval=string  Sets strval to string

`

	type Options struct{
		Xintval  int
		Xuintval uint64
		Xboolval bool
		Xbyteval uint32
		Xrealval float32
		Xstrval  string
	}

	argv := []string{"--intval", intstr, "--uintval", uintstr, "--byteval", bytestr, "--realval", realstr, "--boolval", "--strval", strval}
	arguments, _ := Parse(usage, argv, true, "", false)

	var opts Options
	err := Cast(&opts, arguments)

	if err != nil {
		t.Errorf("Cast: %v", err)
	}
	
	if opts.Xintval != intval {
		t.Errorf("intval: expected %v got %v", intval, opts.Xintval)
	}
	if opts.Xuintval != uintval {
		t.Errorf("uintval: expected %v got %v", uintval, opts.Xuintval)
	}
	if opts.Xbyteval != byteval {
		t.Errorf("byteval: expected %v got %v", byteval, opts.Xbyteval)
	}
	if opts.Xboolval != true {
		t.Errorf("boolval: expected %v got %v", true, opts.Xboolval)
	}
	if opts.Xstrval != strval {
		t.Errorf("strval: expected %v got %v", strval, opts.Xstrval)
	}

}
