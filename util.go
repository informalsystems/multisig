package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// XXX: get the denom
// this is a massive hack. use the chain-registry instead :D
// this just calls `<binary> query staking params` and parses out the staking token denom
func getDenom(binary string) (string, error) {
	cmdArgs := []string{"query", "staking", "params"}
	execCmd := exec.Command(binary, cmdArgs...)
	b, err := execCmd.CombinedOutput()
	if err != nil {
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("call failed")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println(execCmd)
		fmt.Println(string(b))
		return "", err
	}
	fmt.Println(string(b))
	s := strings.Split(string(b), "\n")
	firstLine := s[0]

	s2 := strings.Split(firstLine, " ")
	denom := strings.TrimSpace(s2[1])

	return denom, nil
}
