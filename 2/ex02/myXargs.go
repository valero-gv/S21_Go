package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: myXargs <command>")
		os.Exit(1)
	}
	command := os.Args[1:]

	cmd := exec.Command(command[0], command[1:]...)
	stdin := bufio.NewScanner(os.Stdin)
	var args []string

	for stdin.Scan() {
		line := stdin.Text()
		fmt.Print(line + " ")
		args = append(args, line)
	}

	cmd.Args = append(cmd.Args, args...)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error executing command: %s\n", err)
		os.Exit(1)
	}

	fmt.Print(string(output))
}
