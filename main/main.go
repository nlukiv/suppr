package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/fvbock/trie"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	//supress -input=input -filters="filter1 filter2 filter3"
	//supress -input="input" -filters="f0 f1 f2 f3 f4 f5 f6 f7 f8 f9"
	//supress -input="D:\\go\\input" -filters="D:\\go\\f0 f1 f2 f3 f4 f5 f6 f7 f8 f9"

	// split input to different threads
	// remove line from filter if found ?
	//
	//
	input := "input"                     //"D:\\go\\inputSmall" //flag.String("input", "", "Input file")
	filters := "filter1 filter2 filter3" //"D:\\go\\f0"       //flag.String("filters", "", "Filter files")
	//flag.Parse()
	//checkFlags(*input, *filters)
	start := time.Now()
	//supTree(input, filters)

	supressionChannels(input, filters)

	defer fmt.Println("Processing took: " + time.Since(start).String())

}

func writeResults(match <-chan string, clean <-chan string, wg *sync.WaitGroup) {
	m, err := os.OpenFile("match.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	check(err)
	defer m.Close()

	c, err := os.OpenFile("clean.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	check(err)
	defer c.Close()

	defer wg.Done()

	for {
		select {
		case matchLine := <-match:
			m.WriteString(matchLine + "\n")
		case cleanline := <-clean:
			c.WriteString(cleanline + "\n")
		}
	}

}

func supressionChannels(inputFileName string, filters string) {
	var wg sync.WaitGroup
	inputChannel := make(chan string)
	matchChannel := make(chan string)
	cleanChannel := make(chan string)

	wg.Add(2)
	go func() {
		defer wg.Done()
		inputFile, err := os.Open(inputFileName)
		check(err)
		defer inputFile.Close()

		inputScanner := bufio.NewScanner(inputFile)

		for inputScanner.Scan() {
			line := inputScanner.Text()
			if isEmail(line) {
				inputChannel <- line
			}
		}
		close(inputChannel)
	}()

	go writeResults(matchChannel, cleanChannel, &wg)

	go func() {
		defer wg.Done()
		defer close(matchChannel)
		defer close(cleanChannel)

		for line := range inputChannel {

			linehash := GetMD5Hash(line)
			found := false

			filterFiles := strings.Split(filters, " ")
			for _, f := range filterFiles {
				file, err := os.Open(f)
				check(err)
				defer file.Close()

				fileScanner := bufio.NewScanner(file)

				for fileScanner.Scan() {
					fileLine := fileScanner.Text()

					if fileLine == line || fileLine == linehash {
						matchChannel <- line
						//writeResults("match.txt", line)
						//match.Write([]byte(line + "\n"))
						found = true
						break
					}
				}
				if found {
					break
				}

			}
			if !found {
				cleanChannel <- line
				//writeResults("match.txt", line)
				//clean.Write([]byte(line + "\n"))
			}
		}
	}()

	wg.Wait()
}

func findInSingleFilter(wg *sync.WaitGroup, ch_found chan bool, input_email string, input_hash string, filterFile string) {
	defer wg.Done()
	file, err := os.Open(filterFile)
	check(err)
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		select {
		case value := <-ch_found:
			ch_found <- value
			if value {
				return
			}
		default:
			{
				fileLine := fileScanner.Text()

				if fileLine == input_email || fileLine == input_hash {
					ch_found <- true
					return
				}
			}
		}

	}
}

func findInFilters(ch_found chan bool, input_email string, filterFiles []string) {
	linehash := GetMD5Hash(input_email)

	var wg sync.WaitGroup
	defer wg.Wait()
	for _, f := range filterFiles {
		select {
		case value := <-ch_found:
			ch_found <- value
			if value {
				return
			}
		default:
			wg.Add(1)
			go findInSingleFilter(&wg, ch_found, input_email, linehash, f)

		}
	}
}

func supTree(input string, filters string) {

	inputFile, err := os.Open(input)
	check(err)
	defer inputFile.Close()

	filterFiles := strings.Split(filters, " ")

	match, err := os.Create("D:\\go\\match.txt")
	check(err)
	defer match.Close()

	clean, err := os.Create("D:\\go\\clean.txt")
	check(err)
	defer clean.Close()

	filtersTree := trie.NewTrie()
	for _, f := range filterFiles {
		fmt.Println("Adding to a tree: " + f)

		filterFile, err := os.Open(f)
		check(err)
		defer filterFile.Close()
		count := 0
		filterScanner := bufio.NewScanner(filterFile)

		for filterScanner.Scan() {
			filtersTree.Add(filterScanner.Text())
			count++
			fmt.Println("added " + strconv.Itoa(count/47700))
		}

	}
	//filtersTree.PrintDump()

	inputScanner := bufio.NewScanner(inputFile)
	count := 0
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
		count++
	}
	fmt.Println(count / 100000)
}

func supress(input string, filters string) {
	start := time.Now()
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
	finish := time.Now()

	fmt.Println(finish.Sub(start).String())
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
//func makeBig(filename string, forty string) {
//
//	generateFilter(filename, forty)
//
//	//fileStat, err := os.Stat(filename)
//	//check(err)
//	//currSize := fileStat.Size()
//
//}
//
//func generateFilter(output string, forty string) {
//
//	read, err := os.Open(forty)
//	check(err)
//	defer read.Close()
//
//	f, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	check(err)
//	defer f.Close()
//
//	time := 1
//
//	scanner := bufio.NewScanner(read)
//
//	for scanner.Scan() {
//		line := scanner.Text()
//		f.WriteString(line + "\n")
//
//		for i := 0; i < 12; i++ {
//			if probability(70) {
//				f.WriteString(randomEmail() + "\n")
//			} else {
//				f.WriteString(GetMD5Hash(randomEmail()) + "\n")
//			}
//		}
//		time++
//
//		if time%100000==0 {
//
//			fmt.Println(100*time/280000)
//		}
//	}
//}
//
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
//
//func forty(clean string) {
//	read, err := os.Open(clean)
//	check(err)
//	defer read.Close()
//
//	f, err := os.OpenFile("D:\\go\\forty-"+randomString(3), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	check(err)
//	defer f.Close()
//
//	scanner := bufio.NewScanner(read)
//
//	for scanner.Scan() {
//		line := scanner.Text()
//		if probability(40) {
//			f.WriteString(line + "\n")
//		}
//	}
//
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
