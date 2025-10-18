# Gofuzzler

Gofuzzler is the incredibly quick Golang implementation of Fuzzler (written in Python)

This tool:
1. crawls a webpage to return a list of words from it
2. finds the synonyms for each word to extend the wordlist
3. fuzzes the wordlist by changing case, prepending/appending digits and special characters, and more

### Usage
Download the tool:

`git clone https://github.com/suffs811/gofuzzler.git`

Build the tool:

`go build gofuzzler.go`

Run the tool:

`./gofuzzler https://example.com`

### Credits
This tool makes use of:
> CeWL commandline tool:

https://github.com/digininja/CeWL

> WordNet: 

Princeton University "About WordNet." WordNet. Princeton University. 2010
