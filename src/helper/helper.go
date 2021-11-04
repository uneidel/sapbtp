package helper

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"unicode"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {

		return false
	}
	return !info.IsDir()
}

func ReadJson(path string) ([]byte, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	return byteValue, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
func FlattenAndOmit(m map[string]interface{}, omitinternal bool) map[string]interface{} {
	o := make(map[string]interface{})
	for k, v := range m {
		key := []rune(k)
		if unicode.IsLower(key[0]) && omitinternal {
			//log.Printf("Key: %s", key)
			continue
		}
		//log.Printf("Public")
		switch child := v.(type) {
		case map[string]interface{}:
			nm := FlattenAndOmit(child, omitinternal)
			for nk, nv := range nm {
				o[k+"_"+nk] = nv
			}
		default:
			o[k] = v
		}
	}
	return o
}
func GetStringFromMap(res map[string]interface{}, key string) (string, error) {
	var val string
	var ok bool

	if x, found := res[key]; found == true {
		if val, ok = x.(string); !ok {
			return "", errors.New("val is not of type string")
		}
	} else {
		return "", errors.New("key not found")
	}
	return val, nil
}
func RemoveColon(ipv6 string) string {
	return strings.ReplaceAll(ipv6, ":", "")
}
