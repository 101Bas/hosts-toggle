package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

const hostsFile = "/etc/hosts"

func main() {

	// Hide datetime in logs
	log.SetFlags(0)

	// Check for args
	var flagProject = flag.String("p", "", "The project name as defined in your hosts-file")
	flag.Parse()

	var project = strings.Trim(*flagProject, " ")

	if len(project) < 1 {
		log.Fatal("Invalid arguments, use -p to select project")
	}

	// Check for sudo
	if !isSuperUser() {
		log.Fatal("You have to run this program as super-user!")
	}

	// Scan hosts-file
	lines := getHostsFileLines()

	startLineIndex, err := getProjectStartLine(lines, project)
	if err != nil {
		log.Fatal(err)
	}

	endLineIndex, err := getProjectEndLine(lines, startLineIndex)
	if err != nil {
		log.Fatal(err)
	}

	uncommentedLines := []string{}
	commentedLines := []string{}

	// Update
	for i := startLineIndex + 1; i < endLineIndex; i++ {
		line := lines[i]

		if strings.HasPrefix(line, "#") {
			// Remove comment
			line = strings.TrimLeft(line, "#")
			uncommentedLines = append(uncommentedLines, line)
		} else {
			// Add comment
			line = "#" + line
			commentedLines = append(commentedLines, line)
		}

		lines[i] = line
	}

	// Lines to string
	var newContent string
	lineCount := len(lines)

	for i := 0; i < len(lines); i++ {
		newContent += lines[i]
		if i < lineCount-1 {
			newContent += "\n"
		}
	}

	// Write
	err = ioutil.WriteFile(hostsFile, []byte(newContent), 0644)

	if err != nil {
		log.Println("Error writing hosts file...")
		return
	}

	// Summary
	fmt.Printf("Toggling %s..\n", project)

	displayChanges(uncommentedLines, "\033[0;32mUncommented the following lines:\033[0m")
	displayChanges(commentedLines, "\033[0;31mCommented the following lines:\033[0m")
}

func displayChanges(lines []string, message string) {
	if len(lines) > 0 {
		fmt.Println(message)
		for i := 0; i < len(lines); i++ {
			fmt.Printf("\t%s\n", lines[i])
		}
	}
}

func isSuperUser() bool {
	// Retrieve sudo env
	var sudo string
	sudo = os.Getenv("SUDO_USER")
	if len(sudo) < 1 {
		sudo = os.Getenv("SUDO_UID")
	}

	// Check if sudo
	if len(sudo) < 1 {
		return false
	}

	return true
}

func getHostsFileLines() []string {
	file, err := ioutil.ReadFile(hostsFile)
	if err != nil {
		log.Fatal(err)
	}

	s := string(file)
	lines := strings.Split(s, "\n")

	return lines
}

func getProjectStartLine(hosts []string, project string) (int, error) {
	var projectStartRegex = regexp.MustCompile(fmt.Sprintf("^#[ ]?TOGGLE[ ]+%s$", project))

	for i := 0; i < len(hosts); i++ {
		if projectStartRegex.MatchString(hosts[i]) {
			return i, nil
		}
	}

	return -1, errors.New("Project not found")
}

func getProjectEndLine(hosts []string, startLine int) (int, error) {
	var projectEndRegex = regexp.MustCompile("^#[ ]?END[ ]?TOGGLE$")

	for i := startLine; i < len(hosts); i++ {
		if projectEndRegex.MatchString(hosts[i]) {
			return i, nil
		}
	}

	return -1, errors.New("Project ending not found")
}
