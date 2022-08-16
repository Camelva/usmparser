package main

import (
	"fmt"
	"github.com/pterm/pterm"
	"os"
	"path/filepath"
	"strings"
)

func CoolerMain() {
	options := []string{
		"replaceaudio",
		"dumpfile",
		"dumpsubs",
	}

	command, _ := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		Show("Choose command")

	switch command {
	case "replaceaudio":
		ReplaceAudioUI()
	case "dumpfile":
		DumpFileUI()
	case "dumpsubs":
		DumpSubsUI()
	}
}

func ReplaceAudioUI() {
	input1, _ := pterm.DefaultInteractiveTextInput.
		Show("Input path to main file (or folder of files)")

	info1, err := os.Stat(input1)
	if err != nil {
		pterm.Error.Println("Can't open path: ", err)
		return
	}

	var defaultOutput string

	if info1.IsDir() {
		defaultOutput = filepath.Join(input1, "out")
	} else {
		defaultOutput = fmt.Sprintf("%s-new.usm", strings.TrimSuffix(input1, ".usm"))
	}

	input2, _ := pterm.DefaultInteractiveTextInput.
		Show("Now input path to file (or folder of files) to extract audio from")

	info2, err := os.Stat(input2)
	if err != nil {
		pterm.Fatal.Println("Can't open path: ", err)
		return
	}

	if info2.IsDir() != info1.IsDir() {
		pterm.Fatal.Println("Both inputs should be either file or folder!")
		return
	}

	output, _ := pterm.DefaultInteractiveTextInput.
		Show(fmt.Sprintf("Change output path or leave empty to keep default (%s)", defaultOutput))

	// weird workaround until they fix lib
	if output == "" || output == input2 {
		output = defaultOutput
	}

	pterm.Println()

	ReplaceAudio(input1, input1, output)
}

func DumpFileUI() {
	input, _ := pterm.DefaultInteractiveTextInput.
		Show("Input path to .usm file to dump")

	defaultOutput := fmt.Sprintf("%s-new.usm", strings.TrimSuffix(input, ".usm"))

	output, _ := pterm.DefaultInteractiveTextInput.
		Show(fmt.Sprintf("Change output path or leave empty to keep default (%s)", defaultOutput))

	// weird workaround until they fix lib
	if output == "" || output == input {
		output = defaultOutput
	}

	pterm.Println()

	DumpFile(input, output)
}

func DumpSubsUI() {
	input, _ := pterm.DefaultInteractiveTextInput.
		Show("Input path to .usm file to extract subtitles")

	options := []string{
		"srt: normal subtitle format",
		"txt: plaintext for Scaleform Video Encoder",
	}

	var format string

	formatInput, _ := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		Show("Now choose format for extracted subtitles")

	if strings.HasPrefix(formatInput, "srt") {
		format = "srt"
	} else if strings.HasPrefix(formatInput, "txt") {
		format = "txt"
	} else {
		// impossible but still
		pterm.Error.Println("Wrong format!")
		return
	}

	defaultOutput := filepath.Dir(input)

	output, _ := pterm.DefaultInteractiveTextInput.
		Show(fmt.Sprintf("Change output folder or leave empty to keep default (%s)", defaultOutput))

	// weird workaround until they fix lib
	if output == "" || output == input {
		output = defaultOutput
	}

	pterm.Println()

	DumpSubs(input, output, format)
}

func main() {
	args := os.Args

	// 1st arg is program name
	if len(args) <= 1 {
		CoolerMain()
		return
	}

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
