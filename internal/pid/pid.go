package pid

import (
	"fmt"
	"os"
)

func WritePID(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Error open PID-file(name=%v):%v", filename, err)
	}
	defer file.Close()
	_, err2 := file.Write([]byte(fmt.Sprint(os.Getpid())))
	if err2 != nil {
		return fmt.Errorf("Error write file=(%v), data=(%v):%v", filename, os.Getpid(), err)
	}
	return nil
}
