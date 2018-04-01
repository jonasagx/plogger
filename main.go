package main

import (
	"crypto/sha1"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	separator = "\t"
)

var (
	files = make([]File, 0)
)

type File struct {
	Hash           string
	AbsPath        string
	Name           string
	Size           int
	ProcessingTime int
}

func (f *File) Values() []string {
	return []string{f.Name, f.AbsPath, f.Hash, strconv.Itoa(f.Size), strconv.Itoa(f.ProcessingTime)}
}

func sha1sum(path string) string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func visit(path string, fInfo os.FileInfo, err error) error {
	if !fInfo.IsDir() {
		// fmt.Printf("Visited: %s\n", path)
		start := time.Now()
		hash := sha1sum(path)
		elapsed := time.Since(start)
		f := File{Hash: hash, Name: fInfo.Name(), Size: int(fInfo.Size()), AbsPath: path, ProcessingTime: int(elapsed.Nanoseconds())}
		files = append(files, f)
	}
	return nil
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func export2CSV(files []File) {
	file, err := os.Create("result.csv")
	checkError("Cannot create file", err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range files {
		err := writer.Write(value.Values())
		checkError("Cannot write to file", err)
	}
}

func countHashesDupes(files []File) map[string]int {
	hashes := make(map[string]int)

	for _, f := range files {
		if c, ok := hashes[f.Hash]; ok {
			hashes[f.Hash] = c + 1
		} else {
			hashes[f.Hash] = 1
		}
	}

	return hashes
}

func findPath(files []File, hash string) string {
	for _, f := range files {
		if f.Hash == hash {
			return f.AbsPath
		}
	}

	return ""
}

func findDupes(hashes map[string]int) map[string]string {
	dupes := make(map[string]string)

	for hash, c := range hashes {
		if c > 1 {
			path := findPath(files, hash)
			dupes[hash] = path
		}
	}

	return dupes
}

func main() {
	flag.Parse()
	root := flag.Arg(0)
	err := filepath.Walk(root, visit)
	fmt.Printf("filepath.Walk() returned %v\n", err)

	// export2CSV(files)
	fmt.Println(findDupes(countHashesDupes(files)))
}
