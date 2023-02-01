package pdh

import (
	"fmt"
	"github.com/cbwest3-ntnx/win"
	"strings"
)

var (
	nullPtr *uint16
)

// TODO (cbwest): Do proper error handling.
func LocalizeAndExpandCounter(pdhQuery win.PDH_HQUERY, path string) (paths []string, instances []string, err error) {
	var counterHandle win.PDH_HCOUNTER
	var ret = win.PdhAddEnglishCounter(pdhQuery, path, 0, &counterHandle)
	if ret != win.PDH_CSTATUS_VALID_DATA { // Error checking
		fmt.Printf("ERROR: PdhAddEnglishCounter return code is %s (0x%X)\n",
			win.PDHErrors[ret], ret)
	}

	// Call PdhGetCounterInfo twice to get buffer size, per
	// https://learn.microsoft.com/en-us/windows/win32/api/pdh/nf-pdh-pdhgetcounterinfoa#remarks.
	var bufSize uint32 = 0
	var retrieveExplainText uint32 = 0
	ret = win.PdhGetCounterInfo(counterHandle, uintptr(retrieveExplainText), &bufSize, nil)
	if ret != win.PDH_MORE_DATA { // error checking
		fmt.Printf("ERROR: First PdhGetCounterInfo return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}

	var counterInfo win.PDH_COUNTER_INFO
	ret = win.PdhGetCounterInfo(counterHandle, uintptr(retrieveExplainText), &bufSize, &counterInfo)
	if ret != win.PDH_CSTATUS_VALID_DATA { // error checking
		fmt.Printf("ERROR: Second PdhGetCounterInfo return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}

	// Call PdhExpandWildCardPath twice, per
	// https://learn.microsoft.com/en-us/windows/win32/api/pdh/nf-pdh-pdhexpandwildcardpathha#remarks.
	var flags uint32 = 0
	var pathListLength uint32 = 0
	ret = win.PdhExpandWildCardPath(nullPtr, counterInfo.SzFullPath, nullPtr, &pathListLength, &flags)
	if ret != win.PDH_MORE_DATA { // error checking
		fmt.Printf("ERROR: First PdhExpandWildCardPath return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}
	if pathListLength < 1 {
		fmt.Printf("ERROR: SOMETHING IS WRONG. pathListLength < 1, is %d.\n", pathListLength)
	}

	expandedPathList := make([]uint16, pathListLength)
	ret = win.PdhExpandWildCardPath(nullPtr, counterInfo.SzFullPath, &expandedPathList[0], &pathListLength, &flags)
	if ret != win.PDH_CSTATUS_VALID_DATA { // error checking
		fmt.Printf("ERROR: Second PdhExpandWildCardPath return code is %s (0x%X)\n", win.PDHErrors[ret], ret)
	}

	for i := 0; i < int(pathListLength); i += len(path) + 1 {
		expandedPath := win.UTF16PtrToString(&expandedPathList[i])
		if len(path) < 1 { // expandedPathList has two nulls at the end.
			continue
		}

		// Parse PDH instance from the expanded counter path.
		instanceStartIndex := strings.Index(expandedPath, "(")
		instanceEndIndex := strings.Index(expandedPath, ")")
		if instanceStartIndex < 0 || instanceEndIndex < 0 {
			fmt.Printf("Unable to parse PDH counter instance from '%s'", path)
			continue
		}
		instance := expandedPath[instanceStartIndex+1 : instanceEndIndex]

		if instance == "_Total" { // Skip the _Total instance. That is for users to compute.
			continue
		}
		paths = append(paths, expandedPath)
		instances = append(instances, instance)
		fmt.Printf("Expanded %s to %s\n", path, paths)
	}
	return paths, instances, nil
}
