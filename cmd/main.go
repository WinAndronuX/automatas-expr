package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"expr/pkg/gic"
)

var stdin = bufio.NewReader(os.Stdin)

func readLine(prompt string) string {

	fmt.Print(prompt)

	line, err := stdin.ReadString('\n')
	if err != nil && line == "" {
		log.Fatal(err)
	}

	return strings.TrimSpace(line)
}

func main() {

	var prog, err = gic.Compile(createGIC(), 1000)
	if err != nil {
		log.Fatal(err)
	}

	for {
		var text = readLine("Ingrese la cadena a validar: ")

		if err := prog.Eval(text); err != nil {
			fmt.Printf("Invalid Expression: %s.\n", err.Error())
		} else {
			fmt.Printf("Valid Expression.\n")
		}

		if strings.EqualFold(readLine("Continue? y/N: "), "n") {
			break
		}
	}
}

func createGIC() *gic.GIC {

	var nn = readLine("Ingrese su(s) nombre(s): ")
	var names = strings.Split(strings.ToLower(nn), " ")

	var ln = readLine("Ingrese sus apellidos: ")
	var lastnames = strings.Split(strings.ToLower(ln), " ")

	var sId = readLine("Ingrese su matricula: ")

	var lnsInitials = getIntials(lastnames)
	var fname = names[0]

	return &gic.GIC{
		VT: getAlphabet(lnsInitials, fname, sId),
		VN: []rune{'S', 'M', 'W', 'I', 'J'},
		P: gic.Productions{
			'S': "M Wⁿ M I²ⁿ J²",
			'M': sId,
			'W': lnsInitials,
			'I': reverseString(lnsInitials),
			'J': fname,
		},
		S: 'S',
	}
}

func getIntials(input []string) string {
	var sb strings.Builder

	for _, s := range input {

		if len(s) == 0 {
			continue
		}

		sb.WriteByte(s[0])
	}

	return sb.String()
}

func reverseString(s string) string {
	r := []rune(s)

	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}

	return string(r)
}

func getAlphabet(words ...string) []rune {
	var result = []rune{}

	for _, w := range words {
		for _, r := range w {
			if !slices.Contains(result, r) {
				result = append(result, r)
			}
		}
	}

	return result
}
