package cmdline

import (
	"fmt"
)

type helpExecuter struct {
}

func (this helpExecuter) Execute(args []string) {
	if len(args) > 0 {
		if cmde, exist := cmdpool[args[0]]; exist {
			cmde.ShowUsage()
		}
	} else {
		this.ShowUsage()
		fmt.Println("The commands are:")
		for k, _ := range cmdpool {
			if k != "help" {
				fmt.Println("\t", k)
			}
		}
		fmt.Println("Use \"help [command]\" for more information about a command.")
	}
}

func (this helpExecuter) ShowUsage() {
	fmt.Println("Help is a help command like window or linux shell's command")
	fmt.Println("Usage:")
	fmt.Println("\t", "help command")
}

func init() {
	RegisteCmd("help", &helpExecuter{})
}
