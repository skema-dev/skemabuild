package console

import "fmt"

func Info(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func Infof(format string, args ...any) {
	fmt.Printf(format, args...)
}

func Errorf(format string, args ...any) error {
	return fmt.Errorf("[ERROR]"+format, args...)
}
