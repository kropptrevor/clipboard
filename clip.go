package clipboard

import (
	"fmt"
	"unsafe"
)

// DataFormat is used to specify the kind of data to be get or set
type DataFormat int

// GetFromClipboard gets data from the clipboard of the given type
func GetFromClipboard(hwnd unsafe.Pointer) ([]byte, DataFormat, error) {
	data, dataType, err := getClipboardData(hwnd)
	if err != nil {
		panic(fmt.Sprintf("could not get from clipboard, %v", err))
	}
	return data, dataType, nil
}
