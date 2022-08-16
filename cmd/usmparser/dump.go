package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	parser "USMparser"
)

// DumpFile tries to read file `path` and write result to file `outPath`
func DumpFile(path string, outPath string) {
	src, err := os.Open(path)
	if err != nil {
		log.Fatalln("can't open source file: ", err)
	}

	out, err := os.Create(outPath)
	if err != nil {
		log.Fatalln("can't create output file: ", err)
	}

	defer func() {
		_ = src.Close()
		_ = out.Close()
	}()

	err = parser.DumpAllChunks(src, out)
	if err != nil {
		if err != io.EOF {
			log.Fatalln(err)
		}
	}
}

// DumpSubs will try to extract all the subtitles from provided file
// and save them as {{filename}}_{{language}}.srt in outputFolder
func DumpSubs(inputFile, outputFolder string, format string) {
	src, err := os.Open(inputFile)
	if err != nil {
		log.Fatalln("can't open source file: ", err)
	}

	subs, err := parser.GetSubs(src)
	src.Close()
	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}

	var result = make(map[string]bytes.Buffer)
	if format == "srt" {
		result = parser.SubsToSrt(subs)
	} else if format == "txt" {
		result = parser.SubsToTxt(subs)
	} else {
		log.Fatalln("wrong subtitle format: ", format)
	}

	// make sure output path exists
	if err := os.MkdirAll(outputFolder, os.ModePerm); err != nil {
		log.Fatalln("can't create output folder: ", err)
	}

	filename := strings.TrimSuffix(filepath.Base(inputFile), ".usm")
	for lang, sub := range result {
		newPath := fmt.Sprintf("%s_%s.%s", filepath.Join(outputFolder, filename), lang, format)
		f, err := os.Create(newPath)
		if err != nil {
			log.Printf("cant open %s: %s\n", newPath, err)
			continue
		}

		_, err = f.Write(sub.Bytes())
		//_, err = f.Write([]byte{0x0D, 0x0A}) // add trailing new line
		if err != nil {
			log.Printf("can't write to %s: %s\n", newPath, err)
		}

		log.Println(newPath, " ok!")
	}
}
