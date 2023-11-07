package detail

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"os"
)

func Login() {
	fmt.Println("Log in with your username to access cloud-based software supply-chain risk data from OpenSCA SaaS.")
	fmt.Println("If you don't have an account, please register at https://opensca.xmirror.cn/")

	fmt.Print("Enter username: ")
	username, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print("Enter password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println(err)
		return
	}

	// TODO：登录逻辑
	fmt.Println()
	fmt.Println("username: ", username, "password: ", string(password))

	os.Exit(0)
}
