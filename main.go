package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
)

const version = "0.1.0"

func main() {
	inputCSV := flag.String("input-csv", "", "Input CSV file")
	outputCSV := flag.String("output-csv", "", "Output CSV file (default: stdout)")
	tags := flag.String("tags", "japanese/ddwordlist", "Tags to use in the third field")
	help := flag.Bool("help", false, "Show help message")
	versionFlag := flag.Bool("version", false, "Show version")

	flag.StringVar(inputCSV, "i", "", "Input CSV file")
	flag.StringVar(outputCSV, "o", "", "Output CSV file (default: stdout)")
	flag.BoolVar(help, "h", false, "Show help message")
	flag.BoolVar(versionFlag, "v", false, "Show version")

	flag.Parse()

	if *help {
		printHelp()
		return
	}

	if *versionFlag {
		fmt.Printf("ddwordlist version %s\n", version)
		return
	}

	if *inputCSV == "" {
		fmt.Println("Error: --input-csv is required")
		printHelp()
		os.Exit(1)
	}

	inputFile, err := os.Open(*inputCSV)
	if err != nil {
		fmt.Printf("Error opening input CSV file: %v\n", err)
		os.Exit(1)
	}
	defer inputFile.Close()

	// Prepare output writer
	var outputWriter io.Writer
	if *outputCSV != "" {
		outputFile, err := os.Create(*outputCSV)
		if err != nil {
			fmt.Printf("Error creating output CSV file: %v\n", err)
			os.Exit(1)
		}
		defer outputFile.Close()
		outputWriter = outputFile
	} else {
		outputWriter = os.Stdout
	}

	reader := csv.NewReader(inputFile)
	writer := csv.NewWriter(outputWriter)

	_, err = reader.Read()
	if err != nil {
		fmt.Printf("Error reading header: %v\n", err)
		os.Exit(1)
	}

	// Process each line
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading CSV: %v\n", err)
			os.Exit(1)
		}

		// Process the record
		if len(record) < 6 {
			fmt.Printf("Skipping incomplete record: %v\n", record)
			continue
		}

		firstField := record[0]
		secondField := record[1]
		thirdField := record[2]
		// We can ignore the rest of the fields

		var newFirstField string
		if secondField != "" {
			newFirstField = fmt.Sprintf("%s(%s)", secondField, firstField)
		} else {
			newFirstField = firstField
		}
		newSecondField := thirdField
		newThirdField := *tags

		newRecord := []string{newFirstField, newSecondField, newThirdField}
		err = writer.Write(newRecord)
		if err != nil {
			fmt.Printf("Error writing CSV: %v\n", err)
			os.Exit(1)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		fmt.Printf("Error flushing CSV writer: %v\n", err)
		os.Exit(1)
	}
}
func printHelp() {
	helpMessage := "\033[1;34mUsage:\033[0m\n" +
		"  ddwordlist --input-csv <input-csv> [--output-csv <output-csv>] [--tags <tags>]\n" +
		"  ddwordlist -i <input-csv> [-o <output-csv>] [--tags <tags>]\n\n" +
		"\033[1;34mOptions:\033[0m\n" +
		"  --input-csv, -i    Input CSV file (required)\n" +
		"  --output-csv, -o   Output CSV file (default: stdout)\n" +
		"  --tags             Tags to use in the third field (default: japanese/daumdict)\n" +
		"  --help, -h         Show help message\n" +
		"  --version, -v      Show version\n"
	fmt.Println(helpMessage)
}
