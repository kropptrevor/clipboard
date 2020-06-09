// +build windows

package clipboard

import (
	"errors"
	"reflect"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                   = windows.NewLazyDLL("user32.dll")
	procOpenClipboard        = user32.NewProc("OpenClipboard")
	procCloseClipboard       = user32.NewProc("CloseClipboard")
	procGetClipboardData     = user32.NewProc("GetClipboardData")
	procEnumClipboardFormats = user32.NewProc("EnumClipboardFormats")
	kernel32                 = windows.NewLazyDLL("kernel32")
	procGlobalSize           = kernel32.NewProc("GlobalSize")
	procGlobalLock           = kernel32.NewProc("GlobalLock")
	procGlobalUnlock         = kernel32.NewProc("GlobalUnlock")

	// not implemented yet
	// procSetClipboardData        = user32.NewProc("SetClipboardData")
)

// Standard clipboard formats,
// see https://docs.microsoft.com/en-us/windows/win32/dataxchg/standard-clipboard-formats
const (
	FormatNone            DataFormat = 0
	FormatText            DataFormat = 1      // CF_TEXTuint
	FormatBitmap          DataFormat = 2      // CF_BITMAPuint
	FormatMetaFilePict    DataFormat = 3      // CF_METAFILEPICTuint
	FormatSYLK            DataFormat = 4      // CF_SYLKuint
	FormatDIF             DataFormat = 5      // CF_DIFuint
	FormatTIFF            DataFormat = 6      // CF_TIFFuint
	FormatOEMText         DataFormat = 7      // CF_OEMTEXTuint
	FormatDIB             DataFormat = 8      // CF_DIBuint
	FormatPalette         DataFormat = 9      // CF_PALETTEuint
	FormatPenData         DataFormat = 10     // CF_PENDATAuint
	FormatRIFF            DataFormat = 11     // CF_RIFFuint
	FormatWave            DataFormat = 12     // CF_WAVEuint
	FormatUnicodeText     DataFormat = 13     // CF_UNICODETEXTuint
	FormatENHMetaFile     DataFormat = 14     // CF_ENHMETAFILEuint
	FormatHDROP           DataFormat = 15     // CF_HDROPuint
	FormatLocale          DataFormat = 16     // CF_LOCALEuint
	FormatDIBV5           DataFormat = 17     // CF_DIBV5uint
	FormatOwnerDisplay    DataFormat = 0x0080 // CF_OWNERDISPLAYuint
	FormatDSPText         DataFormat = 0x0081 // CF_DSPTEXTuint
	FormatDSPBitmap       DataFormat = 0x0082 // CF_DSPBITMAPuint
	FormatDSPMetaFilePict DataFormat = 0x0083 // CF_DSPMETAFILEPICTuint
	FormatDSPENHMetaFile  DataFormat = 0x008E // CF_DSPENHMETAFILEuint
	FormatPrivateFirst    DataFormat = 0x0200 // CF_PRIVATEFIRSTuint
	FormatPrivateLast     DataFormat = 0x02FF // CF_PRIVATELASTuint
	FormatGDIObjFirst     DataFormat = 0x0300 // CF_GDIOBJFIRSTuint
	FormatGDIObjLast      DataFormat = 0x03FF // CF_GDIOBJLASTuint
)

func getClipboardData(hwnd unsafe.Pointer) ([]byte, DataFormat, error) {
	openClipboardSuccess, _, err := procOpenClipboard.Call(uintptr(hwnd))
	if isValidError(err) || openClipboardSuccess != 1 {
		return nil, FormatNone, err
	}
	defer procCloseClipboard.Call()
	enumFormat, _, _ := procEnumClipboardFormats.Call(uintptr(0))

	format := DataFormat(enumFormat)

	h, _, err := procGetClipboardData.Call(enumFormat)
	if isValidError(err) {
		return nil, FormatNone, err
	}

	sizep, _, err := procGlobalSize.Call(h)
	if isValidError(err) {
		return nil, FormatNone, err
	}
	usize := uint(sizep)
	const MaxUint = ^uint(0)
	if usize > MaxUint>>1 {
		return nil, FormatNone, errors.New("Size of clipboard too large")
	}
	size := int(usize)

	l, _, err := procGlobalLock.Call(h)
	if l == 0 || isValidError(err) {
		return nil, FormatNone, err
	}
	p := uintptr(unsafe.Pointer(l))

	var data []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	sh.Data = p
	sh.Len = size
	sh.Cap = size

	r, _, err := procGlobalUnlock.Call(h)
	if r == 0 || isValidError(err) {
		return nil, FormatNone, err
	}

	return data, format, nil
}

func isValidError(err error) bool {
	if err == nil || err.Error() == "The operation completed successfully." {
		return false
	}
	return true
}
