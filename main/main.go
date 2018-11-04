package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

func main() {
	//supress -input=input -filters="filter1 filter2 filter3"
	//supress -input="input" -filters="f0 f1 f2 f3 f4 f5 f6 f7 f8 f9"
	//supress -input="D:\\go\\input" -filters="D:\\go\\f0 f1 f2 f3 f4 f5 f6 f7 f8 f9"

	input := flag.String("input", "", "Input file")
	filters := flag.String("filters", "", "Filter files")
	//threads := 10                        //   flag.Int("threads", 1, "Threads")
	flag.Parse()
	checkFlags(*input, *filters)

	var filterMap sync.Map

	start := time.Now()
	filterPreProcessing(*filters, &filterMap)
	searchInputInFilters(*input, &filterMap)
	defer fmt.Println("Processing took: " + time.Since(start).String())

}
func filterPreProcessing(filters string, filterMap *sync.Map) {
	var filterWG sync.WaitGroup
	filterFiles := strings.Split(filters, " ")
	filterWG.Add(len(filterFiles))
	for _, filterFile := range filterFiles {
		go func(filterFile string, wg *sync.WaitGroup, filterMap *sync.Map) {
			defer filterWG.Done()

			file, err := os.Open(filterFile)
			check(err)
			defer file.Close()

			fileScanner := bufio.NewScanner(file)

			for fileScanner.Scan() {
				fileLine := fileScanner.Text()
				if isEmail(fileLine) {
					filterMap.Store(GetMD5Hash(fileLine), true)
				}
				if isMD5(fileLine) {
					filterMap.Store(fileLine, false)
				}

			}
		}(filterFile, &filterWG, filterMap)

	}

	filterWG.Wait()
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
		case matchLine, ok := <-match:
			if ok {
				m.WriteString(matchLine + "\n")
			}
		case cleanline, ok := <-clean:
			if ok {
				c.WriteString(cleanline + "\n")
			}
		}
	}

}

func searchInputInFilters(inputFileName string, filterMap *sync.Map) {
	var wg sync.WaitGroup
	inputChannel := make(chan string, 10946)
	matchChannel := make(chan string, 10946)
	cleanChannel := make(chan string, 10946)

	//read input to inputChannel
	wg.Add(1)
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(matchChannel)
		defer close(cleanChannel)

		for line := range inputChannel {
			linehash := GetMD5Hash(line)

			_, ok := filterMap.Load(line)
			if ok {
				matchChannel <- line
				filterMap.Delete(line)
				continue
			}

			_, ok = filterMap.Load(linehash)
			if ok {
				matchChannel <- line
				filterMap.Delete(line)
				continue
			}

			cleanChannel <- line
		}
	}()

	wg.Wait()
}

func checkFlags(input string, filters string) {
	if len(input) == 0 || len(filters) == 0 {
		panic("Invalid input")
	}
}

func check(e error) {
	if e != nil {
		panic(e)
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

func isMD5(md5 string) bool {
	md5Regexp := regexp.MustCompile("^[a-f0-9]{32}$")
	return md5Regexp.MatchString(md5)
}
