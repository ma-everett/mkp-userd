/* mkp-userd/protocol/protocol.go */
package userd

import (
	"errors"
	"fmt"
	"strings"
)

var (
	InvalidData = errors.New("Invalid Data")
)

func MakeTCheck(key string) []byte {

	return []byte(fmt.Sprintf("CHECK %s", key))
}

func MakeRCheck(status bool) []byte {

	str := "NO"
	if status {
		str = "YES"
	}

	return []byte(fmt.Sprintf("CHECK %s", str))
}

func IsCheckValid(d []byte) (bool, error) {

	parts := strings.Split(string(d), " ")
	if len(parts) != 2 {
		return false, InvalidData
	}

	if parts[0] != "CHECK" {
		return false, InvalidData
	}

	if parts[1] == "YES" {
		return true, nil
	}

	return false, nil
}

func MakeTSet(key string) []byte {

	return []byte(fmt.Sprintf("SET %s", key))
}

func MakeRSet(status bool) []byte {

	str := "FAILED"
	if status {
		str = "OK"
	}

	return []byte(fmt.Sprintf("SET %s", str))
}

func IsSetValid(d []byte) (bool, error) {

	parts := strings.Split(string(d), " ")
	if len(parts) != 2 {
		return false, InvalidData
	}

	if parts[0] != "SET" {
		return false, InvalidData
	}

	if parts[1] == "OK" {
		return true, nil
	}

	return false, nil
}

func MakeTRemove(key string) []byte {

	return []byte(fmt.Sprintf("REMOVE %s", key))
}

func MakeRRemove(status bool) []byte {

	str := "FAILED"
	if status {
		str = "OK"
	}

	return []byte(fmt.Sprintf("REMOVE %s", str))
}

func IsRemoveValid(d []byte) (bool, error) {

	parts := strings.Split(string(d), " ")
	if len(parts) != 2 {
		return false, InvalidData
	}

	if parts[0] != "REMOVE" {
		return false, InvalidData
	}

	if parts[1] == "OK" {
		return true, nil
	}

	return false, nil
}

func MakeTPurge() []byte {

	return []byte("PURGE")
}

func MakeRPurge(status bool) []byte {

	str := "FAILED"
	if status {
		str = "OK"
	}

	return []byte(fmt.Sprintf("PURGE %s", str))
}

func IsPurgeValid(d []byte) (bool, error) {

	parts := strings.Split(string(d), " ")
	if len(parts) != 2 {
		return false, InvalidData
	}

	if parts[0] != "PURGE" {
		return false, InvalidData
	}

	if parts[1] == "OK" {
		return true, nil
	}

	return false, nil
}
