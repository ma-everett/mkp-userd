
/* mkp-userd/protocol/protocol.go */
package userd

import (
	"strings"
	"fmt"
	"errors"
)

var (
	InvalidData = errors.New("Invalid Data")
)

func MakeTCheck(key string) []byte {

	return []byte(fmt.Sprintf("CHECK %s",key))
}

func MakeRCheck(status bool) []byte {

	str := "NO"
	if status {
		str = "YES"
	}

	return []byte(fmt.Sprintf("CHECK %s",str))
}

func IsCheckValid(d []byte) (bool,error) {

	parts := strings.Split(string(d)," ")
	if len(parts) != 2 {
		return false,InvalidData
	}

	if parts[0] != "CHECK" {
		return false,InvalidData
	}

	if parts[1] == "YES" {
		return true,nil
	}

	return false,nil
}
