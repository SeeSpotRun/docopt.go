package docopt

// this file provides casting of docopt Parse() results into struct fields

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Cast attempts to unmarshall the values in optMap into the fields
// in the structure pointed to by optStruct.  The field name must start
// with 'A_' for an argument, 'L_' for a long (--) flag and 'S_' for a
// short (-) flag.
// Example:
//  	type Options struct{
//		L_intval  int
//		L_boolval bool
//		L_strval  string
//	}
//	var opts Options
//	args, _ := docopt.Parse(os.Args[1:])
//	docopts.Cast(&opts, args)  // will map "--intval" to Xintval,
//	// "--boolval" to Xboolval and "--strval" to Xstrval
//
// If optMap[option] == nil then the corresponding field in optStruct will
// be left unchanged.
// Note that when casting to integer fields, if the corresponding option
// string ends with [B|K|M|G|T] then this will be interpreted as a multiplier
// of 1, 1024, 1024*1024 etc; for example '--min=10K' should set optStruct.Xmin
// to 10240.
func Cast(optStruct interface{}, optMap map[string]interface{}) error {

	errstr := ""

	ps := reflect.ValueOf(optStruct)
	vs := reflect.Indirect(ps)
	ts := vs.Type()

	if ts.Kind() != reflect.Struct || ps.Kind() == reflect.Struct {
		return errors.New(fmt.Sprintf("Cast: expected *struct for optStruct, got %v", ps.Kind()))
	}

	for i := 0; i < ts.NumField(); i++ {

		f := ts.Field(i)
		v := vs.Field(i)
		if !v.CanSet() {
			errstr += "docopt.Cast: can't set field " + f.Name + "\n"
			continue
		}

		var prefix, suffix string
		switch f.Name[:2] {
		case "A_":
			prefix = "<"
			suffix = ">"
		case "S_":
			prefix = "-"
		case "L_":
			prefix = "--"
		}
		if prefix == "" || len(f.Name) < 3 {
			errstr += "docopt.Cast: field'" + f.Name + "' not of format A_*, S_* or L_*\n"
			continue
		}

		opt := prefix + f.Name[2:] + suffix
		o, ok := optMap[opt]
		if !ok {
			errstr += "docopt.Cast: '" + opt + "' not found in optMap\n"
			continue
		}

		if o == nil {
			continue
		}

		topt := reflect.TypeOf(o)
		vopt := reflect.ValueOf(o)

		// try for direct assign:
		if topt.AssignableTo(f.Type) {
			v.Set(reflect.ValueOf(o))
			continue
		}

		// unmarshal tries to set value of 'to' based on content of 'from'
		unmarshall := func(from reflect.Value, to reflect.Value) {

			// try for direct assign:
			if from.Type().AssignableTo(to.Type()) {
				to.Set(from)
				return
			}

			// do it the hard way:
			switch from.Kind() {

			case reflect.String:

				s := from.String()
				fmt.Println("s:", s)

				switch to.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					ival, err := strconv.ParseInt(s, 10, to.Type().Bits())
					if err != nil {
						// try again looking for B/K/M/G/T
						ival, err = getBytes(s, err)
						if err != nil {
							errstr += err.Error() + "\n"
							return
						}
					}
					to.SetInt(ival)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					uval, err := strconv.ParseUint(s, 10, to.Type().Bits())
					if err != nil {
						// try again looking for B/K/M/G/T
						ival, e := getBytes(s, err)
						if e != nil {
							errstr += e.Error() + "\n"
							return
						}
						uval = uint64(ival)
					}
					to.SetUint(uval)
				case reflect.Float32, reflect.Float64:
					fval, err := strconv.ParseFloat(s, to.Type().Bits())
					if err != nil {
						errstr += err.Error() + "\n"
						return
					}
					to.SetFloat(fval)
				default:
					errstr += fmt.Sprintf("Unhandled destination kind: %v\n", to.Kind())
					return
				}
			default:
				errstr += fmt.Sprintf("Don't know how to unmarshall %v to %v\n",
					from.Kind(),
					to.Kind())
				return
			}
		}

		if topt.Kind() != reflect.Slice {
			unmarshall(vopt, vs.Field(i))
			continue
		}

		// handle slices:
		if vs.Field(i).Kind() == reflect.Slice {
			// do a deep unmarshall
			vs.Field(i).Set(reflect.MakeSlice(vs.Field(i).Type(), vopt.Len(), vopt.Len()))
			for j := 0; j < vopt.Len(); j++ {
				unmarshall(vopt.Index(j), vs.Field(i).Index(j))
			}
		} else if vopt.Len() == 1 {
			// maps length==1 slice to single field
			unmarshall(vopt.Index(0), vs.Field(i))
		}

	}
	if len(errstr) > 0 {
		return errors.New(errstr[:len(errstr)-1]) // strips trailling newline
	}
	return nil
}

func getBytes(s string, err error) (int64, error) {
	const (
		b = 1
		k = b << 10
		m = k << 10
		g = m << 10
		t = g << 10
	)
	var mult int64
	switch strings.ToUpper(string(s[len(s)-1])) {
	case "B":
		mult = b
	case "K":
		mult = k
	case "M":
		mult = m
	case "G":
		mult = g
	case "T":
		mult = t
	default:
		return 0, err
	}
	ival, err := strconv.ParseInt(s[:len(s)-1], 10, 64)
	if err != nil {
		fval, err := strconv.ParseFloat(s[:len(s)-1], 64)
		if err != nil {
			return 0, err
		}
		return int64(fval * float64(mult)), nil
	}
	return ival * mult, nil
}
