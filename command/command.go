package command

import "strings"

type Command struct {
	name string
	args [][]byte
}

func New(name string, args [][]byte) Command {
	return Command{strings.ToUpper(name), args}
}

func (command Command) Name() string {
	return command.name
}

func (command Command) Args() []interface{} {
	args := make([]interface{}, len(command.args))
	for i, arg := range command.args {
		args[i] = arg
	}
	return args
}
