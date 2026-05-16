//go:build darwin

package tray

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa

#include <stdlib.h>

extern void cmCreateStatusItem(void);
extern void cmUpdateMenu(const char **titles, int count);
*/
import "C"
import (
	"sync"
	"unsafe"
)

var (
	once        sync.Once
	cbShowWindow func()
	cbItemClick  func(int)
	cbClear      func()
	cbQuit       func()
)

//export goTrayShowWindow
func goTrayShowWindow() {
	if cbShowWindow != nil {
		cbShowWindow()
	}
}

//export goTrayItemClicked
func goTrayItemClicked(index C.int) {
	if cbItemClick != nil {
		cbItemClick(int(index))
	}
}

//export goTrayClearHistory
func goTrayClearHistory() {
	if cbClear != nil {
		cbClear()
	}
}

//export goTrayQuit
func goTrayQuit() {
	if cbQuit != nil {
		cbQuit()
	}
}

func setupTrayImpl(showWindow func(), itemClick func(int), clear func(), quit func()) {
	cbShowWindow = showWindow
	cbItemClick = itemClick
	cbClear = clear
	cbQuit = quit
	once.Do(func() {
		C.cmCreateStatusItem()
	})
}

func updateMenuImpl(items []string) {
	if len(items) == 0 {
		C.cmUpdateMenu(nil, 0)
		return
	}

	// Build a C array of char pointers.
	cItems := make([]*C.char, len(items))
	for i, s := range items {
		cItems[i] = C.CString(s)
	}

	C.cmUpdateMenu((**C.char)(unsafe.Pointer(&cItems[0])), C.int(len(items)))

	// Free after cmUpdateMenu has copied the strings.
	for _, p := range cItems {
		C.free(unsafe.Pointer(p))
	}
}
