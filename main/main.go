package main

import (
	"bufio"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	rand.Seed(time.Now().Unix())

	inputFile := "input"

	//cleanEmails, err := os.Open(clean(inputFile))
	//check(err)
	//defer cleanEmails.Close()
	input, err := os.Open(clean(inputFile))
	check(err)
	defer input.Close()

	scanner := bufio.NewScanner(input)

	for i := 1; i <= 3; i++ {
		filterfile := "filter" + strconv.Itoa(i)
		ff, err := os.Create(filterfile)
		check(err)
		defer ff.Close()

		for scanner.Scan() {
			line := scanner.Text()
			if rand.Intn(1000) < 150 {
				ff.Write([]byte(line + "\n"))
			} else {
				ff.Write([]byte(emailOrShit(70)))
			}

		}
	}

}

func clean(filename string) string {
	emailRegexp := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	generateInput(filename)
	read, err := os.Open(filename)
	check(err)
	defer read.Close()

	output := "clean" + strings.Title(filename)
	write, err := os.Create(output)
	check(err)
	defer write.Close()

	scanner := bufio.NewScanner(read)

	for scanner.Scan() {
		line := scanner.Text()
		if emailRegexp.MatchString(line) {
			write.Write([]byte(line + "\n"))
		}
	}

	return output
}

func generateInput(filename string) string {
	f, err := os.Create(filename)
	check(err)
	defer f.Close()

	for i := 0; i < 1000; i++ {
		s := emailOrShit(70)
		f.Write([]byte(s + "\n"))
	}

	return filename
}

func emailOrShit(probability int) string {
	var s string
	if rand.Intn(100) < probability {
		s = randomEmail()
	} else {
		s = randomString(randomInt(1, 52))
	}
	return s + "\n"
}

func randomString(n int) string {
	return randomStringScope(n, "abcdefghijklmnopqrstuvwxyz") // ABCDEFGHIJKLMNOPQRSTUVWXYZ
}

func randomStringScope(n int, scope string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = scope[rand.Intn(len(scope))]
	}
	return string(b)
}

func randomEmail() string {
	tlds := []string{".com", ".org", ".net"}

	namelength := randomInt(5, 16)
	name := randomString(namelength)

	domainlength := randomInt(4, 8)
	domain := randomString(domainlength)

	tld := tlds[rand.Intn(len(tlds))]

	return name + "@" + domain + tld

}

func randomInt(min, max int) int {
	return rand.Intn(max-min) + min
}
