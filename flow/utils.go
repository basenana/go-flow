package flow

import (
	"fmt"
	"github.com/basenana/go-flow/utils"
	"os"
	"path"
)

var logger = utils.NewLogger("go-flow")

func initFlowWorkDir(base, flowID string) error {
	dir := flowWorkdir(base, flowID)
	s, err := os.Stat(dir)

	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		if s.IsDir() {
			return nil
		}
		return fmt.Errorf("init flow workdir failed: %s not dir", dir)
	}
	return os.MkdirAll(dir, 0755)
}

func cleanUpFlowWorkDir(base, flowID string) error {
	dir := flowWorkdir(base, flowID)
	s, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !s.IsDir() {
		return fmt.Errorf("flow workdir cleanup failed: %s not dir", dir)
	}

	return os.RemoveAll(dir)
}

func flowWorkdir(base, flowID string) string {
	return path.Join(base, "flows", flowID)
}
