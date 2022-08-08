package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args

	if len(args) < 3 {
		displayHelp()
	}

	var output string

	// different commands
	switch strings.ToLower(args[1]) {
	case "dumpfile":
		if len(args) < 4 {
			output = strings.TrimSuffix(args[2], ".usm") + ".json"
		} else {
			output = args[3]
		}
		DumpFile(args[2], output)
	case "dumpsubs":
		if len(args) >= 5 {
			output = args[4]
		} else if len(args) == 4 {
			output = filepath.Dir(args[2])
		} else {
			fmt.Println("need to specify output format - srt or txt")
			os.Exit(1)
		}

		DumpSubs(args[2], output, args[3])
	case "replaceaudio":
		// no output provided, use same folder
		if len(args) < 5 {
			output = strings.TrimSuffix(args[2], ".usm") + "-new.usm"
		} else {
			output = args[4]
		}
		ReplaceAudio(args[2], args[3], output)
	default:
		displayHelp()
	}
}

func displayHelp() {
	fmt.Print(Help)
	os.Exit(0)
}

var Help = `Usage:
	usmparser command parameters...

List of available commands:
	- replaceaudio input1 input2 [output]
		Copies audio from input2 to input1.
		Pass folders as parameters to process all files inside them.
		If output parameter not set - will use 
			- in batch mode: {{input1}}/"out"
			- in single file mode: {{input1}}-new.usm

	- dumpfile input [output]
		Dumps everything from provided input file to output.
		If output parameter not set - will use {{input1}}.json

	- dumpsubs input format [output]
		Dumps all subtitles from input file to separated files for each language, each with "_lang" suffix.
		Format can be either:
			- srt: normal subtitle format
			- txt: plaintext for Scaleform Video Encoder
		If output parameter not set - will output result in same folder with input
`
