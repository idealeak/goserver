package srvlib

var (
	sessionServiceERtable = make(map[int32][]int32)
	serviceSessionERtable = make(map[int32][]int32)
)

var arrER = [][]int32{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0},
	{2, 3, 6, 7, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0},
	{3, 0, 0, 0, 0, 0, 0, 0},
	{6, 0, 0, 0, 0, 0, 0, 0},
}

func init() {
	buildSessionTable()
	buildServiceTable()
}

func buildSessionTable() {
	for k1, v1 := range arrER {
		t := make([]int32, 0, MaxServerType)
		for _, v2 := range v1 {
			if v2 != 0 {
				t = append(t, int32(v2))
			}
		}
		sessionServiceERtable[int32(k1)] = t
	}
}

func buildServiceTable() {
	for k1, v1 := range sessionServiceERtable {
		for _, v2 := range v1 {
			if _, has := serviceSessionERtable[v2]; !has {
				serviceSessionERtable[v2] = make([]int32, 0, MaxServerType)
			}

			serviceSessionERtable[v2] = append(serviceSessionERtable[v2], k1)
		}
	}
}

func SessionCareService(sessionType, serviceType int32) bool {
	if v, has := sessionServiceERtable[sessionType]; has {
		for _, service := range v {
			if service == serviceType {
				return true
			}
		}
	}

	return false
}

func GetCareSessionsByService(serviceType int32) []int32 {
	if v, has := serviceSessionERtable[serviceType]; has {
		return v
	}

	return nil
}

func GetCareServicesBySession(sessionType int32) []int32 {
	if v, has := sessionServiceERtable[sessionType]; has {
		return v
	}

	return nil
}
