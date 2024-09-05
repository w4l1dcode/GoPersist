package persist

import (
	"fmt"
	"golang.org/x/sys/windows"
	_ "golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
	"log"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// CreateService creates a new Windows service
func CreateService(serviceName, displayName, executablePath, args string) error {
	// Open a handle to the Service Control Manager
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service control manager: %w", err)
	}
	defer m.Disconnect()

	// Check if the service already exists
	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", serviceName)
	}

	// Create the service
	s, err = m.CreateService(serviceName, executablePath, mgr.Config{
		DisplayName:    displayName,
		BinaryPathName: executablePath + " " + args,
		StartType:      mgr.StartAutomatic,
	}, []string{}...)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer s.Close()

	fmt.Printf("Service %s created successfully.\n", serviceName)
	return nil
}

func StartService(serviceName string) error {
	// Convert the service name to UTF-16
	serviceNamePtr, err := convertStringToUTF16Ptr(serviceName)
	if err != nil {
		return fmt.Errorf("failed to convert service name to UTF-16: %w", err)
	}

	// Open a handle to the Service Control Manager
	mgr, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to connect to service control manager: %w", err)
	}
	defer windows.CloseServiceHandle(mgr)

	// Open a handle to the service
	svc, err := windows.OpenService(mgr, serviceNamePtr, windows.SERVICE_START)
	if err != nil {
		return fmt.Errorf("failed to open service %s: %w", serviceName, err)
	}
	defer windows.CloseServiceHandle(svc)

	// Attempt to start the service
	err = windows.StartService(svc, 0, nil)
	if err != nil {
		return fmt.Errorf("failed to start service %s: %w.", serviceName, err)
	}

	// Wait for the service to start
	timeout := time.After(30 * time.Second)
	ticker := time.Tick(1 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout: service %s did not start in time", serviceName)
		case <-ticker:
			status, err := getServiceStatus(svc)
			if err != nil {
				return err
			}
			if status == windows.SERVICE_RUNNING {
				fmt.Printf("Service %s started successfully.\n", serviceName)
				return nil
			}
		}
	}
}

// getServiceStatus retrieves the current status of the service
func getServiceStatus(svc windows.Handle) (uint32, error) {
	var status windows.SERVICE_STATUS
	err := windows.QueryServiceStatus(svc, &status)
	if err != nil {
		return 0, err
	}
	return status.CurrentState, nil
}

// convertStringToUTF16Ptr converts a Go string to a UTF-16 pointer.
func convertStringToUTF16Ptr(s string) (*uint16, error) {
	utf16, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		return nil, err
	}
	return utf16, nil
}

// DeleteService attempts to delete a Windows service given its name.
func DeleteService(serviceName string) error {
	// Convert the service name to UTF-16
	serviceNamePtr, err := convertStringToUTF16Ptr(serviceName)
	if err != nil {
		return fmt.Errorf("failed to convert service name to UTF-16: %w", err)
	}

	// Open a handle to the Service Control Manager
	mgr, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to connect to service control manager: %w", err)
	}
	defer windows.CloseServiceHandle(mgr)

	// Open a handle to the service
	svc, err := windows.OpenService(mgr, serviceNamePtr, windows.DELETE)
	if err != nil {
		return fmt.Errorf("failed to open service %s: %w", serviceName, err)
	}
	defer windows.CloseServiceHandle(svc)

	// Attempt to delete the service
	if err := windows.DeleteService(svc); err != nil {
		return fmt.Errorf("failed to delete service %s: %w", serviceName, err)
	}

	log.Printf("Service %s deleted successfully", serviceName)
	return nil
}

// CreateServiceBatchFile creates a batch file that runs the specified command with arguments.
func CreateServiceBatchFile(command, args, filePath string) error {
	// Open or create the batch file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create batch file: %w", err)
	}
	defer file.Close()

	// Combine the command and arguments into one string
	fullCommand := fmt.Sprintf("%s %s", command, args)

	// Write the command to the batch file
	_, err = file.WriteString(fullCommand)
	if err != nil {
		return fmt.Errorf("failed to write to batch file: %w", err)
	}

	fmt.Printf("Batch file created successfully at: %s\n", filePath)
	return nil
}

const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

var (
	modadvapi32 = windows.NewLazySystemDLL("advapi32.dll")
	modkernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procGetCurrentThread = modkernel32.NewProc("GetCurrentThread")
	procOpenThreadToken  = modadvapi32.NewProc("OpenThreadToken")
	procImpersonateSelf  = modadvapi32.NewProc("ImpersonateSelf")
	procRevertToSelf     = modadvapi32.NewProc("RevertToSelf")
)

func GetCurrentThread() (pseudoHandle windows.Handle, err error) {
	r0, _, e1 := syscall.Syscall(procGetCurrentThread.Addr(), 0, 0, 0, 0)
	pseudoHandle = windows.Handle(r0)
	if pseudoHandle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func OpenThreadToken(h windows.Handle, access uint32, self bool, token *windows.Token) (err error) {
	var _p0 uint32
	if self {
		_p0 = 1
	} else {
		_p0 = 0
	}
	r1, _, e1 := syscall.Syscall6(procOpenThreadToken.Addr(), 4, uintptr(h), uintptr(access), uintptr(_p0), uintptr(unsafe.Pointer(token)), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func ImpersonateSelf() (err error) {
	r0, _, e1 := syscall.Syscall(procImpersonateSelf.Addr(), 1, uintptr(2), 0, 0)
	if r0 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func RevertToSelf() (err error) {
	r0, _, e1 := syscall.Syscall(procRevertToSelf.Addr(), 0, 0, 0, 0)
	if r0 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func OpenCurrentThreadToken() (windows.Token, error) {
	if e := ImpersonateSelf(); e != nil {
		return 0, e
	}
	defer RevertToSelf()
	t, e := GetCurrentThread()
	if e != nil {
		return 0, e
	}
	var tok windows.Token
	e = OpenThreadToken(t, windows.TOKEN_QUERY, true, &tok)
	if e != nil {
		return 0, e
	}
	return tok, nil
}
