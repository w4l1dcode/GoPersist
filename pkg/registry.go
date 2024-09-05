package pkg

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
)

// AddRegistryPersistence adds a value to the specified registry key.
// Handles both executable paths and PowerShell commands with arguments.
func AddRegistryPersistence(registryKeyPath, valueName, command, args string) error {
	// Open or create the registry key
	key, _, err := registry.CreateKey(registry.CURRENT_USER, registryKeyPath, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to open or create registry key: %w", err)
	}
	defer key.Close()

	// Format the command with arguments
	fullCommand := command
	if args != "" {
		fullCommand = fmt.Sprintf("%s %s", command, args)
	}

	// Set the registry value
	err = key.SetStringValue(valueName, fullCommand)
	if err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	fmt.Printf("Registry persistence for %s added successfully.\n", valueName)
	return nil
}

// RemoveRegistryPersistence removes a registry entry to stop persisting an application.
func RemoveRegistryPersistence(registryKeyPath, valueName string) error {
	// Open the registry key with write access.
	k, err := registry.OpenKey(registry.CURRENT_USER, registryKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer k.Close()

	// Delete the value.
	err = k.DeleteValue(valueName)
	if err != nil {
		return fmt.Errorf("failed to delete registry value: %w", err)
	}

	fmt.Printf("Successfully deleted registry entry: %s\\%s\n", registryKeyPath, valueName)
	return nil
}
