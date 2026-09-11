package bindings

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestCStringGoString(t *testing.T) {
	testStr := "Hello, World!"
	cStr := CString(testStr)
	goStr := GoString(cStr)

	if goStr != testStr {
		t.Errorf("Expected %s, got %s", testStr, goStr)
	}

	// Test empty string
	emptyStr := ""
	cEmpty := CString(emptyStr)
	goEmpty := GoString(cEmpty)
	if goEmpty != emptyStr {
		t.Errorf("Expected empty string, got '%s'", goEmpty)
	}

	// Verify null termination
	// access the byte after the string
	p := uintptr(unsafe.Pointer(cEmpty))
	if *(*byte)(unsafe.Pointer(p)) != 0 {
		t.Errorf("Expected null terminator")
	}
}

func TestMockImplementation(t *testing.T) {
	// Ensure we rely on mock implementation when lib is not present
	// This assumes the test environment doesn't have the shared library in default paths
	
	sysInfo := GetSystemInfo()
	if sysInfo == "" {
		t.Error("GetSystemInfo returned empty string")
	}
	fmt.Printf("System Info: %s\n", sysInfo)

	cores := GetNumPhysicalCores()
	if cores <= 0 {
		t.Error("GetNumPhysicalCores returned invalid number")
	}

	version := GetVersion()
	if version == "" {
		t.Error("GetVersion returned empty string")
	}
	fmt.Printf("Version: %s\n", version)
}

func TestCallbackWrapper(t *testing.T) {
	// This test just ensures the wrapper function doesn't panic
	// Actual callback execution would require the mock to call it back, 
	// which currently the mock implementations (empty bodies) don't do.
	// However, we can verify the setting logic.
	
	logCb := func(level SdLogLevel, text *byte, data unsafe.Pointer) {
		fmt.Printf("Log: %s\n", GoString(text))
	}
	SetLogCallback(logCb, nil)
	
	if currentLogCallback == 0 {
		t.Error("currentLogCallback was not set")
	}
}
