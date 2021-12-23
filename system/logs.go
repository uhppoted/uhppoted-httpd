package system

func Logs(start, count int) []interface{} {
	sys.RLock()
	defer sys.RUnlock()

	return sys.logs.AsObjects(start, count)
}
