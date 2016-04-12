package docopt

import (
	"testing"
	"reflect"
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
  cast_test [options] <values>...

Options:
  --intval=N       Sets intval
  --uintval=N      Sets uintval
  --byteval=N      Sets byteval, optionally add suffix K/M/G/T for kilobytes etc
  --realval=X      Sets realval
  -b               Sets boolval to true
  --strval=string  Sets strval to string

`

	type options struct{
		L_intval  int
		L_uintval uint64
		S_b       bool
		L_byteval uint32
		L_realval float32
		L_strval  string
		A_values  []int
	}

	argv := []string{"--intval", intstr, "--uintval", uintstr, "--byteval", bytestr, "--realval", realstr, "-b", "--strval", strval, "1", "1k" }
	arguments, _ := Parse(usage, argv, true, "", false)

	var opts options
	err := Cast(&opts, arguments)

	if err != nil {
		t.Errorf("Cast: %v", err)
	}
	
	if opts.L_intval != intval {
		t.Errorf("intval: expected %v got %v", intval, opts.L_intval)
	}
	if opts.L_uintval != uintval {
		t.Errorf("uintval: expected %v got %v", uintval, opts.L_uintval)
	}
	if opts.L_byteval != byteval {
		t.Errorf("byteval: expected %v got %v", byteval, opts.L_byteval)
	}
	if opts.S_b != true {
		t.Errorf("boolval: expected %v got %v", true, opts.S_b)
	}
	if opts.L_strval != strval {
		t.Errorf("strval: expected %v got %v", strval, opts.L_strval)
	}
	if !reflect.DeepEqual(opts.A_values, []int {1, 1024}) {
		t.Errorf("values: expected %v got %v", []int {1, 1024}, opts.A_values)
	}

}
