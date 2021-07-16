package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/fatih/color"
	apiLib "github.com/hornbill/goApiLib"
	hornbillHelpers "github.com/hornbill/goHornbillHelpers"
	"github.com/tcnksm/go-latest"
)

var (
	logTimeNow string
)

func main() {
	parseFlags()

	//-- If configVersion just output version number and die
	if configVersion {
		fmt.Printf("%v \n", version)
		return
	}
	logTimeNow = time.Now().Format("20060102150405.999999")

	//Does endpoint exist?
	instanceEndpoint := apiLib.GetEndPointFromName(configInstance)
	if instanceEndpoint == "" {
		color.Red("The provided instance ID [" + configInstance + "] could not be found.")
		return
	}

	//Create new session
	espXmlmc = apiLib.NewXmlmcInstance(configInstance)
	espXmlmc.SetAPIKey(configAPIKey)

	espLogger("********** Cleaner Utility Started **********", "notice")
	defer espLogger("********** Cleaner Utility Completed **********", "notice")

	checkVersion()

	var err error
	//Load the configuration file
	cleanerConf, err = loadConfig()
	if err != nil {
		color.Red("Error decoding configuration file " + configFileName + " : " + err.Error())
		espLogger("Error decoding configuration file "+configFileName+" : "+err.Error(), "error")
		return
	}

	//Ask if we want to delete before continuing
	fmt.Println("")
	fmt.Println("===== Hornbill Cleaner Utility v" + version + " =====")
	if !cleanerConf.CleanRequests && !cleanerConf.CleanAssets && !cleanerConf.CleanUsers && !cleanerConf.CleanServiceAvailabilityHistory && !cleanerConf.CleanContacts && !cleanerConf.CleanOrganisations && !cleanerConf.CleanSuppliers && !cleanerConf.CleanSupplierContracts {
		color.Red("No entity data has been specified for cleansing in " + configFileName)
		espLogger("No entity data has been specified for cleansing in "+configFileName, "error")
		return
	}
	if cleanerConf.CleanServiceAvailabilityHistory && len(cleanerConf.ServiceAvailabilityServiceIDs) == 0 {
		color.Red("CleanServiceAvailabilityHistory is set to true but no ServiceAvailabilityServiceIDs have been specified for cleaning in " + configFileName)
		espLogger("CleanServiceAvailabilityHistory is set to true but no ServiceAvailabilityServiceIDs have been specified for cleaning in "+configFileName, "error")
		return
	}
	fmt.Println("")
	if configDryRun {
		color.Green(" Hornbill instance: " + configInstance)
		color.Green(" DryRun Mode Active. No records will be deleted.")
		color.Green(" This utility will output to the log a list of all records that would have be deleted")
		color.Green(" as specified in your configuration file: ")
	} else {
		color.Red(" Hornbill instance: " + configInstance)
		color.Red(" WARNING!")
		color.Red(" This utility will attempt to delete records from the following entities")
		color.Red(" as specified in your configuration file: ")
	}

	fmt.Println("")
	if cleanerConf.CleanRequests {
		color.Magenta(" * Requests (and related data)")
	}
	if cleanerConf.CleanAssets {
		color.Magenta(" * Assets (and related data)")
	}
	if cleanerConf.CleanUsers {
		color.Magenta(" * Specified Users (and related data)")
	}
	if cleanerConf.CleanServiceAvailabilityHistory && len(cleanerConf.ServiceAvailabilityServiceIDs) > 0 {
		color.Magenta(" * Service Availability History for the configuration defined service IDs")
	}
	if cleanerConf.CleanContacts && len(cleanerConf.ContactIDs) > 0 {
		color.Magenta(" * Specified Contacts")
	}
	if cleanerConf.CleanOrganisations && len(cleanerConf.OrganisationIDs) > 0 {
		color.Magenta(" * Specified Organisations")
	}
	if cleanerConf.CleanSuppliers && len(cleanerConf.SupplierIDs) > 0 {
		color.Magenta(" * Specified Suppliers")
	}
	if cleanerConf.CleanSupplierContracts && len(cleanerConf.SupplierContractIDs) > 0 {
		color.Magenta(" * Specified Supplier Contracts")
	}
	fmt.Println("")

	if !configSkipPrompts {
		fmt.Println("Are you sure you want to permanently delete these records? (yes/no):")
		if !hornbillHelpers.ConfirmResponse("") {
			espLogger("Confirmation Prompts Rejected", "info")
			return
		}
		color.Red("Are you absolutely sure? Type in the word 'delete' to confirm...")
		if !hornbillHelpers.ConfirmResponse("delete") {
			espLogger("Confirmation Prompts Rejected", "info")
			return
		}
		espLogger("Confirmation Prompts Accepted", "info")
	} else {
		espLogger("Confirmation Prompts Skipped", "info")
	}

	//Log the config
	logConfig()
	processRequests()
	processAssets()
	processServiceAvailabilityHistory()
	processContacts()
	processOrgs()
	processSuppliers()
	processSupplierContracts()

	if cleanerConf.CleanUsers {
		espLogger("[USERS] Attempting to delete "+strconv.Itoa(len(cleanerConf.Users))+" Users", "info")

		for _, v := range cleanerConf.Users {
			espLogger("[USER] "+v, "info")
			if !configDryRun {
				deleteUser(v)
			}

		}
	}
}

func logConfig() {
	espLogger("Config File Name: "+configFileName, "info")
	espLogger("Dry Run: "+fmt.Sprintf("%t", configDryRun), "info")
	espLogger("Skip Prompts: "+fmt.Sprintf("%t", configSkipPrompts), "info")

	espLogger("CleanRequests: "+fmt.Sprintf("%t", cleanerConf.CleanRequests), "info")
	if cleanerConf.CleanRequests {
		noFilters := true
		if len(cleanerConf.RequestServices) > 0 {
			noFilters = false
			espLogger("Filtered by Service ID(s)", "info")
			for _, v := range cleanerConf.RequestServices {
				espLogger("Service ID: "+strconv.Itoa(v), "info")
			}
		}
		if len(cleanerConf.RequestStatuses) > 0 {
			noFilters = false
			espLogger("Filtered by Request Status(es)", "info")
			for _, v := range cleanerConf.RequestStatuses {
				espLogger("Request Status: "+v, "info")
			}
		}
		if len(cleanerConf.RequestTypes) > 0 {
			noFilters = false
			espLogger("Filtered by Request Type(s)", "info")
			for _, v := range cleanerConf.RequestTypes {
				espLogger("Request Type: "+v, "info")
			}
		}
		if len(cleanerConf.RequestReferences) > 0 {
			noFilters = false
			espLogger("Filtered by Request Reference(s)", "info")
			for _, v := range cleanerConf.RequestReferences {
				espLogger("Request Reference: "+v, "info")
			}
		}
		if cleanerConf.RequestLogDateFrom != "" {
			noFilters = false
			espLogger("Filtered by RequestLogDateFrom: "+cleanerConf.RequestLogDateFrom, "info")
		}
		if cleanerConf.RequestLogDateTo != "" {
			noFilters = false
			espLogger("Filtered by RequestLogDateTo: "+cleanerConf.RequestLogDateTo, "info")
		}
		if cleanerConf.RequestClosedDateFrom != "" {
			noFilters = false
			espLogger("Filtered by RequestCloseDateFrom: "+cleanerConf.RequestClosedDateFrom, "info")
		}
		if cleanerConf.RequestClosedDateTo != "" {
			noFilters = false
			espLogger("Filtered by RequestCloseDateTo: "+cleanerConf.RequestClosedDateTo, "info")
		}
		if noFilters {
			espLogger("No Request filters specified, all Requests and associated data will be deleted.", "warn")
		}
	}

	espLogger("CleanAssets: "+fmt.Sprintf("%t", cleanerConf.CleanAssets), "info")
	espLogger("CleanUsers: "+fmt.Sprintf("%t", cleanerConf.CleanUsers), "info")
	espLogger("CleanServiceAvailabilityHistory: "+fmt.Sprintf("%t", cleanerConf.CleanServiceAvailabilityHistory), "info")
	espLogger("CleanContacts: "+fmt.Sprintf("%t", cleanerConf.CleanContacts), "info")
	espLogger("CleanOrganisations: "+fmt.Sprintf("%t", cleanerConf.CleanOrganisations), "info")
	espLogger("CleanSuppliers: "+fmt.Sprintf("%t", cleanerConf.CleanSuppliers), "info")
	if cleanerConf.CleanUsers {
		for _, v := range cleanerConf.Users {
			espLogger("User ID: "+v, "info")
		}
	}
}

func isAppInstalled(appName string, buildVer int) bool {
	apps, success := getAppList()
	if success {
		for _, v := range apps {
			if v.Name == appName && v.Build >= buildVer {
				return true
			}
		}
	}
	return false
}

func processContacts() {
	//Process Contact Records
	if cleanerConf.CleanContacts {
		contactCount := 0
		if len(cleanerConf.ContactIDs) > 0 {
			currentBlock = 1
			contactCount = len(cleanerConf.ContactIDs)
			espLogger("Block Size: "+strconv.Itoa(configBlockSize), "debug")
			contactBlocks := float64(contactCount) / float64(configBlockSize)
			totalBlocks = int(math.Ceil(contactBlocks))
			espLogger("Number of Contacts to delete: "+strconv.Itoa(contactCount), "debug")
			color.Green("Number of Contacts to delete: " + strconv.Itoa(contactCount))
			processEntityClean("Contact", configBlockSize)
		} else {
			espLogger("There are no contacts to delete.", "debug")
			color.Red("There are no contacts to delete.")
		}
	}
}

func processOrgs() {
	//Process Org Records
	if cleanerConf.CleanOrganisations {
		orgCount := 0
		if len(cleanerConf.OrganisationIDs) > 0 {
			currentBlock = 1
			orgCount = len(cleanerConf.OrganisationIDs)
			espLogger("Block Size: "+strconv.Itoa(configBlockSize), "debug")
			orgBlocks := float64(orgCount) / float64(configBlockSize)
			totalBlocks = int(math.Ceil(orgBlocks))
			espLogger("Number of Organisations to delete: "+strconv.Itoa(orgCount), "debug")
			color.Green("Number of Organisations to delete: " + strconv.Itoa(orgCount))
			processEntityClean("Organizations", configBlockSize)
		} else {
			espLogger("There are no organisations to delete.", "debug")
			color.Red("There are no organisations to delete.")
		}
	}
}

func processSuppliers() {
	//Process Supplier Records
	if cleanerConf.CleanSuppliers {
		count := 0
		if len(cleanerConf.SupplierIDs) > 0 {
			currentBlock = 1
			count = len(cleanerConf.SupplierIDs)
			espLogger("Block Size: "+strconv.Itoa(configBlockSize), "debug")
			blocks := float64(count) / float64(configBlockSize)
			totalBlocks = int(math.Ceil(blocks))
			espLogger("Number of Suppliers to delete: "+strconv.Itoa(count), "debug")
			color.Green("Number of Suppliers to delete: " + strconv.Itoa(count))
			processEntityClean("Suppliers", configBlockSize)
		} else {
			espLogger("There are no Suppliers to delete.", "debug")
			color.Red("There are no Suppliers to delete.")
		}
	}
}

func processSupplierContracts() {
	//Process Supplier Contract Records
	if cleanerConf.CleanSupplierContracts {
		count := 0
		if len(cleanerConf.SupplierContractIDs) > 0 {
			currentBlock = 1
			count = len(cleanerConf.SupplierContractIDs)
			espLogger("Block Size: "+strconv.Itoa(configBlockSize), "debug")
			blocks := float64(count) / float64(configBlockSize)
			totalBlocks = int(math.Ceil(blocks))
			espLogger("Number of Supplier Contracts to delete: "+strconv.Itoa(count), "debug")
			color.Green("Number of Supplier Contracts to delete: " + strconv.Itoa(count))
			processEntityClean("SupplierContracts", configBlockSize)
		} else {
			espLogger("There are no Supplier Contracts to delete.", "debug")
			color.Red("There are no Supplier Contracts to delete.")
		}
	}
}

func processRequests() {
	//Process Request Records
	if cleanerConf.CleanRequests {
		//Is Board Manager installed
		boardManagerInstalled = isAppInstalled(appBM, minBoardManagerBuild)

		requestCount := 0
		if len(cleanerConf.RequestReferences) > 0 {
			currentBlock = 1
			requestCount = len(cleanerConf.RequestReferences)
			espLogger("Block Size: "+strconv.Itoa(configBlockSize), "debug")
			requestBlocks := float64(requestCount) / float64(configBlockSize)
			totalBlocks = int(math.Ceil(requestBlocks))

			if !configDryRun {
				espLogger("Number of Requests to delete: "+strconv.Itoa(requestCount), "debug")
				color.Green("Number of Requests to delete: " + strconv.Itoa(requestCount))
				processEntityClean("Requests", configBlockSize)
			}
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
		assetCount := getAssetCount()
		if assetCount > 0 {
			currentBlock = 1
			assetBlocks := float64(assetCount) / float64(configBlockSize)
			totalBlocks = int(math.Ceil(assetBlocks))
			espLogger("Number of Assets to delete: "+strconv.Itoa(assetCount), "debug")
			processEntityClean("Asset", configBlockSize)
		} else {
			espLogger("There are no assets to delete.", "debug")
			color.Red("There are no assets to delete.")
		}
		if !configDryRun {
			assetLinkCount := getRecordCount("h_cmdb_links")
			if assetLinkCount > 0 {
				currentBlock = 1
				assetLinkBlocks := float64(assetLinkCount) / float64(configBlockSize)
				totalBlocks = int(math.Ceil(assetLinkBlocks))
				espLogger("Number of Asset Links to delete: "+strconv.Itoa(assetLinkCount), "debug")
				processEntityClean("AssetsLinks", configBlockSize)
			} else {
				espLogger("There are no asset links to delete.", "debug")
				color.Red("There are no asset links to delete.")
			}
		}
	}
}

func processServiceAvailabilityHistory() {
	//Process Service Availability Records
	if cleanerConf.CleanServiceAvailabilityHistory {
		sahCount := getServiceAvailabilityHistoryCount()
		if sahCount > 0 {
			currentBlock = 1
			sahBlocks := float64(sahCount) / float64(configBlockSize)
			totalBlocks = int(math.Ceil(sahBlocks))
			espLogger("Number of ServiceStatusHistory records to delete: "+strconv.Itoa(sahCount), "debug")
			processEntityClean("ServiceStatusHistory", configBlockSize)
		} else {
			espLogger("There are no ServiceStatusHistory records to delete.", "debug")
			color.Red("There are no ServiceStatusHistory records to delete.")
		}
	}
}

func parseFlags() {
	flag.StringVar(&configFileName, "file", "conf.json", "Name of Configuration File To Load")
	flag.IntVar(&configBlockSize, "BlockSize", 3, "Number of records to delete per block")
	flag.StringVar(&configAPIKey, "apikey", "", "API key to authenticate the session with")
	flag.StringVar(&configInstance, "instance", "", "ID of the instance (case sensitive)")
	flag.BoolVar(&configDryRun, "dryrun", false, "DryRun mode - outputs record keys to log for review without deleting anything")
	flag.BoolVar(&configSkipPrompts, "justrun", false, "Set to true to run the cleanup without the prompts")
	flag.BoolVar(&configVersion, "version", false, "Returns the tool version")
	flag.Parse()
}

//processEntityClean - iterates through and processes record blocks of size defined in flag configBlockSize
func processEntityClean(entity string, chunkSize int) {
	if entity == "Requests" && len(cleanerConf.RequestReferences) > 0 {

		//Split request slice in to chunks
		var divided [][]string
		for i := 0; i < len(cleanerConf.RequestReferences); i += chunkSize {
			batch := cleanerConf.RequestReferences[i:getLowerInt(i+chunkSize, len(cleanerConf.RequestReferences))]
			divided = append(divided, batch)
		}
		//range through slice, delete request chunks
		for _, block := range divided {
			var requestDataToStruct []dataStruct
			for _, v := range block {
				requestDataToStruct = append(requestDataToStruct, dataStruct{RequestID: v})
			}
			deleteRecords(entity, requestDataToStruct)
		}

	} else if entity == "Contact" && len(cleanerConf.ContactIDs) > 0 {

		//Split request slice in to chunks
		var divided [][]int
		for i := 0; i < len(cleanerConf.ContactIDs); i += chunkSize {
			batch := cleanerConf.ContactIDs[i:getLowerInt(i+chunkSize, len(cleanerConf.ContactIDs))]
			divided = append(divided, batch)
		}
		//range through slice, delete request chunks
		for _, block := range divided {
			var contactDataToStruct []dataStruct
			for _, v := range block {
				contactDataToStruct = append(contactDataToStruct, dataStruct{ContactID: v})
			}
			deleteRecords(entity, contactDataToStruct)
		}

	} else if entity == "Organizations" && len(cleanerConf.OrganisationIDs) > 0 {

		//Split request slice in to chunks
		var divided [][]int
		for i := 0; i < len(cleanerConf.OrganisationIDs); i += chunkSize {
			batch := cleanerConf.OrganisationIDs[i:getLowerInt(i+chunkSize, len(cleanerConf.OrganisationIDs))]
			divided = append(divided, batch)
		}
		//range through slice, delete request chunks
		for _, block := range divided {
			var orgDataToStruct []dataStruct
			for _, v := range block {
				orgDataToStruct = append(orgDataToStruct, dataStruct{OrgID: v})
			}
			deleteRecords(entity, orgDataToStruct)
		}

	} else if entity == "Suppliers" && len(cleanerConf.SupplierIDs) > 0 {
		//Split request slice in to chunks
		var divided [][]int
		for i := 0; i < len(cleanerConf.SupplierIDs); i += chunkSize {
			batch := cleanerConf.SupplierIDs[i:getLowerInt(i+chunkSize, len(cleanerConf.SupplierIDs))]
			divided = append(divided, batch)
		}
		//range through slice, delete request chunks
		for _, block := range divided {
			var dataToStruct []dataStruct
			for _, v := range block {
				dataToStruct = append(dataToStruct, dataStruct{SuppID: v})
			}
			deleteRecords(entity, dataToStruct)
		}

	} else if entity == "SupplierContracts" && len(cleanerConf.SupplierContractIDs) > 0 {
		//Split request slice in to chunks
		var divided [][]string
		for i := 0; i < len(cleanerConf.SupplierContractIDs); i += chunkSize {
			batch := cleanerConf.SupplierContractIDs[i:getLowerInt(i+chunkSize, len(cleanerConf.SupplierContractIDs))]
			divided = append(divided, batch)
		}
		//range through slice, delete request chunks
		for _, block := range divided {
			var dataToStruct []dataStruct
			for _, v := range block {
				dataToStruct = append(dataToStruct, dataStruct{SuppConID: v})
			}
			deleteRecords(entity, dataToStruct)
		}

	} else {
		exitLoop := false
		for !exitLoop {
			AllRecordIDs := getRecordIDs(entity)
			if len(AllRecordIDs) == 0 {
				exitLoop = true
				continue
			}
			deleteRecords(entity, AllRecordIDs)
		}
	}
}

//getLowerInt
func getLowerInt(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

//loadConfig - loads configuration file in to struct
func loadConfig() (cleanerConfStruct, error) {
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
	return conf, err
}

// espLogger -- Log to ESP
func espLogger(message string, severity string) {
	if configDryRun {
		message = "[DRYRUN] " + message
	}
	espXmlmc.SetParam("fileName", "HornbillCleanerUtility")
	espXmlmc.SetParam("group", "general")
	espXmlmc.SetParam("severity", severity)
	espXmlmc.SetParam("message", message)
	espXmlmc.Invoke("system", "logMessage")

	logger(message)
}

// logger -- function to append to the current log file
func logger(s string) {
	cwd, _ := os.Getwd()
	logPath := cwd + "/log"
	logFileName := logPath + "/HornbillCleaner_" + logTimeNow + ".log"

	//-- If Folder Does Not Exist then create it
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		err := os.Mkdir(logPath, 0777)
		if err != nil {
			color.Red("Error Creating Log Folder %q: %s \r", logPath, err)
			os.Exit(101)
		}
	}

	//-- Open Log File
	f, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		color.Red("Error Creating Log File %q: %s \n", logFileName, err)
		os.Exit(100)
	}
	// don't forget to close it
	defer f.Close()

	// assign the file to the standard logger
	log.SetOutput(f)
	//Write the log entry
	log.Println(s)
}

//-- Check Latest
func checkVersion() {
	githubTag := &latest.GithubTag{
		Owner:      "hornbill",
		Repository: appName,
	}

	res, err := latest.Check(githubTag, version)
	if err != nil {
		msg := "Unable to check utility version against Github repository: " + err.Error()
		color.Red(msg)
		espLogger(msg, "error")
		return
	}
	if res.Outdated {
		msg := "v" + version + " is not latest, you should upgrade to " + res.Current + " by downloading the latest package from: https://github.com/hornbill/" + appName + "/releases/tag/v" + res.Current
		color.Yellow(msg)
		espLogger(msg, "warn")
	}
}
