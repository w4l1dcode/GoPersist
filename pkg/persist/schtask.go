package persist

import (
	"fmt"
	"github.com/go-ole/go-ole/oleutil"
	"time"
)

type SchTask struct {
	taskName   string
	command    string
	commandArg string
	trigger    string
}

func NewSchTask(taskName, command, commandArg, trigger string) *SchTask {
	return &SchTask{
		taskName:   taskName,
		command:    command,
		commandArg: commandArg,
		trigger:    trigger,
	}
}

func (s *SchTask) CreateTask() error {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("Schedule.Service")
	if err != nil {
		return err
	}
	defer unknown.Release()

	service, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer service.Release()

	_, err = oleutil.CallMethod(service, "Connect")
	if err != nil {
		return err
	}

	taskFolder, err := oleutil.CallMethod(service, "GetFolder", "\\")
	if err != nil {
		return err
	}
	defer taskFolder.ToIDispatch().Release()

	// Create Task Definition
	taskDef, err := oleutil.CallMethod(service, "NewTask", 0)
	if err != nil {
		return err
	}
	defer taskDef.ToIDispatch().Release()

	// Set Task Settings
	//settings, err := oleutil.GetProperty(taskDef.ToIDispatch(), "Settings")
	//if err != nil {
	//	return err
	//}
	//oleutil.PutProperty(settings.ToIDispatch(), "DisallowStartIfOnBatteries", false)
	//oleutil.PutProperty(settings.ToIDispatch(), "StopIfGoingOnBatteries", false)

	// Set Task Trigger based on options
	triggerCollection, err := oleutil.GetProperty(taskDef.ToIDispatch(), "Triggers")
	if err != nil {
		return err
	}
	defer triggerCollection.ToIDispatch().Release()

	switch s.trigger {
	case "daily":
		trigger, err := oleutil.CallMethod(triggerCollection.ToIDispatch(), "Create", 1) // 1 is daily trigger type
		if err != nil {
			return err
		}
		triggerDispatch := trigger.ToIDispatch()
		defer triggerDispatch.Release()

		startTime := time.Now().Add(10 * time.Hour)
		oleutil.PutProperty(triggerDispatch, "StartBoundary", startTime.Format(time.RFC3339))
		oleutil.PutProperty(triggerDispatch, "DaysInterval", 1)

	case "hourly":
		trigger, err := oleutil.CallMethod(triggerCollection.ToIDispatch(), "Create", 1) // Time Trigger
		if err != nil {
			return err
		}
		triggerDispatch := trigger.ToIDispatch()
		defer triggerDispatch.Release()

		oleutil.PutProperty(triggerDispatch, "Repetition.Interval", "PT1H") // every hour

	case "logon":
		trigger, err := oleutil.CallMethod(triggerCollection.ToIDispatch(), "Create", 9) // 9 is logon trigger type
		if err != nil {
			return err
		}
		triggerDispatch := trigger.ToIDispatch()
		defer triggerDispatch.Release()
	}

	// Set Action
	actionCollection, err := oleutil.GetProperty(taskDef.ToIDispatch(), "Actions")
	if err != nil {
		return err
	}
	defer actionCollection.ToIDispatch().Release()

	action, err := oleutil.CallMethod(actionCollection.ToIDispatch(), "Create", 0) // 0 is action type execute
	if err != nil {
		return err
	}
	actionDispatch := action.ToIDispatch()
	defer actionDispatch.Release()

	oleutil.PutProperty(actionDispatch, "Path", s.command)
	if s.commandArg != "" {
		oleutil.PutProperty(actionDispatch, "Arguments", s.commandArg)
	}

	// Register Task
	rootFolder := taskFolder.ToIDispatch()
	_, err = oleutil.CallMethod(rootFolder, "RegisterTaskDefinition", s.taskName, taskDef.ToIDispatch(), 6, nil, nil, 3) // 6 = replace existing task, 3 = highest privilege
	if err != nil {
		return err
	}

	fmt.Printf("Task '%s' created successfully!\n", s.taskName)
	return nil
}

func (s *SchTask) RemoveTask() error {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("Schedule.Service")
	if err != nil {
		return err
	}
	defer unknown.Release()

	service, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer service.Release()

	_, err = oleutil.CallMethod(service, "Connect")
	if err != nil {
		return err
	}

	taskFolder, err := oleutil.CallMethod(service, "GetFolder", "\\")
	if err != nil {
		return err
	}
	defer taskFolder.ToIDispatch().Release()

	_, err = oleutil.CallMethod(taskFolder.ToIDispatch(), "DeleteTask", s.taskName, 0)
	if err != nil {
		return err
	}

	fmt.Printf("Task '%s' deleted successfully!\n", s.taskName)
	return nil
}

func (s *SchTask) CheckTask() error {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("Schedule.Service")
	if err != nil {
		return err
	}
	defer unknown.Release()

	service, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer service.Release()

	_, err = oleutil.CallMethod(service, "Connect")
	if err != nil {
		return err
	}

	taskFolder, err := oleutil.CallMethod(service, "GetFolder", "\\")
	if err != nil {
		return err
	}
	defer taskFolder.ToIDispatch().Release()

	task, err := oleutil.CallMethod(taskFolder.ToIDispatch(), "GetTask", s.taskName)
	if err != nil {
		fmt.Printf("Task '%s' does not exist.\n", s.taskName)
		return nil
	}
	defer task.ToIDispatch().Release()

	fmt.Printf("Task '%s' exists!\n", s.taskName)
	return nil
}
