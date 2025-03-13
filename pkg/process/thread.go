package process

import (
	"sync"

	"github.com/cxpsemea/Cx1ClientGo"
	"github.com/sirupsen/logrus"
)

type TestThread struct {
	Id int
}

type TestDirector struct {
	Config    *TestConfig
	Lock      sync.Mutex
	TestIndex int
}

func NewRunner(id int, dir *TestDirector, cx1client *Cx1ClientGo.Cx1Client, logger *logrus.Logger, Config *TestConfig, out chan<- *[]TestResult) {
	logger.Infof("Starting thread %d", id)

	all_results := []TestResult{}

	for {
		testSet := dir.GetNextTestSet()
		if testSet == nil {
			break
		}
		logger.Infof("Thread %d picks up job set: %v [%v]", id, testSet.Name, testSet.File)
		results := testSet.RunTests(cx1client, logger, Config, nil, id)
		all_results = append(all_results, results...)
	}

	out <- &all_results

	logger.Infof("Finished thread %d", id)
}

func NewDirector(Config *TestConfig) TestDirector {
	return TestDirector{Config: Config, TestIndex: 0}
}

func (d *TestDirector) GetNextTestSet() *TestSet {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	returnIndex := d.TestIndex
	if returnIndex >= len(d.Config.Tests) {
		return nil
	}

	d.TestIndex++

	return &d.Config.Tests[returnIndex]
}
