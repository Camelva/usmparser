package main

import (
	parser "USMparser"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ReplaceAudio(in1, in2, out string) {
	var folderMode bool

	f, isDir1 := openFile(in1)

	if isDir1 {
		folderMode = true
		f.Close()
	}

	f2, isDir2 := openFile(in2)
	if isDir2 != isDir1 {
		log.Fatalf("both inputs should be either folders or files")
	}

	if !folderMode {
		_replaceAudio(f, f2, out, log.Default())
		return
	}
	f2.Close()

	// entering batch mode, assuming all in1, in2 and out are folder names
	f1Entries, err := os.ReadDir(in1)
	if err != nil {
		log.Fatalf("can't read folder %s: %s\n", in1, err)
	}

	if out == "" {
		out = filepath.Join(in1, "out")
	}

	if err = os.MkdirAll(out, 0755); err != nil {
		log.Fatalf("can't create output folder %s: %s", out, err)
	}

	logFileName := strings.ReplaceAll(fmt.Sprintf("log-%s.txt", time.Now().Format(time.Stamp)), ":", "_")
	fileLog, err := os.Create(logFileName)
	if err != nil {
		log.Fatalf("can't create log file %s: %s", logFileName, err)
	}

	log.Print("writing logs to ", logFileName)
	newLog := log.New(fileLog, "", log.Ltime)
	newLog.Printf("usmparser replaceaudio %s %s %s\n", in1, in2, out)

	for _, entry := range f1Entries {
		if entry.IsDir() {
			// don't enter sub-folders
			continue
		}

		name := entry.Name()

		if ext := filepath.Ext(name); ext != ".usm" {
			newLog.Printf("%s not usm file, skipping..\n", name)
			continue
		}

		output := filepath.Join(out, name)
		_, err = os.Stat(output)
		if err == nil {
			newLog.Printf("%s already exists, skipping..\n", output)
			continue
		}

		entry1 := filepath.Join(in1, name)
		f, err = os.Open(entry1)
		if err != nil {
			newLog.Printf("can't open file: %s, skipping..\n", err)
			continue
		}

		entry2 := filepath.Join(in2, name)
		f2, err = os.Open(entry2)
		if err != nil {
			newLog.Printf("can't open file: %s, skipping..\n", err)
			continue
		}

		newLog.Print(name, ": ")
		_replaceAudio(f, f2, output, newLog)
	}

	fmt.Println("All done!")
}

func openFile(filename string) (f *os.File, isDir bool) {
	var err error
	f, err = os.Open(filename)
	if err != nil {
		log.Fatalf("can't open file: %s\n", err)
	}

	stat1, err := f.Stat()
	if err != nil {
		log.Fatalf("can't check file info: %s\n", err)
	}

	return f, stat1.IsDir()
}

func _replaceAudio(f, f2 *os.File, out string, logger *log.Logger) {
	origInfo, err := parser.ParseFile(f)
	if err != nil {
		logger.Fatalln("can't parse file: ", err)
	}
	f.Close()

	file2Info, err := parser.ParseFile(f2)
	if err != nil {
		logger.Fatalln("can't parse file: ", err)
	}
	f.Close()

	outF, err := os.Create(out)
	if err != nil {
		logger.Fatalf("can't create output file: %s\n", err)
	}

	if len(file2Info.AudioStreams) <= 0 {
		logger.Println("input2 doesn't have any audio streams, skipping...")
		return
	}

	origInfo = parser.ReplaceAudio(origInfo, file2Info)

	err = origInfo.PrepareStreams().WriteTo(outF)
	if err != nil {
		logger.Fatalf("can't write result to file: %s\n", err)
	}

	logger.Println(out, "ok!")
}
