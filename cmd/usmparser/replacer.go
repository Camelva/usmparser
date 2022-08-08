package main

import (
	parser "USMparser"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
		_replaceAudio(f, f2, out)
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

	for _, entry := range f1Entries {
		if entry.IsDir() {
			// don't enter sub-folders
			continue
		}

		output := filepath.Join(out, entry.Name())
		_, err = os.Stat(output)
		if err == nil {
			log.Printf("%s already exists, skipping..\n", output)
			continue
		}

		entry1 := filepath.Join(in1, entry.Name())
		f, err = os.Open(entry1)
		if err != nil {
			log.Printf("can't open file %s: %s, skipping..\n", entry1, err)
			continue
		}

		entry2 := filepath.Join(in2, entry.Name())
		f2, err = os.Open(entry2)
		if err != nil {
			log.Printf("can't open file %s: %s, skipping..\n", entry2, err)
			continue
		}

		_replaceAudio(f, f2, output)
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
		log.Fatalf("can't check file info %s: %s\n", filename, err)
	}

	return f, stat1.IsDir()
}

func _replaceAudio(f, f2 *os.File, out string) {
	origInfo, err := parser.ParseFile(f)
	if err != nil {
		log.Fatalln("can't parse file: ", err)
	}
	f.Close()

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

	log.Println(out, " - ok!")
}
