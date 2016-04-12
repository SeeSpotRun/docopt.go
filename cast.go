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
// in the structure pointed to by optStruct.  The first letter in
// each field must be a capital letter (to make it exportable) but is
// ignored.
// Example:
//  	type Options struct{
//		Xintval  int
//		Xboolval bool
//		Xstrval  string
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

	ps := reflect.ValueOf(optStruct)
	vs := reflect.Indirect(ps)
	ts := vs.Type()

	if ts.Kind() != reflect.Struct || ps.Kind() == reflect.Struct {
		return errors.New(fmt.Sprintf("Cast: expected *struct for optStruct, got %v", ps.Kind()))
	}

	for i := 0; i < ts.NumField(); i++ {

		f := ts.Field(i)
		v := vs.Field(i)
		fmt.Println("f, v:", f, v)
		if !v.CanSet() {
			return errors.New("Cast: can't set field " + f.Name)
		}

		opt := "--" + f.Name[1:]
		o, ok := optMap[opt]
		if !ok {
			return errors.New("docopt.Cast: '" + opt + "' not found in optMap")
		}

		if o == nil {
			continue
		}

		vopt := reflect.ValueOf(o)
		topt := reflect.TypeOf(o)

		// try for direct assign:
		fmt.Println("o, vopt, topt:", o, vopt, topt)
		if topt.AssignableTo(f.Type) {
			v.Set(vopt)
			continue
		}

		// manual unmarshalling:
		switch topt.Kind() {
		case reflect.String:

			s, _ := o.(string)

			switch f.Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				ival, err := strconv.ParseInt(s, 10, f.Type.Bits())
				if err != nil {
					// try again looking for B/K/M/G/T
					fval, e := getBytes(s, err)
					if e != nil {
						return e
					}
					ival = int64(fval)
				}
				vs.Field(i).SetInt(ival)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				uval, err := strconv.ParseUint(s, 10, f.Type.Bits())
				if err != nil {
					// try again looking for B/K/M/G/T
					fval, e := getBytes(s, err)
					if e != nil {
						return e
					}
					uval = uint64(fval)
				}
				vs.Field(i).SetUint(uval)
			case reflect.Float32, reflect.Float64:
				fval, err := strconv.ParseFloat(s, f.Type.Bits())
				if err != nil {
					return err
				}
				vs.Field(i).SetFloat(fval)
			default:
				return errors.New(fmt.Sprintf("Unhandled kind: %v", f.Type.Kind()))
			}
		default:
			return errors.New(fmt.Sprintf("Don't know how to unmarshall %v to %v",
				topt.Kind(),
				f.Type.Kind()))
		}
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
