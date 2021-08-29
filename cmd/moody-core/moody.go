package moody_core

func StartCore(port string) {
	ssdpMonitorStart()
	api.moodyApi(port)
	ssdpStop()
}
