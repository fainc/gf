package gfile_test

import (
	"fmt"
	"regexp"

	"github.com/gogf/gf/v2/os/gfile"
)

func ExampleReplaceFile() {
	// init
	var (
		fileName = "gflie_example.txt"
		tempDir  = gfile.TempDir("gfile_example_replace")
		tempFile = gfile.Join(tempDir, fileName)
	)

	// write contents
	gfile.PutContents(tempFile, "goframe example content")

	// read contents
	fmt.Println(gfile.GetContents(tempFile))

	gfile.ReplaceFile("content", "replace word", tempFile)

	fmt.Println(gfile.GetContents(tempFile))

	// Output:
	// goframe example content
	// goframe example replace word
}

func ExampleReplaceFileFunc() {
	// init
	var (
		fileName = "gflie_example.txt"
		tempDir  = gfile.TempDir("gfile_example_replace")
		tempFile = gfile.Join(tempDir, fileName)
	)

	// write contents
	gfile.PutContents(tempFile, "goframe example 123")

	// read contents
	fmt.Println(gfile.GetContents(tempFile))

	// replace by yourself
	gfile.ReplaceFileFunc(func(path, content string) string {
		// Replace with regular match
		reg, _ := regexp.Compile(`\d{3}`)
		return reg.ReplaceAllString(content, "[num]")
	}, tempFile)

	fmt.Println(gfile.GetContents(tempFile))

	// Output:
	// goframe example 123
	// goframe example [num]
}

func ExampleReplaceDir() {
	// init
	var (
		fileName = "gflie_example.txt"
		tempDir  = gfile.TempDir("gfile_example_replace")
		tempFile = gfile.Join(tempDir, fileName)
	)

	// write contents
	gfile.PutContents(tempFile, "goframe example content")

	// read contents
	fmt.Println(gfile.GetContents(tempFile))

	gfile.ReplaceDir("content", "replace word", tempDir, "gflie_example.txt", true)

	// read contents
	fmt.Println(gfile.GetContents(tempFile))

	// Output:
	// goframe example content
	// goframe example replace word
}

func ExampleReplaceDirFunc() {
	// init
	var (
		fileName = "gflie_example.txt"
		tempDir  = gfile.TempDir("gfile_example_replace")
		tempFile = gfile.Join(tempDir, fileName)
	)

	// write contents
	gfile.PutContents(tempFile, "goframe example 123")

	// read contents
	fmt.Println(gfile.GetContents(tempFile))

	// replace by yourself
	gfile.ReplaceDirFunc(func(path, content string) string {
		// Replace with regular match
		reg, _ := regexp.Compile(`\d{3}`)
		return reg.ReplaceAllString(content, "[num]")
	}, tempDir, "gflie_example.txt", true)

	fmt.Println(gfile.GetContents(tempFile))

	// Output:
	// goframe example 123
	// goframe example [num]

}
