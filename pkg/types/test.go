package types

import "strings"

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

func (c CRUDTest) GetFlags() []string {
	return c.Flags
}

func (c CRUDTest) GetVersion() string {
	return c.Version
}

func (c CRUDTest) IsForced() bool {
	return c.ForceRun
}

func (c CRUDTest) OnFail() FailAction {
	return c.OnFailAction
}
