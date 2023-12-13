package response

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	decodeError = errors.New("cannot decode to a nil container")
)

// Decode is a helper function that translates a byte slice into a target by
// converting its type.
func Decode(value []byte, out any) error {
	if out == nil {
		return decodeError
	}

	switch v := out.(type) {
	case *[]string:
		if v == nil {
			return decodeError
		}

		p := strings.Split(string(value), ",")
		*v = p

	case *string:
		if v == nil {
			return decodeError
		}

		*v = strings.Clone(string(value))

	case *bool:
		if v == nil {
			return decodeError
		}

		b, err := strconv.ParseBool(string(value))
		if err != nil {
			return err
		}
		*v = b

	case *int:
		if v == nil {
			return decodeError
		}

		i, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return err
		}
		*v = int(i)

	case *int32:
		if v == nil {
			return decodeError
		}

		i, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return err
		}
		*v = int32(i)

	case *int64:
		if v == nil {
			return decodeError
		}

		i, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return err
		}
		*v = i

	case *uint:
		if v == nil {
			return decodeError
		}

		i, err := strconv.ParseUint(string(value), 10, 64)
		if err != nil {
			return err
		}
		*v = uint(i)

	case *uint32:
		if v == nil {
			return decodeError
		}

		i, err := strconv.ParseUint(string(value), 10, 64)
		if err != nil {
			return err
		}
		*v = uint32(i)

	case *uint64:
		if v == nil {
			return decodeError
		}

		i, err := strconv.ParseUint(string(value), 10, 64)
		if err != nil {
			return err
		}
		*v = i

	case *float32:
		if v == nil {
			return decodeError
		}

		f, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			return err
		}
		*v = float32(f)

	case *float64:
		if v == nil {
			return decodeError
		}

		f, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			return err
		}
		*v = f

	default:
		return errors.New("unsupported type conversion")
	}

	return nil
}

// Zero clears the value of the argument to its zero value according its
// type.
func Zero(value interface{}) {
	switch v := value.(type) {
	case *string:
		*v = ""
	case *int:
		*v = 0
	case *int32:
		*v = int32(0)
	case *int64:
		*v = int64(0)
	case *uint:
		*v = uint(0)
	case *uint32:
		*v = uint32(0)
	case *uint64:
		*v = uint64(0)
	case *float32:
		*v = float32(0)
	case *float64:
		*v = float64(0)
	case *bool:
		*v = false
	case *byte:
		*v = byte(0)
	case *[]string:
		*v = []string{}
	case *[]int:
		*v = []int{}
	case *[]int32:
		*v = []int32{}
	case *[]int64:
		*v = []int64{}
	case *[]uint:
		*v = []uint{}
	case *[]uint32:
		*v = []uint32{}
	case *[]uint64:
		*v = []uint64{}
	case *[]float32:
		*v = []float32{}
	case *[]float64:
		*v = []float64{}
	case *[]byte:
		*v = []byte{}
	case *time.Time:
		*v = time.Time{}
	case *map[string]interface{}:
		*v = make(map[string]interface{})
	default:
		zero := reflect.New(reflect.TypeOf(value).Elem())
		p := reflect.ValueOf(v)
		q := p.Elem()
		q.Set(zero.Elem())
	}
}
