package moody

func StartCore() {
	ssdpMonitorStart()
	moodyApi()
	ssdpStop()
}
