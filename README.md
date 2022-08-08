# USMParser

Small tool to work with CRIWare's USM format

## Get the latest version

Go to github actions tab, open latest workflow and download artifact

## Build manually

1. Install go 1.17+
1. Run:
   ```shell
   go build .cmd/usmparser
   ```
   
### Usage

```shell
usmparser command parameters...
```

### List of commands

- 
    ```shell
    replaceaudio input1 input2 [output]
    ```
    Copies audio from input2 to input1.
    Pass folders as parameters to process all files inside them.
    If output parameter not set - will use
    - in batch mode: {{input1}}/"out"
    - in single file mode: {{input1}}-new.usm
    
- 
    ```shell
    dumpfile input [output]
    ```
    Dumps everything from provided input file to output.
    If output parameter not set - will use {{input1}}.json
    
- 
    ```shell
    dumpsubs input format [output]
    ```
    Dumps all subtitles from input file to separated files for each language, each with "_lang" suffix.
    Format can be either:
    - srt: normal subtitle format
    - txt: plaintext for Scaleform Video Encoder
    
    If output parameter not set - will output result in same folder with input
