package cmdline

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/idealeak/goserver/core/module"
)

var cmdpool = make(map[string]cmdExecuter)

type cmdGoroutineWapper struct {
}

type CmdArg struct {
	Flag         string
	SimplifyFlag string
	Required     bool
}

type CmdArgParser struct {
	cmeKV map[string]string
}

type cmdExecuter interface {
	Execute(args []string)
	ShowUsage()
}

func NewCmdArgParser(args []string) *CmdArgParser {
	parser := &CmdArgParser{
		cmeKV: make(map[string]string),
	}
	for _, arg := range args {
		kv := strings.Split(arg, "=")
		if len(kv) == 2 {
			k := kv[0]
			v := kv[1]
			parser.cmeKV[k] = v
		}
	}
	return parser
}

func RegisteCmd(cmdName string, executer cmdExecuter) {
	cmdpool[strings.ToLower(cmdName)] = executer
}

func (cw *cmdGoroutineWapper) Start() {
	if Config.SupportCmdLine {
		go func() {
			var (
				reader   = bufio.NewReader(os.Stdin)
				cmdLine  []byte
				isPrefix bool
				err      error
			)

			for {
				cmdLine, isPrefix, err = reader.ReadLine()
				if err == nil && isPrefix == false {
					params := strings.Split(string(cmdLine), " ")
					if len(params) >= 1 {
						cmdName := strings.ToLower(params[0])
						if cmdExecute, exist := cmdpool[cmdName]; exist {
							PostCmd(module.AppModule.Object, cmdExecute, params[1:])
						}
					}
				}
				time.Sleep(time.Second)
			}
		}()
	}
}

func (this *CmdArgParser) ExtraIntArg(arg *CmdArg, val *int) {
	if v, exist := this.cmeKV[arg.SimplifyFlag]; !exist {
		if v, exist := this.cmeKV[arg.Flag]; !exist && arg.Required {
			fmt.Println(arg.Flag, "must be give")
			return
		} else {
			*val, _ = strconv.Atoi(v)
		}
	} else {
		*val, _ = strconv.Atoi(v)
	}
}

func (this *CmdArgParser) ExtraInt64Arg(arg *CmdArg, val *int64) {
	if v, exist := this.cmeKV[arg.SimplifyFlag]; !exist {
		if v, exist := this.cmeKV[arg.Flag]; !exist && arg.Required {
			fmt.Println(arg.Flag, "must be give")
			return
		} else {
			*val, _ = strconv.ParseInt(v, 10, 64)
		}
	} else {
		*val, _ = strconv.ParseInt(v, 10, 64)
	}
}

func (this *CmdArgParser) ExtraStringArg(arg *CmdArg, val *string) {
	if v, exist := this.cmeKV[arg.SimplifyFlag]; !exist {
		if v, exist := this.cmeKV[arg.Flag]; !exist && arg.Required {
			fmt.Println(arg.Flag, "must be give")
			return
		} else {
			*val = v
		}
	} else {
		*val = v
	}
}

func init() {
	//module.RegistePreloadModule(&cmdGoroutineWapper{}, 0)
}
