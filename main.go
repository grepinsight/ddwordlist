// main.go
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/extrame/xls"
)

const version = "0.2.0"

func main() {
	// Define command-line flags
	inputFile := flag.String("input-file", "", "Input CSV or XLS file")
	outputCSV := flag.String("output-csv", "", "Output CSV file (default: stdout)")
	tags := flag.String("tags", "japanese/ddwordlist", "Tags to use in the third field")
	help := flag.Bool("help", false, "Show help message")
	versionFlag := flag.Bool("version", false, "Show version")
	useLatestDownload := flag.Bool("use-latest-file-from-download-dir", false, "Use the latest file from the Downloads directory matching the pattern")
	downloadFilePattern := flag.String("download-file-pattern", "wordbook\\.xls", "Regex pattern to match files in the Downloads directory")

	// Short options
	flag.StringVar(inputFile, "i", "", "Input CSV or XLS file")
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

	if *useLatestDownload {
		latestFile, err := findLatestFileInDownloads(*downloadFilePattern)
		if err != nil {
			fmt.Printf("Error finding latest file in Downloads: %v\n", err)
			os.Exit(1)
		}
		if latestFile == "" {
			fmt.Println("No matching files found in Downloads directory.")
			os.Exit(1)
		}
		*inputFile = latestFile
	}

	if *inputFile == "" {
		fmt.Println("Error: --input-file is required")
		printHelp()
		os.Exit(1)
	}

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

	// Determine file type based on extension
	ext := strings.ToLower(filepath.Ext(*inputFile))
	if ext == ".csv" {
		err := processCSVFile(*inputFile, outputWriter, *tags)
		if err != nil {
			fmt.Printf("Error processing CSV file: %v\n", err)
			os.Exit(1)
		}
	} else if ext == ".xls" {
		err := processXLSFile(*inputFile, outputWriter, *tags)
		if err != nil {
			fmt.Printf("Error processing XLS file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Unsupported file extension '%s'. Please provide a .csv or .xls file.\n", ext)
		os.Exit(1)
	}
}

func processCSVFile(filename string, outputWriter io.Writer, tags string) error {
	inputFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening input CSV file: %v", err)
	}
	defer inputFile.Close()

	reader := csv.NewReader(inputFile)
	writer := csv.NewWriter(outputWriter)
	defer writer.Flush()

	// Read and discard the header
	_, err = reader.Read()
	if err != nil {
		return fmt.Errorf("error reading header: %v", err)
	}

	// Process each line
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading CSV: %v", err)
		}

		if len(record) < 6 {
			fmt.Printf("Skipping incomplete record: %v\n", record)
			continue
		}

		newRecord := transformRecord(record, tags)
		err = writer.Write(newRecord)
		if err != nil {
			return fmt.Errorf("error writing CSV: %v", err)
		}
	}

	if err := writer.Error(); err != nil {
		return fmt.Errorf("error flushing CSV writer: %v", err)
	}

	return nil
}

func processXLSFile(filename string, outputWriter io.Writer, tags string) error {
	xlsFile, err := xls.Open(filename, "utf-8")
	if err != nil {
		return fmt.Errorf("error opening XLS file: %v", err)
	}

	sheet := xlsFile.GetSheet(0)
	if sheet == nil {
		return fmt.Errorf("no sheets found in XLS file")
	}

	writer := csv.NewWriter(outputWriter)
	defer writer.Flush()

	// Skip header row
	startRow := 1

	for rowIndex := startRow; rowIndex <= int(sheet.MaxRow); rowIndex++ {
		row := sheet.Row(rowIndex)
		if row == nil {
			continue
		}

		record := make([]string, 0, 6)
		for colIndex := 0; colIndex < 6; colIndex++ {
			cell := row.Col(colIndex)
			record = append(record, cell)
		}

		if len(record) < 6 {
			fmt.Printf("Skipping incomplete record at row %d: %v\n", rowIndex+1, record)
			continue
		}

		newRecord := transformRecord(record, tags)
		err = writer.Write(newRecord)
		if err != nil {
			return fmt.Errorf("error writing CSV: %v", err)
		}
	}

	if err := writer.Error(); err != nil {
		return fmt.Errorf("error flushing CSV writer: %v", err)
	}

	return nil
}

func transformRecord(record []string, tags string) []string {
	firstField := record[0]
	secondField := record[1]
	thirdField := record[2]

	var newFirstField string
	if secondField != "" {
		newFirstField = fmt.Sprintf("%s(%s)", secondField, firstField)
	} else {
		newFirstField = firstField
	}
	newSecondField := thirdField
	newThirdField := tags

	return []string{newFirstField, newSecondField, newThirdField}
}

func findLatestFileInDownloads(pattern string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %v", err)
	}

	downloadsDir := filepath.Join(homeDir, "Downloads")

	files, err := os.ReadDir(downloadsDir)
	if err != nil {
		return "", fmt.Errorf("unable to read Downloads directory: %v", err)
	}

	var latestFile string
	var latestModTime time.Time

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !regex.MatchString(file.Name()) {
			continue
		}

		fileInfo, err := file.Info()
		if err != nil {
			continue
		}

		modTime := fileInfo.ModTime()
		if latestFile == "" || modTime.After(latestModTime) {
			latestFile = filepath.Join(downloadsDir, file.Name())
			latestModTime = modTime
		}
	}

	return latestFile, nil
}

func printHelp() {
	helpMessage := "\033[1;34mUsage:\033[0m\n" +
		"  ddwordlist --input-file <input-file> [--output-csv <output-csv>] [--tags <tags>]\n" +
		"  ddwordlist -i <input-file> [-o <output-csv>] [--tags <tags>]\n\n" +
		"\033[1;34mOptions:\033[0m\n" +
		"  --input-file, -i                   Input CSV or XLS file (required)\n" +
		"  --output-csv, -o                   Output CSV file (default: stdout)\n" +
		"  --tags                             Tags to use in the third field (default: japanese/ddwordlist)\n" +
		"  --use-latest-file-from-download-dir  Use the latest file from the Downloads directory matching the pattern\n" +
		"  --download-file-pattern            Regex pattern to match files in Downloads (default: wordbook\\.xls)\n" +
		"  --help, -h                         Show help message\n" +
		"  --version, -v                      Show version\n"
	fmt.Println(helpMessage)
}
