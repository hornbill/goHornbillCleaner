package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/hornbill/color"
	"github.com/hornbill/goApiLib"
	"github.com/hornbill/goHornbillHelpers"
)

func main() {

	parseFlags()

	//Does endpoint exist?
	instanceEndpoint := apiLib.GetEndPointFromName(configInstance)
	if instanceEndpoint == "" {
		color.Red("The provided instance ID [" + configInstance + "] could not be found.")
		return
	}

	//Load the configuration file
	cleanerConf = loadConfig()

	//Ask if we want to delete before continuing
	fmt.Println("")
	fmt.Println("===== Hornbill Cleaner Utility v" + toolVer + " =====")
	if !cleanerConf.CleanRequests && !cleanerConf.CleanAssets && !cleanerConf.CleanUsers {
		color.Red("No entity data has been specified for cleansing in " + configFileName)
		return
	}
	fmt.Println("")
	color.Red(" WARNING!")
	color.Red(" This utility will delete all records from the following entities in")
	color.Red(" Hornbill instance: " + configInstance)
	color.Red(" as specified in your configuration file: ")
	fmt.Println("")
	if cleanerConf.CleanRequests {
		color.Magenta(" * Requests (and related data)")
	}
	if cleanerConf.CleanAssets {
		color.Magenta(" * All Assets (and related data)")
	}
	if cleanerConf.CleanUsers {
		color.Magenta(" * All Specified Users (and related data)")
	}
	fmt.Println("")
	fmt.Println("Are you sure you want to permanently delete these records? (yes/no):")
	if hornbillHelpers.ConfirmResponse("") != true {
		return
	}
	color.Red("Are you absolutely sure? Type in the word 'delete' to confirm...")
	if hornbillHelpers.ConfirmResponse("delete") != true {
		return
	}

	//Create new session
	espXmlmc = apiLib.NewXmlmcInstance(configInstance)
	espXmlmc.SetAPIKey(configAPIKey)

	maxResults = getMaxRecordsSetting()

	processRequests()
	processAssets()

	if cleanerConf.CleanUsers {
		for _, v := range cleanerConf.Users {
			deleteUser(v)
		}
	}
	//Reset maxResultsAllowed system setting back to what it was before the process ran
	if resetResults {
		setMaxRecordsSetting(maxResults)
	}
}

func processRequests() {
	//Process Request Records
	espLogger("System Setting Max Results: "+strconv.Itoa(maxResults), "debug")
	if cleanerConf.CleanRequests {
		requestCount := 0
		if len(cleanerConf.RequestReferences) > 0 {
			currentBlock = 1
			requestCount = len(cleanerConf.RequestReferences)
			espLogger("Block Size: "+strconv.Itoa(configBlockSize), "debug")
			requestBlocks := float64(requestCount) / float64(configBlockSize)
			totalBlocks = int(math.Ceil(requestBlocks))

			espLogger("Number of Requests to delete: "+strconv.Itoa(requestCount), "debug")
			color.Green("Number of Requests to delete: " + strconv.Itoa(requestCount))
			if maxResults > 0 && requestCount > maxResults {
				resetResults = true
				//Update maxResultsAllowed system setting to match number of records to be deleted
				setMaxRecordsSetting(requestCount)
			}
			processEntityClean("Requests", configBlockSize)
		} else {
			requestCount = getRecordCount("h_itsm_requests")
			if requestCount > 0 {
				currentBlock = 1
				espLogger("Block Size: "+strconv.Itoa(configBlockSize), "debug")
				requestBlocks := float64(requestCount) / float64(configBlockSize)
				totalBlocks = int(math.Ceil(requestBlocks))
				espLogger("Request Blocks: "+strconv.Itoa(int(requestBlocks)), "debug")
				espLogger("Total Blocks: "+strconv.Itoa(totalBlocks), "debug")

				espLogger("Number of Requests to delete: "+strconv.Itoa(requestCount), "debug")
				if maxResults > 0 && requestCount > maxResults {
					resetResults = true
					//Update maxResultsAllowed system setting to match number of records to be deleted
					setMaxRecordsSetting(requestCount)
				}
				processEntityClean("Requests", configBlockSize)
			} else {
				espLogger("There are no requests to delete.", "debug")
				color.Red("There are no requests to delete.")
			}
		}
	}
}

func processAssets() {
	//Process Asset Records
	if cleanerConf.CleanAssets {
		assetCount := getRecordCount("h_cmdb_assets")
		if assetCount > 0 {
			currentBlock = 1
			assetBlocks := float64(assetCount) / float64(configBlockSize)
			totalBlocks = int(math.Ceil(assetBlocks))
			espLogger("Number of Assets to delete: "+strconv.Itoa(assetCount), "debug")
			if maxResults > 0 && assetCount > maxResults {
				resetResults = true
				//Update maxResultsAllowed system setting to match number of records to be deleted
				setMaxRecordsSetting(assetCount)
			}
			processEntityClean("Asset", configBlockSize)
		} else {
			espLogger("There are no assets to delete.", "debug")
			color.Red("There are no assets to delete.")
		}

		assetLinkCount := getRecordCount("h_cmdb_links")
		if assetLinkCount > 0 {
			currentBlock = 1
			assetLinkBlocks := float64(assetLinkCount) / float64(configBlockSize)
			totalBlocks = int(math.Ceil(assetLinkBlocks))
			espLogger("Number of Asset Links to delete: "+strconv.Itoa(assetLinkCount), "debug")
			if maxResults > 0 && assetLinkCount > maxResults {
				resetResults = true
				//Update maxResultsAllowed system setting to match number of records to be deleted
				setMaxRecordsSetting(assetLinkCount)
			}
			processEntityClean("AssetsLinks", configBlockSize)
		} else {
			espLogger("There are no asset links to delete.", "debug")
			color.Red("There are no asset links to delete.")
		}
	}
}

func parseFlags() {
	flag.StringVar(&configFileName, "file", "conf.json", "Name of Configuration File To Load")
	flag.IntVar(&configBlockSize, "BlockSize", 3, "Number of records to delete per block")
	flag.StringVar(&configAPIKey, "apikey", "", "API key to authenticate the session with")
	flag.StringVar(&configInstance, "instance", "", "ID of the instance (case sensitive)")
	flag.Parse()
}

//setMaxRecordsSetting - takes integer, updates maxResultsAllowed system setting to value
func setMaxRecordsSetting(newMaxResults int) bool {
	espXmlmc.OpenElement("option")
	espXmlmc.SetParam("key", "api.xmlmc.queryExec.maxResultsAllowed")
	espXmlmc.SetParam("value", strconv.Itoa(newMaxResults))
	espXmlmc.CloseElement("option")
	browse, err := espXmlmc.Invoke("admin", "sysOptionSet")
	if err != nil {
		color.Red("Call to sysOptionSet for api.xmlmc.queryExec.maxResultsAllowed data failed.")
		espLogger("Call to sysOptionSet for api.xmlmc.queryExec.maxResultsAllowed data failed.", "error")
		return false
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		color.Red("Unmarshalling of data for sysOptionGet for api.xmlmc.queryExec.maxResultsAllowed failed.")
		espLogger("Unmarshalling of data for sysOptionGet for api.xmlmc.queryExec.maxResultsAllowed failed.", "error")
		return false
	}
	if xmlRespon.MethodResult != "ok" {
		color.Red("sysOptionSet was unsuccessful: " + xmlRespon.State.ErrorRet)
		espLogger("sysOptionSet was unsuccessful: "+xmlRespon.State.ErrorRet, "error")
		return false
	}
	espLogger("Max Results system setting set to: "+strconv.Itoa(newMaxResults), "debug")
	return true
}

//processEntityClean - iterates through and processes record blocks of size defined in flag configBlockSize
func processEntityClean(entity string, chunkSize int) {
	if len(cleanerConf.RequestReferences) > 0 {

		//Split request slice in to chunks
		var divided [][]string
		for i := 0; i < len(cleanerConf.RequestReferences); i += chunkSize {
			batch := cleanerConf.RequestReferences[i:getLowerInt(i+chunkSize, len(cleanerConf.RequestReferences))]
			divided = append(divided, batch)
		}
		//range through slice, delete request chunks
		for _, block := range divided {
			//fmt.Println(block)
			deleteRecords(entity, block)
		}

	} else {
		exitLoop := false
		for exitLoop == false {
			AllRecordIDs := getRecordIDs(entity)
			if len(AllRecordIDs) == 0 {
				exitLoop = true
				continue
			}
			deleteRecords(entity, AllRecordIDs)
		}
	}

	return
}

//getLowerInt
func getLowerInt(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

//loadConfig - loads configuration file in to struct
func loadConfig() cleanerConfStruct {
	cwd, _ := os.Getwd()
	configurationFilePath := cwd + "/" + configFileName
	if _, fileCheckErr := os.Stat(configurationFilePath); os.IsNotExist(fileCheckErr) {
		log.Fatal(fileCheckErr)
	}
	file, fileError := os.Open(configurationFilePath)
	if fileError != nil {
		log.Fatal(fileError)
	}
	decoder := json.NewDecoder(file)
	conf := cleanerConfStruct{}

	err := decoder.Decode(&conf)
	if err != nil {
		color.Red("Error decoding configuration file!")
	}
	return conf
}

// espLogger -- Log to ESP
func espLogger(message string, severity string) {
	espXmlmc.SetParam("fileName", "Hornbill_Clean")
	espXmlmc.SetParam("group", "general")
	espXmlmc.SetParam("severity", severity)
	espXmlmc.SetParam("message", message)
	espXmlmc.Invoke("system", "logMessage")
}
