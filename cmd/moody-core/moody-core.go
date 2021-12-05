package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/akamensky/argparse"
	"github.com/antima/moody-core/pkg/api"
	"github.com/antima/moody-core/pkg/http"
	"github.com/antima/moody-core/pkg/mqtt"
)

const (
	name    = "moody-core"
	desc    = "the moody core engine"
	version = "0.1.0"

	defaultBrokerString = "tcp://localhost:1883"
	defaultServiceDir   = "./services"
	defaultApiPort      = ":8080"

	versionHelp    = "Print out the current version"
	brokerHelp     = "Pass the broker connection string in the <scheme>://<host>:<port> format"
	apiPortHelp    = "Start the HTTP API server on the specified port, in the :<port> format"
	serviceDirHelp = "Pass the directory from where to load the services"
	configHelp     = "Pass the location of a file specifying the needed configurations in json format"

	antimaLogo = `
               -/////////////////:                
           .////-               .:///:            
         ///.                        ://:         
      ./O-            .///:             /O/       
     /O.              OOOO0-              /O,     
   .O:                ./O//                 O/    
  ,0.                   /                    /O   
 .0.                    /                     /O  
 0,                ,:   /      /O:             O: 
:O                   ,  /    .-.               .# 
O:                .   ,./-.--                   0-
#                  ..-::O/:....:                O/
#             :,......:/O//-                    O/
O:            -.    -//.,,.//                   0-
:O                ,/:, ..,  .//                .# 
 0,             ,/:/:    :/    //.,.           O: 
 .0.      ./00OO: .:-    .:     :0OO:         /O  
  ,0.     /#####:               .///.        /O   
   .O:    .O##0/                            O/    
     /O.                                  /O,     
      ./O,                              /O/       
         ///.                        ://:         
           .////,               .:///:            
               ./////////////////:                `
)

type Config struct {
	BrokerString string `json:"brokerString"`
	ApiPort      string `json:"apiPort"`
	ServiceDir   string `json:"serviceDir"`
}

func fromConfigFile(configFilePath string) (*Config, error) {
	fileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(fileBytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func startCore(brokerString string, serviceDir string, apiPort string) {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	fmt.Printf("%s\n\n", antimaLogo)
	fmt.Printf("\tmoody-core v%s - Powered by Antima.it\n", version)

	deviceTable := http.NewDeviceList()
	dataTable := mqtt.NewDataTable()
	serviceMap := mqtt.NewServiceMap()

	apiServer := api.StartMoodyApi(deviceTable, serviceMap, apiPort)
	monitor := http.NewMonitor(deviceTable)
	mqtt.StartServiceManager(serviceDir, serviceMap, dataTable)
	mqtt.StartMqttManager(brokerString, dataTable)
	monitor.Start()
	<-quit
	fmt.Println("moody-core - stopping")
	mqtt.StopMqttManager()
	monitor.Stop()
	api.StopMoodyApi(apiServer)
	fmt.Println("Bye!")
}

func main() {
	parser := argparse.NewParser(name, desc)

	printVersion := parser.Flag("v", "version", &argparse.Options{
		Help: versionHelp,
	})

	brokerString := parser.String("b", "broker", &argparse.Options{
		Help:    brokerHelp,
		Default: defaultBrokerString,
	})

	apiPort := parser.String("p", "port", &argparse.Options{
		Help:    apiPortHelp,
		Default: defaultApiPort,
	})

	serviceDir := parser.String("s", "service-dir", &argparse.Options{
		Help:    serviceDirHelp,
		Default: defaultServiceDir,
	})

	configFile := parser.String("c", "config", &argparse.Options{
		Help: configHelp,
	})

	err := parser.Parse(os.Args)
	if err != nil {
		log.Fatal(parser.Usage(err))
	}

	if *printVersion {
		fmt.Println(fmt.Sprintf("%s - version v%s", name, version))
		return
	}

	if *configFile != "" {
		config, err := fromConfigFile(*configFile)
		if err != nil {
			log.Fatal(err)
		}
		*brokerString = config.BrokerString
		*serviceDir = config.ServiceDir
		*apiPort = config.ApiPort
	}

	startCore(*brokerString, *serviceDir, *apiPort)
}
