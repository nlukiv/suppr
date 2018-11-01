package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/fvbock/trie"
	"os"
	"regexp"
	"strings"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	//input := flag.String("input", "", "Input file")
	//filters := flag.String("filters", "", "Filter files")
	//flag.Parse()

	input := "input"
	filters := "filter1 filter2 filter3"
	//checkFlags(*input, *filters)
	for i := 0; i < 1; i++ {

		startTime := time.Now()

		//supress(input, filters)
		supTree(input, filters)

		fmt.Println("Processing took " + time.Since(startTime).String())

	}
}

func supTree(input string, filters string) {

	inputFile, err := os.Open(input)
	check(err)
	defer inputFile.Close()

	filterFiles := strings.Split(filters, " ")

	match, err := os.Create("match.txt")
	check(err)
	defer match.Close()

	clean, err := os.Create("clean.txt")
	check(err)
	defer clean.Close()

	filtersTree := trie.NewTrie()
	for _, f := range filterFiles {
		filterFile, err := os.Open(f)
		check(err)
		defer filterFile.Close()

		filterScanner := bufio.NewScanner(filterFile)

		for filterScanner.Scan() {
			filtersTree.Add(filterScanner.Text())
		}
	}
	//filtersTree.PrintDump()

	inputScanner := bufio.NewScanner(inputFile)

	for inputScanner.Scan() {
		line := inputScanner.Text()
		lineHash := GetMD5Hash(line)

		if isEmail(line) {
			if filtersTree.Has(line) || filtersTree.Has(lineHash) {
				match.Write([]byte(line + "\n"))
				filtersTree.Delete(line)
			} else {
				clean.Write([]byte(line + "\n"))
			}
		}
	}
}

func supress(input string, filters string) {
	inputFile, err := os.Open(input)
	check(err)
	defer inputFile.Close()

	filterFiles := strings.Split(filters, " ")

	match, err := os.Create("match.txt")
	check(err)
	defer match.Close()

	clean, err := os.Create("clean.txt")
	check(err)
	defer clean.Close()

	inputScanner := bufio.NewScanner(inputFile)

	for inputScanner.Scan() {
		line := inputScanner.Text()
		linehash := GetMD5Hash(line)
		found := false
		if isEmail(line) {
			for _, f := range filterFiles {
				file, err := os.Open(f)
				check(err)
				defer file.Close()

				fileScanner := bufio.NewScanner(file)

				for fileScanner.Scan() {
					fileLine := fileScanner.Text()

					if fileLine == line || fileLine == linehash {
						match.Write([]byte(line + "\n"))
						found = true
						break
					}
				}
				if found {
					break
				}

			}
			if !found {
				clean.Write([]byte(line + "\n"))
			}
		}
	}

}

func checkFlags(input string, filters string) {
	if len(input) == 0 || len(filters) == 0 {
		panic("Invalid input")
	}
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func isEmail(email string) bool {
	emailRegexp := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return emailRegexp.MatchString(email)
}

//func clean(filename string) string {
//
//	//generateInput(filename)
//	read, err := os.Open(filename)
//	check(err)
//	defer read.Close()
//
//	output := "clean" + strings.Title(filename)
//	write, err := os.Create(output)
//	check(err)
//	defer write.Close()
//
//	scanner := bufio.NewScanner(read)
//
//	for scanner.Scan() {
//		line := scanner.Text()
//		if isEmail(line) {
//			write.Write([]byte(line + "\n"))
//		}
//	}
//
//	return output
//}
//
//func makeBig(filename string, size int64) {
//
//	var currSize int64 = 0
//
//	for currSize < size {
//
//		generateInput(filename)
//
//		fileStat, err := os.Stat(filename)
//		check(err)
//		currSize = fileStat.Size()
//	}
//}
//
//func generateInput(filename string) string {
//	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	check(err)
//
//	for i := 0; i < 100000; i++ {
//		s := emailOrShit(80)
//		f.WriteString(s)
//	}
//	f.Close()
//	return filename
//}
//func generateFilter(filename string, inputfile string) string {
//
//	read, err := os.Open(inputfile)
//	check(err)
//	defer read.Close()
//
//	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	check(err)
//	defer f.Close()
//
//	for i := 0; i < 100000; i++ {
//		s := emailOrShit(80)
//		f.WriteString(s)
//
//	}
//
//	return filename
//}
//
//func cleanize(input string) {
//	read, err := os.Open(input)
//	check(err)
//	defer read.Close()
//
//	f, err := os.OpenFile("D:\\go\\cleanized0", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	check(err)
//	defer f.Close()
//
//	scanner := bufio.NewScanner(read)
//
//	for scanner.Scan() {
//		line := scanner.Text()
//
//		if isEmail(line) && probability(40) {
//			if probability(30) {
//				f.WriteString(GetMD5Hash(line) + "\n")
//			} else {
//				f.WriteString(line + "\n")
//			}
//		}
//	}
//
//}
//
//func probability(percent int) bool {
//	return rand.Intn(100) < percent
//}
//
//func emailOrShit(probability int) string {
//	var s string
//	if rand.Intn(100) < probability {
//		s = randomEmail()
//	} else {
//		s = randomString(randomInt(16, 52))
//	}
//	return s + "\n"
//}
//
//func randomString(n int) string {
//	return randomStringScope(n, "abcdefghijklmnopqrstuvwxyz") // ABCDEFGHIJKLMNOPQRSTUVWXYZ
//}
//
//func randomStringScope(n int, scope string) string {
//	b := make([]byte, n)
//	for i := range b {
//		b[i] = scope[rand.Intn(len(scope))]
//	}
//	return string(b)
//}
//
//func randomEmail() string {
//	tlds := []string{".com", ".org", ".net"}
//
//	namelength := randomInt(5, 16)
//	name := randomStringScope(namelength, "abcdefghijklmnopqrstuvwxyz0123456789")
//
//	domainlength := randomInt(4, 8)
//	domain := randomString(domainlength)
//
//	tld := tlds[rand.Intn(len(tlds))]
//
//	return name + "@" + domain + tld
//
//}
//
//func randomInt(min, max int) int {
//	return rand.Intn(max-min) + min
//}
