package main

import (
	"fmt"

	"github.com/extism/go-pdk"
)

//export count_vowels
func count_vowels() int32 {
	input := pdk.Input()
	output := CountVowels(input)
	mem := pdk.AllocateString(string(output))
	pdk.OutputMemory(mem)

	return 0
}

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

func main() {}
