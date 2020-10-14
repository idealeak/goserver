package admin

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/schedule"
	"github.com/idealeak/goserver/core/utils"
)

// MyAdminApp is the default AdminApp used by admin module.
var MyAdminApp *AdminApp

// AdminIndex is the default http.Handler for admin module.
// it matches url pattern "/".
func AdminIndex(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("Welcome to Admin Dashboard\n"))
	rw.Write([]byte("There are servral functions:\n"))
	rw.Write([]byte(fmt.Sprintf("1. Get runtime profiling data by the pprof, http://%s:%d/prof\n", Config.AdminHttpAddr, Config.AdminHttpPort)))
	rw.Write([]byte(fmt.Sprintf("2. Get healthcheck result from http://%s:%d/healthcheck\n", Config.AdminHttpAddr, Config.AdminHttpPort)))
	rw.Write([]byte(fmt.Sprintf("3. Get current task infomation from task http://%s:%d/task \n", Config.AdminHttpAddr, Config.AdminHttpPort)))
	rw.Write([]byte(fmt.Sprintf("4. To run a task passed a param http://%s:%d/runtask\n", Config.AdminHttpAddr, Config.AdminHttpPort)))
	rw.Write([]byte(fmt.Sprintf("5. Get all confige & router infomation http://%s:%d/listconf\n", Config.AdminHttpAddr, Config.AdminHttpPort)))

}

// ListConf is the http.Handler of displaying all configuration values as key/value pair.
// it's registered with url pattern "/listconf" in admin module.
func ListConf(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("unimpletement"))
}

// ProfIndex is a http.Handler for showing profile command.
// it's in url pattern "/prof" in admin module.
func ProfIndex(rw http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	command := r.Form.Get("command")
	if command != "" {
		utils.ProcessInput(command, rw)
	} else {
		rw.Write([]byte("request url like '/prof?command=lookup goroutine'\n"))
		rw.Write([]byte("the command have below types:\n"))
		rw.Write([]byte("1. lookup goroutine\n"))
		rw.Write([]byte("2. lookup heap\n"))
		rw.Write([]byte("3. lookup threadcreate\n"))
		rw.Write([]byte("4. lookup block\n"))
		rw.Write([]byte("5. start cpuprof\n"))
		rw.Write([]byte("6. stop cpuprof\n"))
		rw.Write([]byte("7. get memprof\n"))
		rw.Write([]byte("8. gc summary\n"))
		rw.Write([]byte("9. logic statistics\n"))
	}
}

// Healthcheck is a http.Handler calling health checking and showing the result.
// it's in "/healthcheck" pattern in admin module.
func Healthcheck(rw http.ResponseWriter, req *http.Request) {
	defer utils.DumpStackIfPanic("Admin Healthcheck")
	for name, h := range utils.AdminCheckList {
		if err := h.Check(); err == nil {
			fmt.Fprintf(rw, "%s : ok\n", name)
		} else {
			fmt.Fprintf(rw, "%s : %s\n", name, err.Error())
		}
	}
}

// TaskStatus is a http.Handler with running task status (task name, status and the last execution).
// it's in "/task" pattern in admin module.
func TaskStatus(rw http.ResponseWriter, req *http.Request) {
	tasks := schedule.GetAllTask()
	for tname, tk := range tasks {
		fmt.Fprintf(rw, "%s:%s:%s", tname, tk.GetStatus(), tk.GetPrev().String())
	}
}

type TaskRunResult struct {
	Code int
	Err  string
}

// RunTask is a http.Handler to run a Task from the "query string.
// the request url likes /runtask?taskname=sendmail.
func RunTask(rw http.ResponseWriter, req *http.Request) {
	defer req.ParseForm()
	taskname := req.Form.Get("taskname")
	trr := &TaskRunResult{}
	t := schedule.GetTask(taskname)
	if t != nil {
		err := t.Run()
		if err != nil {
			trr.Code = 1
			trr.Err = err.Error()
		} else {
			trr.Code = 0
		}
	} else {
		trr.Err = fmt.Sprintf("there's no task which named:%s", taskname)
		trr.Code = 2
	}
	b, _ := json.Marshal(trr)
	fmt.Println(string(b[:]))
	rw.Write(b)
}

// AdminApp is an http.HandlerFunc map used as AdminApp.
type AdminApp struct {
	routers map[string]http.HandlerFunc
}

// Route adds http.HandlerFunc to AdminApp with url pattern.
func (admin *AdminApp) Route(pattern string, f http.HandlerFunc) {
	admin.routers[pattern] = f
}

// Start AdminApp http server.
// Its addr is defined in configuration file as adminhttpaddr and adminhttpport.
func (admin *AdminApp) Start(AdminHttpAddr string, AdminHttpPort int) {
	for p, f := range admin.routers {
		http.Handle(p, f)
	}

	addr := fmt.Sprintf("%s:%d", AdminHttpAddr, AdminHttpPort)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Logger.Critical("Admin Listen error: ", err)
		return
	}

	logger.Logger.Infof("Admin Serve: %s", l.Addr())

	go func() {
		server := &http.Server{}
		err = server.Serve(l)
		if err != nil {
			logger.Logger.Critical("Admin Serve: ", err)
		}
	}()
}

func init() {
	MyAdminApp = &AdminApp{
		routers: make(map[string]http.HandlerFunc),
	}
	//MyAdminApp.Route("/", AdminIndex)
	//MyAdminApp.Route("/prof", ProfIndex)
	//MyAdminApp.Route("/healthcheck", Healthcheck)
	//MyAdminApp.Route("/task", TaskStatus)
	//MyAdminApp.Route("/runtask", RunTask)
	//MyAdminApp.Route("/listconf", ListConf)
}
