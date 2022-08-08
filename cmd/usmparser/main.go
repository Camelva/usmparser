package main

import (
	parser "USMparser"
	"fmt"
	"log"
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
		if len(args) < 4 {
			output = filepath.Dir(args[2])
		} else {
			output = args[3]
		}
		DumpSubs(args[2], output)
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
	fmt.Println(`Usage:
	usmparse replaceaudio input1 input2 [output]
	usmparse dumpfile input [output]
	usmparse dumpsubs input [output]

List of commands:
	replaceaudio - simply replace audio in input1 with audio from input2
	dumpfile - dumps everything from provided input file to output.json
	dumpsubs - dumps all the subtitles from input file and extracts to {{filename}}_{{language}}.srt inside output folder

If output doesn't exist - it will be created
If no output parameter - result will be stored in same folder as input`)
	os.Exit(0)
}

func ReplaceAudio(in1, in2, out string) {
	f, err := os.Open(in1)
	if err != nil {
		log.Fatalf("can't open file %s: %s\n", in1, err)
	}

	origInfo, err := parser.ParseFile(f)
	if err != nil {
		log.Fatalln("can't parse file: ", err)
	}

	f.Close()

	f2, err := os.Open(in2)
	if err != nil {
		log.Fatalf("can't open file %s: %s\n", in2, err)
	}

	file2Info, err := parser.ParseFile(f2)
	if err != nil {
		log.Fatalln("can't parse file: ", err)
	}

	f.Close()

	origInfo = parser.ReplaceAudio(origInfo, file2Info)

	outF, err := os.Create(out)
	if err != nil {
		log.Fatalf("can't create output file %s: %s\n", out, err)
	}

	err = origInfo.PrepareStreams().WriteTo(outF)
	if err != nil {
		log.Fatalf("can't write result to file %s: %s\n", out, err)
	}

	log.Println("All done!")
}
