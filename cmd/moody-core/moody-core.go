package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/akamensky/argparse"
	"github.com/antima/moody-core/pkg/api"
	"github.com/antima/moody-core/pkg/http"
	"github.com/antima/moody-core/pkg/mqtt"
)

const (
	name    = "moody-core"
	desc    = "TODO"
	version = "0.1.0"

	versionHelp    = ""
	brokerHelp     = ""
	apiPortHelp    = ""
	serviceDirHelp = ""
	configHelp     = ""

	defaultBrokerString = "tcp://localhost:1883"
	defaultServiceDir   = "./services"
	defaultApiPort      = 8080

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
	ApiPort      int    `json:"apiPort"`
	ServiceDir   string `json:"serviceDir"`
}

func fromConfigFile(configFilePath string) (*Config, error) {
	fileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	var config *Config
	if err := json.Unmarshal(fileBytes, config); err != nil {
		return nil, err
	}

	return config, nil
}

func startCore(brokerString string, serviceDir string, apiPort int) {
	fmt.Println(antimaLogo + "\n")
	fmt.Println("\tmoody-core - Powered by Antima.it (c) 2021")

	deviceTable := http.NewDeviceList()
	dataTable := mqtt.NewDataTable()
	mqtt.StartServiceManager(serviceDir, dataTable)
	mqtt.StartMqttManager(brokerString, dataTable)
	http.NewMonitor(deviceTable).Start()
	api.MoodyApi(deviceTable, fmt.Sprintf(":%d", apiPort))
	fmt.Println("moody-core - stopping")
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

	apiPort := parser.Int("p", "port", &argparse.Options{
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
	}

	if *configFile != "" {
		config, err := fromConfigFile(*configFile)
		if err != nil {
			log.Fatal(parser.Usage(err))
		}
		*brokerString = config.BrokerString
		*serviceDir = config.ServiceDir
		*apiPort = config.ApiPort
	}

	startCore(*brokerString, *serviceDir, *apiPort)
}
