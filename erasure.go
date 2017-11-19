package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	humanize "github.com/dustin/go-humanize"
	"github.com/klauspost/reedsolomon"
	"github.com/olekukonko/tablewriter"
)

const outputDir = "./output/"

var totalShards = flag.Int("t", 6, "Sum of data & parity")
var fileName = flag.String("f", "testfile.txt", "Input file to perform erasure-coding")
var showReadQuorum = flag.Bool("r", false, "Show read quorum values for all combinations")

func main() {
	// create output directory if not present
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, os.ModePerm)
	}

	// Parse command line parameters.
	flag.Parse()
	args := flag.Args()

	if len(args) != 0 {
		fmt.Printf("Arguments not allowed. Use below flags:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	validateInput(totalShards)

	// initialize with half data and parity shards
	dataShards := *totalShards / 2
	parityShards := *totalShards / 2

	// get input file size
	inputFileSize := getInputFileSize(*fileName)
	fmt.Printf("Input file size: %d bytes, Total shards: %d\n\n", inputFileSize, *totalShards)

	// initialize table writer
	table := tablewriter.NewWriter(os.Stdout)
	if *showReadQuorum {
		table.SetHeader([]string{"Data Shards", "Parity Shards", "Storage Usage Ratio", "Read Quorum"})
	} else {
		table.SetHeader([]string{"Data Shards", "Parity Shards", "Storage Usage Ratio"})
	}

	// get the possible data and parity shards for given disk count.
	// Following are the constraints we follow here:
	// 1. totalShards is always an even number, greater than 3, less than 257.
	// 2. parityShards are at least 2.
	// 3. dataShards are always greater than or equal to parityShards.
	for parityShards >= 2 {

		// Create encoding matrix.
		enc, err := reedsolomon.New(dataShards, parityShards)
		checkErr(err)

		// read the given file
		b, err := ioutil.ReadFile(*fileName)
		checkErr(err)

		// Split the file into equally sized shards.
		shards, err := enc.Split(b)
		checkErr(err)

		// Encode parity
		err = enc.Encode(shards)
		checkErr(err)

		// Write out the resulting files.
		for i, shard := range shards {
			outFile := fmt.Sprintf("%s.%d", fileName, i)
			err = ioutil.WriteFile(filepath.Join(outputDir, outFile), shard, os.ModePerm)
			checkErr(err)
		}

		// calculate total disk space needed for this configuration
		//fmt.Printf("Total disk space usage after erasure coding: %d bytes\n", len(shards)*len(shards[0]))

		// space utilization factor
		ratio := humanize.FormatFloat("#.##", float64(len(shards)*len(shards[0]))/float64(inputFileSize))
		tablerow := []string{strconv.Itoa(dataShards), strconv.Itoa(parityShards), ratio}

		// start reducing the parityshards to check the readQuorum
		for i := parityShards; i >= 2; i-- {
			// set shards to nil
			for j := 1; j <= i+1; j++ {
				shards[*totalShards-j] = nil
			}

			err := enc.Reconstruct(shards)
			// err means reconstruction is not possible and we have less shards than readquorum
			if err != nil && *showReadQuorum {
				tablerow = append(tablerow, strconv.Itoa(*totalShards-i))
				break
			}
		}

		dataShards++
		parityShards--

		// clean up all the text files in output/
		cleanUpShards()
		table.Append(tablerow)
	}
	table.Render()
}

func getInputFileSize(fileName string) int64 {
	file, err := os.Open(fileName)
	checkErr(err)
	fileInfo, err := file.Stat()
	checkErr(err)
	return fileInfo.Size()
}

func cleanUpShards() {
	files, err := ioutil.ReadDir(outputDir)
	checkErr(err)
	for _, f := range files {
		os.Remove(outputDir + f.Name())
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(2)
	}
}

func validateInput(totalShards *int) {
	if *totalShards > 257 {
		fmt.Printf("Error: Too many shards\n")
		os.Exit(1)
	}

	if *totalShards < 4 {
		fmt.Printf("Error: Too few shards\n")
		os.Exit(1)
	}

	if *totalShards%2 != 0 {
		fmt.Printf("Error: Total shards should be even\n")
		os.Exit(1)
	}
}
