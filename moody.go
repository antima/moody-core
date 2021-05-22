package moody

func StartCore(port string) {
	ssdpMonitorStart()
	moodyApi(port)
	ssdpStop()
}
