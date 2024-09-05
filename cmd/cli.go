package main

import (
	"flag"
	"fmt"
	"github.com/w4l1dcode/GoPersist/pkg/GoPersist"
	"golang.org/x/sys/windows"
	"log"
	"os"
)

func main() {
	// Define command-line flags
	var (
		technique   string
		action      string
		taskName    string
		schCommand  string
		schArgs     string
		trigger     string
		startupCmd  string
		startupArgs string
		serviceName string
		serviceDesc string
		servicePath string
		serviceArgs string
		regKey      string
		regValue    string
		regCmd      string
		regArgs     string
	)

	// Define flags for the technique to use
	flag.StringVar(&technique, "t", "", "Technique to use: 'schtask', 'service', 'reg', 'startup'")

	// Define flags for action (add or remove)
	flag.StringVar(&action, "action", "", "Action: 'add' to add or 'remove' to delete")

	// Flags for scheduled task technique
	flag.StringVar(&schCommand, "sch-cmd", "", "Command for the scheduled task")
	flag.StringVar(&schArgs, "sch-args", "", "Arguments for the scheduled task")
	flag.StringVar(&taskName, "sch-name", "", "File name for the scheduled task")
	flag.StringVar(&trigger, "trigger", "", "Trigger for the scheduled task")

	// Flags for startup technique
	flag.StringVar(&startupCmd, "startup-cmd", "", "Command (path) for the startup entry")
	flag.StringVar(&startupArgs, "startup-args", "", "Arguments for the startup command")

	// Flags for service technique
	flag.StringVar(&serviceName, "svc-name", "", "Name of the service")
	flag.StringVar(&serviceDesc, "svc-desc", "", "Description of the service")
	flag.StringVar(&servicePath, "svc-path", "", "Path to the executable for the service")
	flag.StringVar(&serviceArgs, "svc-args", "", "Arguments for the service")

	// Flags for registry technique
	flag.StringVar(&regKey, "reg-key", "", "Registry key path")
	flag.StringVar(&regValue, "reg-val", "", "Registry value name")
	flag.StringVar(&regCmd, "reg-cmd", "", "Path to the executable or PowerShell command for registry persistence")
	flag.StringVar(&regArgs, "reg-args", "", "Arguments for the PowerShell command")

	// Parse the command-line arguments
	flag.Parse()

	// Validate the combination of flags
	if technique == "" {
		fmt.Println("Error: Technique is required. Use -tech flag to specify the technique.")
		flag.Usage()
		os.Exit(1)
	}

	if action != "add" && action != "remove" {
		fmt.Println("Error: Action must be 'add' or 'remove'.")
		flag.Usage()
		os.Exit(1)
	}

	switch technique {
	case "schtask":
		if action == "add" {
			if schCommand == "" || schArgs == "" || taskName == "" {
				log.Println("Error: -sch-cmd (command), -sch-args (arguments), and -file (file-name) are required for adding a scheduled task.")
				flag.Usage()
				os.Exit(1)
			}
			task := GoPersist.NewSchTask(taskName, schCommand, schArgs, trigger)
			err := task.CreateTask()
			if err != nil {
				log.Fatalf("Failed to create task: %v", err)
			}
			log.Println("Scheduled task created successfully.")
		} else if action == "remove" {
			task := GoPersist.NewSchTask(taskName, "", "", "")
			err := task.RemoveTask()
			if err != nil {
				log.Fatalf("Failed to remove task: %v", err)
			}
			log.Println("Scheduled task removed successfully.")
		}

	case "startup":
		if action == "add" {
			if startupCmd == "" || taskName == "" {
				log.Println("Error: -startup-cmd (command) and -file (file-name) are required for adding a startup entry.")
				flag.Usage()
				os.Exit(1)
			}
			if startupArgs != "" {
				err := GoPersist.CreateStartupBatchFile(startupCmd, startupArgs, taskName)
				if err != nil {
					log.Fatalf("Error creating startup entry: %v", err)
				}
				log.Println("Startup entry created successfully.")
			} else {
				err := GoPersist.DropFileToStartup(startupCmd, taskName)
				if err != nil {
					log.Fatalf("Error dropping file to startup: %v", err)
				}
			}
		} else if action == "remove" {
			if taskName == "" {
				log.Println("Error: -file (file-name) is required for removing a file from Startup.")
				flag.Usage()
				os.Exit(1)
			}
			err := GoPersist.RemoveFileFromStartup(taskName)
			if err != nil {
				log.Fatalf("Error removing file from Startup: %v", err)
			}
			log.Println("Startup entry removed successfully.")
		}

	case "service":
		// Check if user is admin
		var sid *windows.SID
		err := windows.AllocateAndInitializeSid(&windows.SECURITY_NT_AUTHORITY, 2, windows.SECURITY_BUILTIN_DOMAIN_RID, windows.DOMAIN_ALIAS_RID_ADMINS, 0, 0, 0, 0, 0, 0, &sid)
		if err != nil {
			panic(err)
		}

		token, err := GoPersist.OpenCurrentThreadToken()
		if err != nil {
			panic(err)
		}

		member, err := token.IsMember(sid)
		if err != nil {
			panic(err)
		}
		if member == false {
			log.Fatalf("You need admin permissions to create a service.")
		}

		if action == "add" {
			if serviceName == "" || serviceDesc == "" || servicePath == "" {
				log.Println("Error: -svc-name (service-name), -svc-desc (description), and -svc-path (executable-path) are required for adding a service.")
				flag.Usage()
				os.Exit(1)
			}
			err := GoPersist.CreateService(serviceName, serviceDesc, servicePath, serviceArgs)
			if err != nil {
				log.Fatalf("Error creating service: %v", err)
			}
			log.Println("Service created successfully.")
			err = GoPersist.StartService(serviceName)
			if err != nil {
				log.Fatalf("Error starting service: %v", err)
			}
		} else if action == "remove" {
			if serviceName == "" {
				log.Println("Error: -svc-name (service-name) is required for deleting a service.")
				flag.Usage()
				os.Exit(1)
			}
			err := GoPersist.DeleteService(serviceName)
			if err != nil {
				log.Fatalf("Error deleting service: %v", err)
			}
			fmt.Println("Service deleted successfully.")
		} else if startupCmd != "" && serviceName != "" {
			if taskName == "" {
				log.Println("Error: -file (file-name) is required for creating a batch file for the service.")
				flag.Usage()
				os.Exit(1)
			}
			err := GoPersist.CreateServiceBatchFile(startupCmd, serviceArgs, taskName)
			if err != nil {
				log.Fatalf("Error creating service batch file: %v", err)
			}
			log.Println("Service batch file created successfully.")
		}

	case "reg":
		if action == "add" {
			if regKey == "" || regValue == "" || regCmd == "" {
				fmt.Println("Error: -reg-key (registry-key-path), -reg-val (registry-value), and -reg-cmd (path) are required for adding registry persistence.")
				flag.Usage()
				os.Exit(1)
			}
			err := GoPersist.AddRegistryPersistence(regKey, regValue, regCmd, regArgs)
			if err != nil {
				log.Fatalf("Error adding registry persistence: %v", err)
			}
			fmt.Println("Registry persistence added successfully.")
		} else if action == "remove" {
			if regKey == "" || regValue == "" {
				log.Println("Error: -reg-key (registry-key-path) and -reg-val (registry-value) are required for removing registry persistence.")
				flag.Usage()
				os.Exit(1)
			}
			err := GoPersist.RemoveRegistryPersistence(regKey, regValue)
			if err != nil {
				log.Fatalf("Error removing registry persistence: %v", err)
			}
			log.Println("Registry persistence removed successfully.")
		}

	default:
		log.Println("Error: Unknown technique specified. Use 'schtask', 'service', 'reg', or 'startup'.")
		flag.Usage()
		os.Exit(1)
	}
}
