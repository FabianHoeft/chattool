package main

import (
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func haspassword() bool {
	_, err := os.Stat("saves/password.txt")
	if err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		os.Remove("saves/password.txt")
		return false
	}
}
func writepassword(psw string) {
	ok := haspassword()
	if ok {
		return
	}
	file, err := os.Create("saves/password.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	hash := SHA256(psw)
	hashstring := ""
	for _, i := range hash {
		hashstring = hashstring + strconv.FormatUint(uint64(i), 16)
	}

	_, err = file.WriteString(hashstring)
	if err != nil {
		fmt.Println(err)
	}
}

func checkpassword(psw string) bool {
	file, err := os.Open("saves/password.txt")
	if err != nil {
		return false
	}
	defer file.Close()
	data := make([]byte, 65)
	count, err := file.Read(data)
	if err != nil {
		return false
	}
	hash := SHA256(psw)
	hashstring := ""
	for _, i := range hash {
		hashstring = hashstring + strconv.FormatUint(uint64(i), 16)
	}
	if hashstring == string(data[:count]) {
		return true
	}
	return false
}

func load(loc string) (string, error) {
	file, err := os.Open(loc)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	len := info.Size()
	data := make([]byte, len)
	count, err := file.Read(data)
	if err != nil {
		return "", err
	}
	return string(data[:count]), nil
}

func store(loc string, data string) error {
	_, err := os.Stat(loc)
	if err == nil {
	} else if os.IsNotExist(err) {
		os.Remove(loc)
	} else {
		os.Remove(loc)
	}
	file, err := os.Create(loc)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = file.WriteString(data)
	return err
}

// func(P person) save_to_disk() {
//   _=os.Remove("saves/friends.txt")
//   file, err := os.Create("saves/password.txt")
//   if err != nil {
//     fmt.Println(err)
//   }
//   out=""
//   for _,fr := range P.friends {
//     ident:=fr.name
//     frstr=ident+";;"+fr.Network+";;"+fr.IP+";;"+strconv.FormatUint(uint64(fr.status),10)+";;"+fr.keypriv.to_string()
//   }
// }

func deletefiles() {
	_ = os.Remove("saves/password.txt")
	_ = os.Remove("saves/friends.txt")
	_ = os.Remove("saves/chats.txt")
	_ = os.Remove("saves/log.txt")
}

type mm struct {
	a int
	b bool
	c int
}

func printObject(obj interface{}) string {
	switch o := obj.(type) {
	case int:
		return "int{" + strconv.FormatInt(int64(o), 10) + "}"
	case int64:
		return "int64{" + strconv.FormatInt(int64(o), 10) + "}"
	case int32:
		return "int32{" + strconv.FormatInt(int64(o), 10) + "}"
	case uint:
		return "uint{" + strconv.FormatUint(uint64(o), 10) + "}"
	case uint64:
		return "uint64{" + strconv.FormatUint(uint64(o), 10) + "}"
	case uint32:
		return "uint32{" + strconv.FormatUint(uint64(o), 10) + "}"
	case bool:
		if o {
			return "bool{true}"
		}
		return "bool{false}"
	case string:
		return "string{" + o + "}"
	case big.Int:
		return "bigInt{" + (&o).Text(10) + "}"
	case mm:
		return "mm{" + printObject(o.a) + "," + printObject(o.b) + "}"
	case mod:
		return "mod{" + printObject(o.value) + "," + printObject(o.mod) + "," + printObject(o.root) + "}"
	case AES256key:
		return "AES256key{" + printObject(o.value[0]) + "," + printObject(o.value[1]) + "," + printObject(o.value[2]) + "," + printObject(o.value[3]) + "," + printObject(o.value[4]) + "," + printObject(o.value[5]) + "," + printObject(o.value[6]) + "," + printObject(o.value[7]) + "}"
	case friend:
		return "friend{" + printObject(o.name) + "," + printObject(o.Network) + "," + printObject(o.IP) + "," + printObject(o.status) + "," + printObject(o.keypriv) + "," + printObject(o.keycom) + "," + printObject(o.unread) + "}"
	case []int:
		out := "[]int{"
		for _, oi := range o {
			out = out + strconv.FormatInt(int64(oi), 10) + ","
		}
		return out[:len(out)-1] + "}"
	case []int64:
		out := "[]int64{"
		for _, oi := range o {
			out = out + strconv.FormatInt(int64(oi), 10) + ","
		}
		return out[:len(out)-1] + "}"
	case []int32:
		out := "[]int32{"
		for _, oi := range o {
			out = out + strconv.FormatInt(int64(oi), 10) + ","
		}
		return out[:len(out)-1] + "}"
	case []uint:
		out := "[]uint{"
		for _, oi := range o {
			out = out + strconv.FormatUint(uint64(oi), 10) + ","
		}
		return out[:len(out)-1] + "}"
	case []uint64:
		out := "[]uint64{"
		for _, oi := range o {
			out = out + strconv.FormatUint(uint64(oi), 10) + ","
		}
		return out[:len(out)-1] + "}"
	case []uint32:
		out := "[]uint32{"
		for _, oi := range o {
			out = out + strconv.FormatUint(uint64(oi), 10) + ","
		}
		return out[:len(out)-1] + "}"
	case []uint8:
		out := "[]uint8{"
		for _, oi := range o {
			out = out + strconv.FormatUint(uint64(oi), 10) + ","
		}
		return out[:len(out)-1] + "}"
	case []string:
		out := "[]string{"
		for _, oi := range o {
			out = out + oi + "\n"
		}
		return out[:len(out)-1] + "}"
	case []big.Int:
		out := "[]bigInt{"
		for _, oi := range o {
			out = out + oi.Text(10) + ","
		}
		return out[:len(out)-1] + "}"
	case []mm:
		out := "[]mm{"
		for _, oi := range o {
			out = out + printObject(oi) + ","
		}
		return out[:len(out)-1] + "}"
	case []friend:
		out := "[]friend{"
		for _, oi := range o {
			out = out + printObject(oi) + ","
		}
		return out[:len(out)-1] + "}"
	case nil:
		return "nil{nil}"
	default:
		fmt.Printf("unknown type to print: %T\n", o)
		return "nil{nil}"
	}
}

func read(input string) (out interface{}, err error) {
	structs := "mm|friend|mod|AES256key"
	basetype := "int|int64|int32|uint|uint64|uint32|uint8|bool|string|nil|bigInt"
	initregex, _ := regexp.Compile("(|\\[\\])(" + basetype + ")\\{([^\\{]+?\\})")
	conregex, _ := regexp.Compile("(|\\[\\])(" + structs + ")\\{$")
	var backtrace func(string) ([]interface{}, int, int, error)
	var forwardtrace func(string, []interface{}) (interface{}, int, int, error)
	var trace func(string) (interface{}, int, int, error)
	var findinit func(string) (interface{}, int, int, error)

	findinit = func(s string) (interface{}, int, int, error) {
		m := initregex.FindStringSubmatchIndex(s)
		if len(m) == 0 {
			return nil, 0, 0, &MyError{"not type to init :" + s}
		}
		if s[m[2]:m[3]] == "" {
			switch s[m[4]:m[5]] {
			case "int":
				out, _ := strconv.ParseInt(s[m[6]:m[7]-1], 0, 64)
				return int(out), m[0], m[1], nil
			case "int64":
				out, _ := strconv.ParseInt(s[m[6]:m[7]-1], 0, 64)
				return int64(out), m[0], m[1], nil
			case "int32":
				out, _ := strconv.ParseInt(s[m[6]:m[7]-1], 0, 32)
				return int32(out), m[0], m[1], nil
			case "uint":
				out, _ := strconv.ParseUint(s[m[6]:m[7]-1], 0, 64)
				return uint(out), m[0], m[1], nil
			case "uint64":
				out, _ := strconv.ParseUint(s[m[6]:m[7]-1], 0, 64)
				return uint64(out), m[0], m[1], nil
			case "uint32":
				out, _ := strconv.ParseUint(s[m[6]:m[7]-1], 0, 32)
				return uint32(out), m[0], m[1], nil
			case "uint8":
				out, _ := strconv.ParseUint(s[m[6]:m[7]-1], 0, 8)
				return uint8(out), m[0], m[1], nil
			case "bool":
				out, _ := strconv.ParseBool(s[m[6] : m[7]-1])
				return out, m[0], m[1], nil
			case "string":
				return s[m[6] : m[7]-1], m[0], m[1], nil
			case "nil":
				return nil, m[0], m[1], nil
			case "bigInt":
				var out1 big.Int
				out2, _ := (out1).SetString(s[m[6]:m[7]-1], 0)
				return *out2, m[0], m[1], nil
			default:
				return nil, 0, 0, &MyError{"basetype not implemented :" + s[m[4]:m[5]]}
			}
		} else {
			values := strings.Split(s[m[6]:m[7]-1], ",")
			switch s[m[4]:m[5]] {
			case "int":
				out := make([]int, len(values))
				for i, value := range values {
					temp, _ := strconv.ParseInt(value, 0, 64)
					out[i] = int(temp)
				}
				return out, m[0], m[1], nil
			case "int64":
				out := make([]int64, len(values))
				for i, value := range values {
					temp, _ := strconv.ParseInt(value, 0, 64)
					out[i] = int64(temp)
				}
				return out, m[0], m[1], nil
			case "int32":
				out := make([]int32, len(values))
				for i, value := range values {
					temp, _ := strconv.ParseInt(value, 0, 32)
					out[i] = int32(temp)
				}
				return out, m[0], m[1], nil
			case "uint":
				out := make([]uint, len(values))
				for i, value := range values {
					temp, _ := strconv.ParseUint(value, 0, 64)
					out[i] = uint(temp)
				}
				return out, m[0], m[1], nil
			case "uint64":
				out := make([]uint64, len(values))
				for i, value := range values {
					temp, _ := strconv.ParseUint(value, 0, 64)
					out[i] = uint64(temp)
				}
				return out, m[0], m[1], nil
			case "uint32":
				out := make([]uint32, len(values))
				for i, value := range values {
					temp, _ := strconv.ParseUint(value, 0, 32)
					out[i] = uint32(temp)
				}
				return out, m[0], m[1], nil
			case "uint8":
				out := make([]uint8, len(values))
				for i, value := range values {
					temp, _ := strconv.ParseUint(value, 0, 8)
					out[i] = uint8(temp)
				}
				return out, m[0], m[1], nil
			case "bool":
				out := make([]bool, len(values))
				for i, value := range values {
					out[i], _ = strconv.ParseBool(value)
				}
				return out, m[0], m[1], nil
			case "string":
				values := strings.Split(s[m[6]:m[7]-1], "\n")
				out := make([]string, len(values))
				for i, value := range values {
					out[i] = value
				}
				return out, m[0], m[1], nil
			case "bigInt":
				out := make([]big.Int, len(values))
				var outi *big.Int
				for i, value := range values {
					temp, _ := outi.SetString(value, 0)
					out[i] = *temp
				}
				return out, m[0], m[1], nil
			default:
				return nil, 0, 0, &MyError{"cant make directly :[]" + s[m[4]:m[5]]}
			}
		}
	}

	forwardtrace = func(s string, temp []interface{}) (interface{}, int, int, error) {
		m := conregex.FindStringSubmatchIndex(s)
		if len(m) > 0 {
			if s[m[2]:m[3]] == "[]" {
				switch s[m[4]:m[5]] {
				case "mm":
					out := make([]mm, len(temp))
					for i, tempi := range temp {
						out[i] = tempi.(mm)
					}
					return out, m[0], m[1], nil
				case "friend":
					out := make([]friend, len(temp))
					for i, tempi := range temp {
						out[i] = tempi.(friend)
					}
					return out, m[0], m[1], nil
				case "mod":
					out := make([]mod, len(temp))
					for i, tempi := range temp {
						out[i] = tempi.(mod)
					}
					return out, m[0], m[1], nil
				case "AES256key":
					out := make([]AES256key, len(temp))
					for i, tempi := range temp {
						out[i] = tempi.(AES256key)
					}
					return out, m[0], m[1], nil
				default:
					return nil, 0, 0, &MyError{"invalid struct :[]" + s[m[2]:m[3]]}
				}
			} else {
				switch s[m[4]:m[5]] {
				case "mm":
					return mm{temp[0].(int), temp[1].(bool), temp[2].(int)}, m[0], m[1], nil
				case "mod":
					return mod{temp[0].(big.Int), temp[1].(big.Int), temp[2].(big.Int)}, m[0], m[1], nil
				case "AES256key":
					return AES256key{[8]uint32{temp[0].(uint32), temp[1].(uint32), temp[2].(uint32), temp[3].(uint32), temp[4].(uint32), temp[5].(uint32), temp[6].(uint32), temp[7].(uint32)}}, m[0], m[1], nil
				case "friend":
					var k key
					if temp[5] != nil {
						k = temp[5].(key)
					}
					return friend{temp[0].(string), temp[1].(string), temp[2].(string), temp[3].(uint32), temp[4].(mod), k, temp[6].(bool)}, m[0], m[1], nil
				default:
					return nil, 0, 0, &MyError{"not implemented struct :" + s[m[2]:m[3]]}
				}
			}
		} else {
			return nil, 0, 0, &MyError{"cant forwardtrace on: " + s}
		}
	}

	trace = func(s string) (interface{}, int, int, error) {
		out, st, en, err := findinit(s)
		if err != nil {
			return nil, 0, 0, err
		}
		if st == 0 {
			return out, st, en, nil
		}
		var temp []interface{}
		endglobal := en
		for st != 0 {
			temp, _, en, err = backtrace(s[endglobal:])
			endglobal += en + 1
			if err != nil {
				return nil, 0, 0, err
			}
			temp = append([]interface{}{out}, temp...)
			out, st, _, err = forwardtrace(s[:st], temp)
			if err != nil {
				return nil, 0, 0, err
			}
		}
		return out, st, endglobal, nil
	}

	backtrace = func(s string) ([]interface{}, int, int, error) {
		temp := make([]interface{}, 0)
		en := 0
		for s[en:en+1] == "," {
			va1, _, en1, err := trace(s[en+1:] + " ")
			if err != nil {
				return nil, 0, 0, err
			}
			temp = append(temp, va1)
			en = en + en1 + 1
		}
		return temp, 0, en, nil
	}

	recover := func() {
		if r := recover(); r != nil {
			out, err = 0, &MyError{fmt.Sprint(r)}
		}
	}
	defer recover()
	out, _, _, err = trace(input)
	return out, err
}
