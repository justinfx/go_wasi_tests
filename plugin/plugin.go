package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func CountVowels(input []byte) []byte {
	count := 0
	for _, a := range input {
		switch a {
		case 'A', 'I', 'E', 'O', 'U', 'a', 'e', 'i', 'o', 'u':
			count++
		default:
		}
	}

	output := fmt.Sprintf(`{"count": %d}`, count)
	return []byte(output)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "missing arg\n")
		os.Exit(1)
	}
	input := strings.Join(os.Args[1:], " ")
	output := CountVowels([]byte(input))
	io.Copy(os.Stdout, bytes.NewReader(output))
}
