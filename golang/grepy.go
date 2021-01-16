package main

import (
	"strings"
	"regexp"
	"bufio"
	"os"
	"flag"
	"fmt"
	"errors"
	"log"
)

type formatter func(in <-chan lineInfo) <-chan string

func mutuallyExclusiveSet(target **formatter, value formatter) error {
	if *target != nil {
		return errors.New("Cannot set a mutually exclusive param")
	}
	*target = &value
	return nil
}

var (
	format *formatter
	regex string
	files []string
)

func parseArgs() error {
	underline := flag.Bool("underline", false, "Highlight the matches with '^' under the line")
	color := flag.Bool("color", false, "Highlights the matches in color")
	machine := flag.Bool("machine", false, "Print in machine format [file name]:[line number]:[match text]")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s [options] <regex> [filename ...]:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	switch len(flag.Args()) {
	case 0:
		return errors.New("Must pass at least regex")
	case 1:
		regex = flag.Arg(0)
	default:
		regex = flag.Arg(0)
		files = flag.Args()[1:]
	}

	if *underline {
		if err := mutuallyExclusiveSet(&format, underlineFormat); err != nil {
			return fmt.Errorf("underline: %v", err)
		}
	}
	if *color {
		if err := mutuallyExclusiveSet(&format, colorFormat); err != nil {
			return fmt.Errorf("color: %v", err)
		}
	}
	if *machine {
		if err := mutuallyExclusiveSet(&format, machineFormat); err != nil {
			return fmt.Errorf("machine: %v", err)
		}
	}
	mutuallyExclusiveSet(&format, machineFormat)

	return nil
}

type lineInfo struct {
	filename string
	lineNum int
	line string
	matches [][]int
}

func filesReader(in <-chan string) <-chan lineInfo {
	out := make(chan lineInfo)
	go func() {
		defer close(out)

		for name := range in {
			file, err := os.Open(name)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			index := 0
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				index++
				out <- lineInfo{name, index, scanner.Text(), nil}
			}
		}
	}()

	return out
}

func lineMatcher(str string, in <-chan lineInfo) <-chan lineInfo {
	regex := regexp.MustCompile(str)
	out := make(chan lineInfo)
	go func() {
		defer close(out)
		for li := range in {
			li.matches = regex.FindAllStringIndex(li.line, -1)
			if len(li.matches) > 0 {
				out <- li
			}
		}
	}()

	return out
}

func machineFormat(in <-chan lineInfo) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for li := range in {
			out <- fmt.Sprintf("%v:%v:%v", li.filename, li.lineNum, li.line)
		}
	}()

	return out
}

func colorFormat(in <-chan lineInfo) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for li := range in {
			str := fmt.Sprintf("%v (%v) ", li.filename, li.lineNum)
			last := 0
			for _, m := range li.matches {
				str += fmt.Sprintf("%v%v%v%v", li.line[last:m[0]], "\033[32m", li.line[m[0]:m[1]], "\033[39m")
				last = m[1]
			}
			out <- str
		}
	}()

	return out
}

func underlineFormat(in <-chan lineInfo) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for li := range in {
			// print the originl line first
			str := fmt.Sprintf("%v (%v) ", li.filename, li.lineNum)
			l := len(str)
			str += fmt.Sprintf("%v\n", li.line)

			// now print the underlines
			str += strings.Repeat(" ", l)
			last := 0
			for _, m := range li.matches {
				str += strings.Repeat(" ", m[0]-last)
				str += strings.Repeat("^", m[1]-m[0])
				last = m[1]
			}
			out <- str
		}
	}()

	return out
}

func main() {
	err := parseArgs()
	if err != nil {
		fmt.Println(err)
		return
	}

	i := make(chan string)
	c := (*format)(lineMatcher(regex, filesReader(i)))

	for _, name := range(files) {
		i <- name
	}
	close(i)

	for li := range c {
		fmt.Printf("%v\n", li)
	}

}
