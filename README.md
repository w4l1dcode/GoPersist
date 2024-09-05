# GoPersist
GoPersist is a command-line utility designed for managing persistence techniques on Windows systems (based on [SharPersist](https://github.com/mandiant/SharPersist)). It supports adding and removing persistence using scheduled tasks, startup entries, Windows services, and registry modifications.

## Features
- Scheduled Tasks: Create and remove scheduled tasks with specified commands and arguments.
- Startup Entries: Add and remove startup entries by creating batch files or directly dropping files into the startup folder.
- Windows Services: Create, start, and remove Windows services with specified executables and arguments.
- Registry Persistence: Add and remove registry entries for persistence.

- ## Installation
To use GoPersist, you need to have Go installed on your system. Follow these steps to build and use the program:

### Clone the repository:
```sh
git clone https://github.com/yourusername/GoPersist.git
cd cmd
```

### Build the project:
```sh
go build -o GoPersist cli.go
```

### Run the program:
```sh
./GoPersist --help
```

## Usage
### General Syntax
```sh
GoPersist -t <technique> -action <add|remove> [options]
```

### Available Techniques
schtask: Manage scheduled tasks.
startup: Manage startup entries.
service: Manage Windows services.
reg: Manage registry persistence.

### Command-Line Flags
#### Scheduled Task
- -t schtask
    - -action add or -action remove
    - -sch-cmd : Command to execute (required for add action).
    - -sch-args : Arguments for the command (required for add action).
    - -file : File name for the scheduled task (required for add action).

##### Example:

- Add a scheduled task:

```sh
GoPersist -t schtask -action add -sch-cmd "C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe" -sch-args "Start-Up notepad.exe" -sch-name "MyTask" -trigger "daily"
```
- Remove a scheduled task:

```sh
GoPersist -t schtask -action remove -sch-name "MyTask"
```

#### Startup Entry
- -t startup
    - -action add or -action remove
    - -startup-cmd : Command (path) to add to startup (required for add action).
    - -startup-args : Arguments for the command (optional).
    - -file : File name for the startup entry (required for add action).

##### Example:

- Add a startup entry:
```sh
GoPersist -t startup -action add -startup-cmd "C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe" -startup-args "Start-Up notepad.exe" -file "MyStartup"
```

- Remove a startup entry:

```sh
GoPersist -t startup -action remove -file "MyStartup"
```

#### Windows Service

- -t service
    - -action add or -action remove
    - -svc-name : Name of the service (required for both add and remove actions).
    - -svc-desc : Description of the service (required for add action).
    - -svc-path : Path to the executable for the service (required for add action).
    - -svc-args : Arguments for the service (optional).

##### Example:

- Add a Windows service:

```sh
GoPersist -t service -action add -svc-name "NotepadService" -svc-desc "Notepad Service" -svc-path "C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe" -svc-args "Start-Up notepad.exe"
```

- Remove a Windows service:
```sh
GoPersist -t service -action remove -svc-name "NotepadService"
```

#### Registry Persistence (reg)
- -t reg
    - -action add or -action remove
    - -reg-key : Registry key path (required for both add and remove actions).
    - -reg-val : Registry value name (required for both add and remove actions).
    - -reg-cmd : Path to the executable or PowerShell command (required for add action).
    - -reg-args : Arguments for the PowerShell command (optional).

##### Example:

- Add a registry entry:

```sh
GoPersist -t reg -action add -reg-key "Software\Microsoft\Windows\CurrentVersion\Run" -reg-val "MyValue" -reg-cmd "C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe" -reg-args "Start-Up notepad.exe"
```

- Remove a registry entry:

```sh
GoPersist -t reg -action remove -reg-key "Software\Microsoft\Windows\CurrentVersion\Run" -reg-val "MyValue"
```

## Contributing
Feel free to contribute to this project by opening issues or submitting pull requests.

## License
This project is licensed under the MIT License. See the LICENSE file for details.