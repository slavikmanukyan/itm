package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"io/ioutil"

	"github.com/pkg/sftp"
	fsftp "github.com/slavikmanukyan/itm/fs/sftp"
	"github.com/slavikmanukyan/itm/itmconfig"
)

// ReadLines reads a whole file into memory
// and returns a slice of its lines.
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// WriteLines writes the lines to the given file.
func WriteLines(lines []string, path string) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func ReadLinesRemote(path string, client *sftp.Client) ([]string, error) {
	file, err := client.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func WriteLinesRemote(lines []string, path string, client *sftp.Client) error {
	file, err := client.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func CreateRemoteDir(dir string, config itmconfig.ITMConfig) {
	if config.USE_SSH {
		fsftp.Client.Mkdir(dir)
	} else {
		os.Mkdir(dir, os.ModePerm)
	}
}

func WriteRemoteFile(file string, config itmconfig.ITMConfig, data []byte) {
	if config.USE_SSH {
		out, _ := fsftp.Client.Create(file)
		out.Write(data)
	} else {
		ioutil.WriteFile(file, data, 0644)
	}
}

func ClearLine(n int) {
	fmt.Print("\r", strings.Repeat(" ", n))
}
