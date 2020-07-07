package portadapter

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/r3labs/diff"
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

// TODO: logic bellow must be not in portadapter, it's domain
func compareDomainServicesData(actualServicesInfo, storageServicesInfo []*domain.ServiceInfo) error {
	if len(actualServicesInfo) == 0 && len(storageServicesInfo) == 0 {
		return nil
	}

	// var sortedActualServicesInfo []*domain.ServiceInfo
	// if len(actualServicesInfo) == 0 {
	sortedActualServicesInfo := SortServicesInfoAndApplicationServers(actualServicesInfo)
	// }

	// var sortedStorageServicesInfo []*domain.ServiceInfo
	// if len(actualServicesInfo) == 0 {
	sortedStorageServicesInfo := SortServicesInfoAndApplicationServers(storageServicesInfo)
	// }
	changelog, err := diff.Diff(sortedActualServicesInfo, sortedStorageServicesInfo)
	if err != nil {
		return fmt.Errorf("can't get diff between current and storage config: %v", err)
	}

	if len(changelog) != 0 {
		errSlice := []error{fmt.Errorf("find %v difference(s) between configs: ", len(changelog))}

		for _, change := range changelog {
			errMsg := fmt.Errorf("\npath: %s\ntype: %s\ncurrent config: %s\nstorage config: %s",
				change.Path,
				change.Type,
				change.From,
				change.To)
			errSlice = append(errSlice, errMsg)
		}
		return combineErrors(errSlice)
	}
	return nil
}

// SortServicesInfoAndApplicationServers - sort all services include child object application servers
func SortServicesInfoAndApplicationServers(unsortedServicesInfo []*domain.ServiceInfo) []*domain.ServiceInfo {
	sortedServicesInfo := []*domain.ServiceInfo{}
	newServicesInfo := sortedOnlyServices(unsortedServicesInfo)
	for _, sortedOnlyServiceInfo := range newServicesInfo { // terrible second 'for' loop..
		for _, unsortedServiceInfo := range unsortedServicesInfo {
			if sortedOnlyServiceInfo.ServiceIP == unsortedServiceInfo.ServiceIP &&
				sortedOnlyServiceInfo.ServicePort == unsortedServiceInfo.ServicePort {
				sortedApplicationServers := sortApplicationServers(unsortedServiceInfo.ApplicationServers)
				sortedServiceInfo := &domain.ServiceInfo{
					ServiceIP:          sortedOnlyServiceInfo.ServiceIP,
					ServicePort:        sortedOnlyServiceInfo.ServicePort,
					ApplicationServers: sortedApplicationServers,
					HealthcheckType:    unsortedServiceInfo.HealthcheckType,
					ExtraInfo:          unsortedServiceInfo.ExtraInfo,
				}
				sortedServicesInfo = append(sortedServicesInfo, sortedServiceInfo)
			}
		}

	}
	return sortedServicesInfo
}

func sortedOnlyServices(unsortedServicesInfo []*domain.ServiceInfo) []*domain.ServiceInfo {
	servicesInfoSlice := formServicesInfoFromDomainModel(unsortedServicesInfo)
	onlyServicesInfoSortedSlice := sortIPs(servicesInfoSlice)
	return formServicesInfoDomainModelFromSlice(onlyServicesInfoSortedSlice)
}

func formServicesInfoFromDomainModel(servicesInfo []*domain.ServiceInfo) []string {
	servicesInfoSlice := []string{}
	for _, serviceInfo := range servicesInfo {
		servicesInfoSlice = append(servicesInfoSlice, serviceInfo.ServiceIP+":"+serviceInfo.ServicePort)
	}
	return servicesInfoSlice
}

func formServicesInfoDomainModelFromSlice(servicesInfoSlice []string) []*domain.ServiceInfo { // never check len servicesInfoSlice 2, hardcoded
	servicesInfo := []*domain.ServiceInfo{}
	for _, serviceInfo := range servicesInfoSlice {
		serviceInfoSlice := strings.Split(serviceInfo, ":")
		serviceInfo := &domain.ServiceInfo{
			ServiceIP:   serviceInfoSlice[0],
			ServicePort: serviceInfoSlice[1],
		}
		servicesInfo = append(servicesInfo, serviceInfo)
	}
	return servicesInfo
}

func sortApplicationServers(applicationServers []*domain.ApplicationServer) []*domain.ApplicationServer {
	applicationServersSlice := formApplicationServersSliceFromDomainModel(applicationServers)
	sortedApplicationServersSlice := sortIPs(applicationServersSlice)
	return formApplicationServersDomainModelFromSlice(sortedApplicationServersSlice)
}

func formApplicationServersDomainModelFromSlice(applicationServersSlice []string) []*domain.ApplicationServer { // never check len applicationServersSlice 2, hardcoded
	applicationServers := []*domain.ApplicationServer{}
	for _, applicationServer := range applicationServersSlice {
		applicationServerSlice := strings.Split(applicationServer, ":")
		//
		applicationServer := &domain.ApplicationServer{
			ServerIP:   applicationServerSlice[0],
			ServerPort: applicationServerSlice[1],
		}
		applicationServers = append(applicationServers, applicationServer)
	}
	return applicationServers
}

func formApplicationServersSliceFromDomainModel(applicationServers []*domain.ApplicationServer) []string {
	applicationServersSlice := []string{}
	for _, applicationServer := range applicationServers {
		applicationServersSlice = append(applicationServersSlice, applicationServer.ServerIP+":"+applicationServer.ServerPort)
	}
	return applicationServersSlice
}

// sortIPs sorts IP addresses of an array in asc. order (quicksort)
func sortIPs(addrs []string) []string {
	//recursion base case
	if len(addrs) < 2 {
		return addrs
	}
	pivot := addrs[rand.Intn(len(addrs))] //random ip in addrs as pivot

	var left []string   //for IPs<pivot
	var middle []string //for IPs=pivot
	var right []string  //for IPs>pivot

	for _, ip := range addrs {
		if orderIPPair(ip, pivot)[0] == ip {
			left = append(left, ip)
		} else if ip == pivot {
			middle = append(middle, ip)
		} else {
			right = append(right, ip)
		}
	}

	//combine and return
	left, right = sortIPs(left), sortIPs(right)
	sortedIPs := append(left, middle...)
	sortedIPs = append(sortedIPs, right...)
	return sortedIPs
}

// orderIPPair returns an array of two IP:PORTs in ascending order
func orderIPPair(firstIP string, secondIP string) [2]string {
	//extract numbers out of IP strings
	firstIPPortArr := strings.Split(firstIP, ":")
	firstIPArr := strings.Split(firstIPPortArr[0], ".")
	secondIPPortArr := strings.Split(secondIP, ":")
	secondIPArr := strings.Split(secondIPPortArr[0], ".")

	//compare and return ordered string literal array
	for index, num := range firstIPArr {
		if num == secondIPArr[index] {
			continue
		} else if num > secondIPArr[index] {
			return [2]string{secondIP, firstIP}
		} else {
			return [2]string{firstIP, secondIP}
		}
	} //now check ports if nums were the same and return ordered IP:Ports
	if firstIPPortArr[1] < secondIPPortArr[1] {
		return [2]string{firstIP, secondIP}
	}
	return [2]string{secondIP, firstIP}
}
