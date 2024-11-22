package main

import (
	"bytes"
	"fmt"
	"os"
)

func main() {
	var mbs int
	fmt.Print("mbs to add: ")
	fmt.Scan(&mbs)
	d, _ := os.ReadFile("./sryxen-built.exe")
	out, _ := os.Create("srxen-pumped.exe")
	defer out.Close()
	out.Write(d)
	out.Write(bytes.Repeat([]byte{0}, mbs*1024*1024))
}
