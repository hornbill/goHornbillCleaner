package main

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/hornbill/color"
	"github.com/hornbill/goApiLib"
	"log"
	"math"
	"os"
	"strconv"
)

const (
	toolVer = "1.0.1"
)

var (
	espXmlmc        *apiLib.XmlmcInstStruct
	cleanerConf     cleanerConfStruct
	configFileName  string
	configBlockSize string
	maxResults      int
	resetResults    bool
	currentBlock    int
	totalBlocks     int
	blockSize       int
)

type xmlmcResponse struct {
	MethodResult string       `xml:"status,attr"`
	Params       paramsStruct `xml:"params"`
	State        stateStruct  `xml:"state"`
}
type stateStruct struct {
	Code     string `xml:"code"`
	ErrorRet string `xml:"error"`
}
type paramsStruct struct {
	SessionID   string   `xml:"sessionId"`
	RequestIDs  []string `xml:"rowData>row>h_pk_reference"`
	AssetIDs    []string `xml:"rowData>row>h_pk_asset_id"`
	RecordCount int      `xml:"rowData>row>cnt"`
	MaxResults  int      `xml:"option>value"`
}

type cleanerConfStruct struct {
	UserName      string
	Password      string
	URL           string
	CleanRequests bool
	CleanAssets   bool
}

func main() {

	flag.StringVar(&configFileName, "file", "conf.json", "Name of Configuration File To Load")
	flag.StringVar(&configBlockSize, "blocksize", "20", "Number of records to delete per block")
	flag.Parse()

	blockSize, err := strconv.Atoi(configBlockSize)
	if err != nil {
		color.Red("Unable to convert block size of " + configBlockSize + " to type INT for processing")
		return
	}

	//Load the configuration file
	cleanerConf = loadConfig()

	//Ask if we want to delete before continuing
	fmt.Println("")
	fmt.Println("===== Hornbill Cleaner Utility v" + toolVer + " =====")
	if !cleanerConf.CleanRequests && !cleanerConf.CleanAssets {
		color.Red("No entity data has been specified for cleansing in " + configFileName)
		return
	}
	fmt.Println("")
	color.Red(" WARNING!")
	color.Red(" This utility will delete all records from the following entities in")
	color.Red(" your Hornbill instance, as specified in your configuration file: ")
	fmt.Println("")
	if cleanerConf.CleanRequests {
		color.Magenta(" * All Requests (and related data)")
	}
	if cleanerConf.CleanAssets {
		color.Magenta(" * All Assets (and related data)")
	}
	fmt.Println("")
	fmt.Println("Are you sure you want to permanently delete these records? (yes/no):")
	if confirmResponse() != true {
		return
	}
	color.Red("Are you absolutely sure? Type in the word 'delete' to confirm...")
	if confirmDelete() != true {
		return
	}
	//Try to login to the server if not succesfully exit
	success := login()
	if success != true {
		log.Fatal("Could not login to your Hornbill instance.")
	}
	defer logout()

	maxResults = getMaxRecordsSetting()

	//Process Request Records
	espLogger("System Setting Max Results: "+strconv.Itoa(maxResults), "debug")
	if cleanerConf.CleanRequests {
		requestCount := getRecordCount("h_itsm_requests")
		if requestCount > 0 {
			currentBlock = 1
			espLogger("Block Size: "+strconv.Itoa(blockSize), "debug")
			requestBlocks := float64(requestCount) / float64(blockSize)
			totalBlocks = int(math.Ceil(requestBlocks))
			espLogger("Request Blocks: "+strconv.Itoa(int(requestBlocks)), "debug")
			espLogger("Total Blocks: "+strconv.Itoa(totalBlocks), "debug")

			espLogger("Number of Requests to delete: "+strconv.Itoa(requestCount), "debug")
			if maxResults > 0 && requestCount > maxResults {
				resetResults = true
				//Update maxResultsAllowed system setting to match number of records to be deleted
				setMaxRecordsSetting(requestCount)
			}
			processEntityClean("Requests")
		} else {
			espLogger("There are no requests to delete.", "debug")
			color.Red("There are no requests to delete.")
		}
	}

	//Process Asset Records
	if cleanerConf.CleanAssets {
		assetCount := getRecordCount("h_cmdb_assets")
		if assetCount > 0 {
			currentBlock = 1
			assetBlocks := float64(assetCount) / float64(blockSize)
			totalBlocks = int(math.Ceil(assetBlocks))
			espLogger("Number of Assets to delete: "+strconv.Itoa(assetCount), "debug")
			if maxResults > 0 && assetCount > maxResults {
				resetResults = true
				//Update maxResultsAllowed system setting to match number of records to be deleted
				setMaxRecordsSetting(assetCount)
			}
			processEntityClean("Asset")
		} else {
			espLogger("There are no assets to delete.", "debug")
			color.Red("There are no assets to delete.")
		}
	}

	//Reset maxResultsAllowed system setting back to what it was before the process ran
	if resetResults {
		setMaxRecordsSetting(maxResults)
	}
}

//getRecordCount - takes a table name, returns the total number of records in the entity
func getRecordCount(table string) int {
	espXmlmc.SetParam("database", "swdata")
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("query", "SELECT COUNT(*) AS cnt FROM "+table)
	browse, err := espXmlmc.Invoke("data", "sqlQuery")
	if err != nil {
		espLogger("Get Record Count API Invoke failed for table: ["+table+"]", "error")
		return 0
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Get Record Count data unmarshalling failed for table: ["+table+"]", "error")
		return 0
	}
	if xmlRespon.MethodResult != "ok" {
		color.Red("sqlQuery was unsuccessful")
		espLogger("Count sqlQuery was unsuccessful for table ["+table+"]: "+xmlRespon.MethodResult, "error")
		return 0
	}
	return xmlRespon.Params.RecordCount
}

//getMaxRecordsSetting - gets and returns current maxResultsAllowed sys setting value
func getMaxRecordsSetting() int {
	espXmlmc.SetParam("filter", "api.xmlmc.queryExec.maxResultsAllowed")
	browse, err := espXmlmc.Invoke("admin", "sysOptionGet")
	if err != nil {
		espLogger("Call to sysOptionGet for api.xmlmc.queryExec.maxResultsAllowed failed.", "error")
		return 0
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Unmarshalling of data for sysOptionGet for api.xmlmc.queryExec.maxResultsAllowed failed.", "error")
		return 0
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("sysOptionGet was unsuccessful: "+xmlRespon.MethodResult, "error")
		color.Red("sysOptionGet was unsuccessful in exiting")
		return 0
	}
	return xmlRespon.Params.MaxResults
}

//setMaxRecordsSetting - takes integer, updates maxResultsAllowed system setting to value
func setMaxRecordsSetting(newMaxResults int) bool {
	espXmlmc.OpenElement("option")
	espXmlmc.SetParam("key", "api.xmlmc.queryExec.maxResultsAllowed")
	espXmlmc.SetParam("value", strconv.Itoa(newMaxResults))
	espXmlmc.CloseElement("option")
	browse, err := espXmlmc.Invoke("admin", "sysOptionSet")
	if err != nil {
		espLogger("Call to sysOptionSet for api.xmlmc.queryExec.maxResultsAllowed data failed.", "error")
		return false
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Unmarshalling of data for sysOptionGet for api.xmlmc.queryExec.maxResultsAllowed failed.", "error")
		return false
	}
	if xmlRespon.MethodResult != "ok" {
		color.Red("sysOptionSet was unsuccessful")
		espLogger("sysOptionSet was unsuccessful: "+xmlRespon.MethodResult, "error")
		return false
	}
	espLogger("Max Results system setting set to: "+strconv.Itoa(newMaxResults), "debug")
	return true
}

//processEntityClean - iterates through and processes record blocks of size defined in flag blockSize
func processEntityClean(entity string) {
	exitLoop := false
	for exitLoop == false {
		AllRecordIDs := getRecordIDs(entity)
		if len(AllRecordIDs) == 0 {
			exitLoop = true
			continue
		}
		deleteRecords(entity, AllRecordIDs)
	}
	return
}

//deleteRecords - deletes the assets in the array generated by getRecordIDs
func deleteRecords(entity string, records []string) {
	espLogger("Deleting block "+strconv.Itoa(currentBlock)+" of "+strconv.Itoa(totalBlocks)+" blocks of records from "+entity+" entity. Please wait...", "debug")
	fmt.Println("Deleting block " + strconv.Itoa(currentBlock) + " of " + strconv.Itoa(totalBlocks) + " blocks of records from " + entity + " entity. Please wait...")
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("entity", entity)
	for _, v := range records {
		espXmlmc.SetParam("keyValue", v)
	}
	deleted, err := espXmlmc.Invoke("data", "entityDeleteRecord")
	if err != nil {
		espLogger("Delete Records failed for entity ["+entity+"], block " + strconv.Itoa(currentBlock), "error")
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(deleted), &xmlRespon)
	if err != nil {
		espLogger("Delete Records response unmarshall failes for entity ["+entity+"], block " + strconv.Itoa(currentBlock), "error")
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("entityDeleteRecords was unsuccessful: "+xmlRespon.MethodResult, "error")
		color.Red("Could not delete records from " + entity + " entity.")
		return
	}
	currentBlock++
	return
}

//getRecordIDs - returns an array of records for deletion
func getRecordIDs(entity string) []string {
	if currentBlock <= totalBlocks {
		fmt.Println("Returning block " + strconv.Itoa(currentBlock) + " of " + strconv.Itoa(totalBlocks) + " blocks of records from " + entity + " entity...")
	} else {
		color.Green("All " + entity + " records processed.")
	}

	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("entity", entity)
	espXmlmc.SetParam("maxResults", configBlockSize)
	browse, err := espXmlmc.Invoke("data", "entityBrowseRecords")
	if err != nil {
		espLogger("Call to entityBrowseRecords ["+entity+"] failed when returning block " + strconv.Itoa(currentBlock), "error")
		return nil
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Unmarshal of entityBrowseRecords ["+entity+"] data failed when returning block " + strconv.Itoa(currentBlock), "error")
		return nil
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("entityBrowseRecords was unsuccessful: "+xmlRespon.MethodResult, "error")
		color.Red("entityBrowseRecords was unsuccessful")
		return nil
	}
	if entity == "Requests" {
		return xmlRespon.Params.RequestIDs
	}
	return xmlRespon.Params.AssetIDs
}

//login - Starts a new ESP session
func login() bool {

	espXmlmc = apiLib.NewXmlmcInstance(cleanerConf.URL)
	espXmlmc.SetParam("userId", cleanerConf.UserName)
	espXmlmc.SetParam("password", base64.StdEncoding.EncodeToString([]byte(cleanerConf.Password)))
	XMLLogin, err := espXmlmc.Invoke("session", "userLogon")
	if err != nil {
		return false
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(XMLLogin), &xmlRespon)
	if err != nil {
		return false
	}
	if xmlRespon.MethodResult != "ok" {
		return false
	}
	return true
}

//logout - Log out of ESP
func logout() {
	espXmlmc.Invoke("session", "userLogoff")
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

//confirmResponse - prompts user, expects a fuzzy yes or no response, does not continue until this is given
func confirmResponse() bool {
	var cmdResponse string
	_, errResponse := fmt.Scanln(&cmdResponse)
	if errResponse != nil {
		log.Fatal(errResponse)
	}
	if cmdResponse == "y" || cmdResponse == "yes" || cmdResponse == "Y" || cmdResponse == "Yes" || cmdResponse == "YES" {
		return true
	} else if cmdResponse == "n" || cmdResponse == "no" || cmdResponse == "N" || cmdResponse == "No" || cmdResponse == "NO" {
		return false
	} else {
		color.Red("Please enter yes or no to continue:")
		return confirmResponse()
	}
}

//confirmDelete - prompts user, expects a delete or no response, does not continue until this is given
func confirmDelete() bool {
	var cmdResponse string
	_, errResponse := fmt.Scanln(&cmdResponse)
	if errResponse != nil {
		log.Fatal(errResponse)
	}
	if cmdResponse == "delete" {
		return true
	} else if cmdResponse == "n" || cmdResponse == "no" || cmdResponse == "N" || cmdResponse == "No" || cmdResponse == "NO" {
		return false
	} else {
		color.Red("Please enter delete or no to continue:")
		return confirmDelete()
	}
}

// espLogger -- Log to ESP
func espLogger(message string, severity string) {
	espXmlmc.SetParam("fileName", "Hornbill_Clean")
	espXmlmc.SetParam("group", "general")
	espXmlmc.SetParam("severity", severity)
	espXmlmc.SetParam("message", message)
	espXmlmc.Invoke("system", "logMessage")
}
