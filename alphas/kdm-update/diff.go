package main

import (
	"bufio"
	"os"
	"strings"
)

func agentAndServerArgsDiff(url, newTag, oldTag string) ([]string, error) {
	// client := http.NewClient(15 * time.Second)
	// response, err := client.Get(fmt.Sprintf("%s/%s...%s.diff", url, oldTag, newTag))
	// if err != nil {
	// 	return nil, err
	// }
	// defer response.Body.Close()
	b, err := os.Open("./k3s.diff")
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(b)
	// scanner := bufio.NewScanner(response.Body)
	diff := []string{}
	insideDiff := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "+++") || strings.Contains(line, "---") {
			if strings.Contains(line, "pkg/cli/cmds/agent.go") || strings.Contains(line, "pkg/cli/cmds/server.go") {
				insideDiff = true
			}
		}
		if strings.Contains(line, "diff --git") {
			insideDiff = false
		}
		if insideDiff {
			// replace tabs with spaces
			line = strings.ReplaceAll(line, "\u0009", " ")
			diff = append(diff, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return diff, nil
}
