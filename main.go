package main

import (
	"fmt"
	"unicode"
)

func isNum(str []byte) bool {
	for i := 0; i < len(str); i++ {
		unicode.IsDigit(str + i)
	}
	return
}

func main() {
	ret := isNum([]byte("123123123"))
	fmt.Println(ret)
}
