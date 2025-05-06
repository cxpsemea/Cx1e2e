package types

import (
	"fmt"
	"strings"
)

func (c CRUDTest) IsNegative() bool {
	return c.FailTest
}

func (c CRUDTest) IsType(CRUD string) bool {
	switch CRUD {
	case OP_CREATE:
		return strings.Contains(c.Test, "C")
	case OP_READ:
		return strings.Contains(c.Test, "R")
	case OP_UPDATE:
		return strings.Contains(c.Test, "U")
	case OP_DELETE:
		return strings.Contains(c.Test, "D")
	}
	return false
}

func (c CRUDTest) GetSource() string {
	return c.TestSource
}

func (c CRUDTest) GetID() uint {
	return c.TestID
}

func (c CRUDTest) GetFlags() []string {
	return c.Flags
}

func (c CRUDTest) GetVersionStr() string {
	versions := []string{}
	if c.Version.CxOne.IsSet() {
		versions = append(versions, fmt.Sprintf("Cx1 version %v", c.Version.CxOne.String()))
	}
	if c.Version.SAST.IsSet() {
		versions = append(versions, fmt.Sprintf("SAST version %v", c.Version.SAST.String()))
	}
	if c.Version.IAC.IsSet() {
		versions = append(versions, fmt.Sprintf("IAC version %v", c.Version.IAC.String()))
	}
	return strings.Join(versions, ", ")
}

func (c CRUDTest) GetVersion() ProductVersion {
	return c.Version
}

func (c CRUDTest) IsForced() bool {
	return c.ForceRun
}

func (c CRUDTest) OnFail() FailAction {
	return c.OnFailAction
}

func (c CRUDTest) GetCurrentThread() int {
	return c.ActiveThread
}
