package portadapter

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func stringToUINT16(sval string) (uint16, error) {
	v, err := strconv.ParseUint(sval, 0, 16)
	if err != nil {
		return 0, err
	}
	return uint16(v), nil
}

func removeRowFromFile(fileFullPath, rowToRemove string) error {
	foundLines, err := detectLines(fileFullPath, rowToRemove)
	if err != nil {
		return fmt.Errorf("can't detect lines %v in file %v, got error %v", rowToRemove, fileFullPath, err)
	}
	if len(foundLines) >= 2 {
		return fmt.Errorf("expect find 1 line (like %v) in file %v, but %v lines is found",
			rowToRemove,
			fileFullPath,
			len(foundLines))
	} else if len(foundLines) == 0 {
		return fmt.Errorf("expect find 1 line (like %v) in file %v, but no lines where found", rowToRemove, fileFullPath)
	}
	err = removeLineFromFile(fileFullPath, foundLines[0])
	if err != nil {
		return fmt.Errorf("can't remove line %v (nubmer %v) in file %v, got error %v", rowToRemove, foundLines[0], fileFullPath, err)
	}
	return nil
}

func detectLines(fullFilePath, searchedLine string) ([]int, error) {
	foundLines := []int{}
	f, err := os.Open(fullFilePath)
	if err != nil {
		return foundLines, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	lineNumber := 1
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), searchedLine) {
			foundLines = append(foundLines, lineNumber)
		}
		lineNumber++
	}
	if err := scanner.Err(); err != nil {
		return foundLines, err
	}
	return foundLines, nil
}

func detectLinesForRemove(fullFilePath, startSearch, endSearch string) (int, int, error) {
	var startLine, endLine int
	file, err := os.Open(fullFilePath)
	if err != nil {
		return startLine, endLine, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	line := 1
	isStartFinded := true
	for scanner.Scan() {
		var rowContains string
		if isStartFinded {
			rowContains = startSearch
		} else {
			rowContains = endSearch
		}
		if strings.Contains(scanner.Text(), rowContains) {
			if isStartFinded {
				startLine = line
				isStartFinded = false
				continue
			}
			endLine = line + 1
			return startLine, endLine, nil
		}
		line++
	}
	if err := scanner.Err(); err != nil {
		return startLine, endLine, err
	}
	return startLine, endLine, fmt.Errorf("can't find lines for remove in file %v", fullFilePath)
}

func removeLineFromFile(fullFilePath string, lineNubmer int) (err error) {
	file, err := os.OpenFile(fullFilePath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	cut, ok := skip(fileBytes, lineNubmer-1)
	if !ok {
		return nil // fmt.Errorf("less than %d lines", lineNubmer)
	}

	tail, ok := skip(cut, 1)
	if !ok {
		return nil // fmt.Errorf("less than %d lines after line %d", 1, lineNubmer)
	}
	t := int64(len(fileBytes) - len(cut))
	if err = file.Truncate(t); err != nil {
		return
	}
	if len(tail) > 0 {
		_, err = file.WriteAt(tail, t)
	}
	return
}

func removeLinesFromFile(fullFilePath string, startLine, numberOfLinesForRemove int) (err error) {
	if startLine < 1 {
		return errors.New("invalid request. line numbers start at 1")
	}
	if numberOfLinesForRemove < 0 {
		return errors.New("invalid request. negative number to remove")
	}
	var file *os.File
	if file, err = os.OpenFile(fullFilePath, os.O_RDWR, 0); err != nil {
		return
	}
	defer func() {
		if cErr := file.Close(); err == nil {
			err = cErr
		}
	}()
	var b []byte
	if b, err = ioutil.ReadAll(file); err != nil {
		return
	}
	cut, ok := skip(b, startLine-1)
	if !ok {
		return fmt.Errorf("less than %d lines", startLine)
	}
	if numberOfLinesForRemove == 0 {
		return nil
	}
	tail, ok := skip(cut, numberOfLinesForRemove)
	if !ok {
		return fmt.Errorf("less than %d lines after line %d", numberOfLinesForRemove, startLine)
	}
	t := int64(len(b) - len(cut))
	if err = file.Truncate(t); err != nil {
		return
	}
	if len(tail) > 0 {
		_, err = file.WriteAt(tail, t)
	}
	return
}

func skip(b []byte, n int) ([]byte, bool) {
	for ; n > 0; n-- {
		if len(b) == 0 {
			return nil, false
		}
		x := bytes.IndexByte(b, '\n')
		if x < 0 {
			x = len(b)
		} else {
			x++
		}
		b = b[x:]
	}
	return b, true
}

func combineErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}

	var errorsStringSlice []string
	for _, err := range errors {
		errorsStringSlice = append(errorsStringSlice, err.Error())
	}
	return fmt.Errorf(strings.Join(errorsStringSlice, "\n"))
}

func filesContains(sliceForSearch, allFiles []string) ([]string, error) {
	var errors []error
	var totalFinded int
	filesFinded := []string{}
	for _, routeFile := range allFiles {
		data, err := ioutil.ReadFile(routeFile)
		if err != nil {
			errors = append(errors, fmt.Errorf("Read file error: %v", err))
			continue
		}
		if findInFile(string(data), sliceForSearch) {
			filesFinded = append(filesFinded, routeFile)
			totalFinded++
		}
	}
	if totalFinded != len(sliceForSearch) {
		errors = append(errors, fmt.Errorf("find in files %v coincidences, expect %v", totalFinded, len(sliceForSearch)))
		return nil, combineErrors(errors)
	}

	return filesFinded, combineErrors(errors)
}

func findInFile(fileData string, searchSlice []string) bool {
	re := regexp.MustCompile(`(.*)/32`)
	finded := re.FindAllStringSubmatch(fileData, -1)
	if len(finded) >= 1 {
		lastWithGroup := finded[len(finded)-1]
		for _, searchedElement := range searchSlice {
			if searchedElement == lastWithGroup[1] {
				return true
			}
		}
	}
	return false
}

func trimSuffix(filePath, suffix string) (string, error) {
	dataBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("can't read file: %v", err)
	}
	return strings.TrimSuffix(string(dataBytes), suffix), nil
}