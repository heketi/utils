package main

import (
	"fmt"
	"os"

	"github.com/heketi/utils/ssh"
)

func runDemo(s ssh.SshExecutor) {
	// scp scpdemo
	fmt.Print("Copying scpdemo to server...")
	host := "127.0.0.1:22"
	err := s.CopyPath("scpdemo.go", host, "/tmp/scpdemo-copy.go")
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	fmt.Println("Done")

	// run a few commands
	fmt.Println("Running commands...")
	commands := []string{
		"date",
		"echo \"HELLO\" > /tmp/file",
		"cat /tmp/file",
		"ls -al",
		"rm /tmp/file",
		"rm /tmp/scpdemo-copy.go",
	}

	out, err := s.Exec(host, commands, 10, false)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	fmt.Printf("%+v\n", out)
}

func main() {
	fmt.Println("- Real Demo -")
	s, err := ssh.NewSshExecWithAuth(os.Getenv("USER"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	runDemo(s)

	fmt.Println("- Mock Demo -")
	// Now run with a mock demo
	m := ssh.NewMockSshExecutor()
	m.MockExec = func(host string, commands []string, timeoutMinutes int, useSudo bool) ([]string, error) {
		return []string{
			"In Mock function",
		}, nil
	}
	runDemo(m)
}
