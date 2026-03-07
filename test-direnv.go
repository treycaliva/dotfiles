package main

import (
	"fmt"
	"regexp"
)

func main() {
	re := regexp.MustCompile(`^export ([A-Za-z0-9_]+)=\{\{ (op://.+) \}\}$`)
	matches := re.FindStringSubmatch("export GITHUB_TOKEN={{ op://Personal/GitHub/token }}")
	fmt.Printf("%q\n", matches)
}
