package flociaws_test

import (
	"bufio"
	"os"
	"strings"
)

// testImage is the pinned floci/floci tag from .github/docker-images.txt, the single place
// the integration tests and CI take it from.
var testImage = pinnedImage("floci/floci:")

func pinnedImage(prefix string) string {
	f, err := os.Open("../.github/docker-images.txt")
	if err != nil {
		panic("reading the pinned images: " + err.Error())
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); strings.HasPrefix(line, prefix) {
			return line
		}
	}
	panic("no " + prefix + " image in .github/docker-images.txt")
}
