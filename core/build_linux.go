package core

/*
const char* build_time(void)
{
	static const char* psz_build_time = "["__DATE__ " " __TIME__ "]";
    return psz_build_time;
}
*/
import "C"

var (
	_linux_buildTime = C.GoString(C.build_time())
)

func buildTime() string {
	return _linux_buildTime
}
