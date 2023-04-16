package gohashmob

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/TylerBrock/colorjson"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

const apiUrl = "https://hashmob.net/api/v2/search/paid"

type HashRequest struct {
	Hashes []string `json:"hashes"`
}

func getDotfile() (string, error) {
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".zshrc"), nil
	} else if strings.Contains(shell, "bash") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".bashrc"), nil
	} else {
		return ".bashrc", nil
	}
}

func main() {
	// boldblue := color.New(color.FgBlue, color.Bold)
	blue := color.New(color.FgBlue)
	red := color.New(color.FgRed)

	var quiet = flag.Bool("q", false, "Output founds as hash:plain instead of the full API response")
	var noformatting = flag.Bool("n", false, "Disable JSON response prettifying")
	var nocolor = flag.Bool("no-color", false, "Disable colored log output (automatically disabled when piping)")
	flag.Usage = func() {
		fmt.Println("Reads a list of hashes and looks for their cleartext counterparts in HashMob's database.")
		fmt.Println("If no positional argument is provided and the program detects a pipe, hashes are read from STDIN.")
		fmt.Println("A valid API key must be supplied via the HASHMOB_API environment variable.")
		fmt.Fprintf(os.Stdout, "\nUsage: %s [-q] [-n] <hash input> (single hash / comma separated hashes / file path)\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stdout, "\nExamples: \n")
		fmt.Fprintf(os.Stdout, "   %s -q 098f6bcd4621d373cade4e832627b4f6\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "   cat hashes.txt | %s -q\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "   %s 098f6bcd4621d373cade4e832627b4f6,5f4dcc3b5aa765d61d8327deb882cf99\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "   %s -q path/to/hashes.txt\n", os.Args[0])
	}
	flag.Parse()

	// Everything is blue
	color.Set(color.FgBlue)

	// Except when it's not
	if *nocolor {
		red.Fprintf(os.Stderr, "")
		blue.DisableColor()
		red.DisableColor()
	}

	api_key := os.Getenv("HASHMOB_KEY")

	// If API variable not set, make a scene
	if api_key == "" {
		dotfile, err := getDotfile()
		if err != nil {
			dotfile = ".bashrc"
		}
		red.Fprintln(os.Stderr, "[!] ERROR: A valid API key must be specified in the environment variable HASHMOB_KEY")
		if !*nocolor {
			color.Set(color.FgBlue)
		}
		fmt.Fprintln(os.Stderr, "\n    Example: export HASHMOB_KEY=329b9b8c-dc02-11ed-8c5d-e7484ed0ea8c")
		fmt.Fprintln(os.Stderr, "\n    To keep it between sessions, add the above command to your shell's dotfile like so:")
		fmt.Fprintf(os.Stderr, "    echo 'export HASHMOB_KEY=329b9b8c-dc02-11ed-8c5d-e7484ed0ea8c' >> %s\n", dotfile)
		color.Unset()
		os.Exit(1)
	}

	var hashes []string
	// If run without args or STDIN, mansplain
	if isatty.IsTerminal(os.Stdin.Fd()) && flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [-q] [-n] <hash input>\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Run with -h, --help for full usage details.")
		color.Unset()
		os.Exit(1)
	} else if !isatty.IsTerminal(os.Stdin.Fd()) && flag.NArg() == 0 {
		// Read input from STDIN
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			hashes = append(hashes, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "[!] Error reading input:", err)
			os.Exit(1)
		}
	} else {
		hashArg := flag.Arg(0)
		// Check if the input is a file
		if fileInfo, err := os.Stat(hashArg); err == nil && !fileInfo.IsDir() {
			if !*quiet {
				fmt.Fprintln(os.Stderr, "[+] Reading hashes from file:", hashArg)
			}
			color.Unset()
			file, err := os.Open(hashArg)
			if err != nil {
				red.Fprintln(os.Stderr, "[!] Error opening file:", err)
				os.Exit(1)
			}
			defer file.Close()

			// Read file content and split by lines
			scanner := bufio.NewScanner(file)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				line := scanner.Text()
				if line != "" {
					hashes = append(hashes, line)
				}
			}
		} else {
			// Split by commas
			hashes = strings.Split(hashArg, ",")
		}
	}

	// Remove any leading/trailing whitespace from hashes
	for i, hash := range hashes {
		hashes[i] = strings.TrimSpace(hash)
	}

	// Create request payload
	reqPayload := HashRequest{Hashes: hashes}
	payloadBytes, err := json.Marshal(reqPayload)
	if err != nil {
		red.Fprintln(os.Stderr, "[!] Error creating request payload:", err)
		os.Exit(1)
	}

	// Send request to API
	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		red.Fprintln(os.Stderr, "[!] Error creating request:", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "*/*")
	req.Header.Set("User-Agent", "gohashmob v0.1 (github.com/n0kovo/gohashmob)")
	req.Header.Set("api-key", api_key)
	req.Header.Set("X-CSRF-TOKEN", "")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		red.Fprintln(os.Stderr, "[!] Error sending request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Parse response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		red.Fprintln(os.Stderr, "[!] Error reading response:", err)
		os.Exit(1)
	}

	if *quiet {
		// Extract plain value if found
		type ApiResponse struct {
			Data struct {
				Found []struct {
					Hash  string `json:"hash"`
					Plain string `json:"plain"`
				} `json:"found"`
			} `json:"data"`
		}

		var apiResponse ApiResponse
		err = json.Unmarshal(respBody, &apiResponse)
		if err != nil {
			red.Fprintln(os.Stderr, "[!] Error parsing response:", err)
		} else {
			if len(apiResponse.Data.Found) != 0 {
				for _, found := range apiResponse.Data.Found {
					fmt.Printf("%s:%s\n", found.Hash, found.Plain)
				}
			} else {
				fmt.Fprintln(os.Stderr, "[!] No results :(")
				os.Exit(1)
			}
		}
	} else {
		if *noformatting {
			// Ugly print JSON response
			fmt.Println(string(respBody))
		} else {
			// Pretty print JSON response
			var jsonResponse interface{}
			err = json.Unmarshal(respBody, &jsonResponse)
			if err != nil {
				red.Fprintln(os.Stderr, "[!] Error parsing response:", err)
				os.Exit(1)
			}
			color.Unset()
			f := colorjson.NewFormatter()
			f.Indent = 4
			formattedJson, _ := f.Marshal(jsonResponse)

			fmt.Println(string(formattedJson))
		}
	}
}
