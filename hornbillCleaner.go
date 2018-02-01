package main

import (
	"encoding/base64"
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
)

const (
	toolVer           = "1.3.0"
	appServiceManager = "com.hornbill.servicemanager"
)

var (
	cleanerConf     cleanerConfStruct
	configFileName  string
	configBlockSize string
	maxResults      int
	resetResults    bool
	currentBlock    int
	totalBlocks     int
	//BlockSize - Size of request or asset blocks to delete
	BlockSize int
	espXmlmc  *apiLib.XmlmcInstStruct
)

type xmlmcResponse struct {
	MethodResult string       `xml:"status,attr"`
	Params       paramsStruct `xml:"params"`
	State        stateStruct  `xml:"state"`
}

type xmlmcTaskResponse struct {
	MethodResult string      `xml:"status,attr"`
	State        stateStruct `xml:"state"`
	TasksJSON    string      `xml:"params>tasks"`
	Counters     string      `xml:"params>counters"`
}

type stateStruct struct {
	Code     string `xml:"code"`
	ErrorRet string `xml:"error"`
}

type paramsStruct struct {
	SessionID    string   `xml:"sessionId"`
	RequestIDs   []string `xml:"rowData>row>h_pk_reference"`
	AssetIDs     []string `xml:"rowData>row>h_pk_asset_id"`
	AssetLinkIDs []string `xml:"rowData>row>h_pk_id"`
	TimerIDs     []string `xml:"rowData>row>h_pk_tid"`
	BPMID        string   `xml:"primaryEntityData>record>h_bpm_id"`
	RecordCount  int      `xml:"count"`
	MaxResults   int      `xml:"option>value"`
}

type taskStruct struct {
	TaskID    string `xml:"h_task_id"`
	TaskTitle string `xml:"h_title"`
}

type workflowStruct struct {
	WorkflowID string
}

type cleanerConfStruct struct {
	UserName           string
	Password           string
	URL                string
	CleanRequests      bool
	RequestServices    []int
	RequestStatuses    []string
	RequestTypes       []string
	RequestLogDateFrom string
	RequestLogDateTo   string
	RequestReferences  []string
	CleanAssets        bool
	CleanUsers         bool
	Users              []string
}

func main() {

	flag.StringVar(&configFileName, "file", "conf.json", "Name of Configuration File To Load")
	flag.StringVar(&configBlockSize, "BlockSize", "3", "Number of records to delete per block")
	flag.Parse()

	BlockSize, err := strconv.Atoi(configBlockSize)
	fmt.Println(BlockSize)
	if err != nil {
		color.Red("Unable to convert block size of " + configBlockSize + " to type INT for processing")
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
	color.Red(" Hornbill instance: " + cleanerConf.URL)
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
		requestCount := 0
		if len(cleanerConf.RequestReferences) > 0 {
			currentBlock = 1
			requestCount = len(cleanerConf.RequestReferences)
			espLogger("Block Size: "+strconv.Itoa(BlockSize), "debug")
			requestBlocks := float64(requestCount) / float64(BlockSize)
			totalBlocks = int(math.Ceil(requestBlocks))

			espLogger("Number of Requests to delete: "+strconv.Itoa(requestCount), "debug")
			color.Green("Number of Requests to delete: " + strconv.Itoa(requestCount))
			if maxResults > 0 && requestCount > maxResults {
				resetResults = true
				//Update maxResultsAllowed system setting to match number of records to be deleted
				setMaxRecordsSetting(requestCount)
			}
			processEntityClean("Requests", BlockSize)
		} else {
			requestCount = getRecordCount("h_itsm_requests")
			if requestCount > 0 {
				currentBlock = 1
				espLogger("Block Size: "+strconv.Itoa(BlockSize), "debug")
				requestBlocks := float64(requestCount) / float64(BlockSize)
				totalBlocks = int(math.Ceil(requestBlocks))
				espLogger("Request Blocks: "+strconv.Itoa(int(requestBlocks)), "debug")
				espLogger("Total Blocks: "+strconv.Itoa(totalBlocks), "debug")

				espLogger("Number of Requests to delete: "+strconv.Itoa(requestCount), "debug")
				if maxResults > 0 && requestCount > maxResults {
					resetResults = true
					//Update maxResultsAllowed system setting to match number of records to be deleted
					setMaxRecordsSetting(requestCount)
				}
				processEntityClean("Requests", BlockSize)
			} else {
				espLogger("There are no requests to delete.", "debug")
				color.Red("There are no requests to delete.")
			}
		}
	}

	//Process Asset Records
	if cleanerConf.CleanAssets {
		assetCount := getRecordCount("h_cmdb_assets")
		if assetCount > 0 {
			currentBlock = 1
			assetBlocks := float64(assetCount) / float64(BlockSize)
			totalBlocks = int(math.Ceil(assetBlocks))
			espLogger("Number of Assets to delete: "+strconv.Itoa(assetCount), "debug")
			if maxResults > 0 && assetCount > maxResults {
				resetResults = true
				//Update maxResultsAllowed system setting to match number of records to be deleted
				setMaxRecordsSetting(assetCount)
			}
			processEntityClean("Asset", BlockSize)
		} else {
			espLogger("There are no assets to delete.", "debug")
			color.Red("There are no assets to delete.")
		}

		assetLinkCount := getRecordCount("h_cmdb_links")
		if assetLinkCount > 0 {
			currentBlock = 1
			assetLinkBlocks := float64(assetLinkCount) / float64(BlockSize)
			totalBlocks = int(math.Ceil(assetLinkBlocks))
			espLogger("Number of Asset Links to delete: "+strconv.Itoa(assetLinkCount), "debug")
			if maxResults > 0 && assetLinkCount > maxResults {
				resetResults = true
				//Update maxResultsAllowed system setting to match number of records to be deleted
				setMaxRecordsSetting(assetLinkCount)
			}
			processEntityClean("AssetsLinks", BlockSize)
		} else {
			espLogger("There are no asset links to delete.", "debug")
			color.Red("There are no asset links to delete.")
		}
	}

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

//getRecordCount - takes a table name, returns the total number of records in the entity
func getRecordCount(table string) int {
	strQuery := ""
	if table == "h_itsm_requests" {
		if len(cleanerConf.RequestTypes) > 0 {
			strQuery += " h_requesttype IN ("
			firstElement := true
			for _, reqType := range cleanerConf.RequestTypes {
				if firstElement == false {
					strQuery += ","
				}
				strQuery += "'" + reqType + "'"
				firstElement = false
			}
			strQuery += ")"
		}

		if len(cleanerConf.RequestStatuses) > 0 {
			if strQuery != "" {
				strQuery += " AND"
			}
			strQuery += " h_status IN ("
			firstElement := true
			for _, reqStatus := range cleanerConf.RequestStatuses {
				if firstElement == false {
					strQuery += ","
				}
				strQuery += "'" + reqStatus + "'"
				firstElement = false
			}
			strQuery += ")"
		}

		if len(cleanerConf.RequestServices) > 0 {
			if strQuery != "" {
				strQuery += " AND"
			}
			strQuery += " h_fk_serviceid IN ("
			firstElement := true
			for _, reqService := range cleanerConf.RequestServices {
				if firstElement == false {
					strQuery += ","
				}
				strQuery += strconv.Itoa(reqService)
				firstElement = false
			}
			strQuery += ")"
		}

		if cleanerConf.RequestLogDateFrom != "" {
			if strQuery != "" {
				strQuery += " AND"
			}
			strQuery += " h_datelogged >= '" + cleanerConf.RequestLogDateFrom + "'"
		}
		if cleanerConf.RequestLogDateTo != "" {
			if strQuery != "" {
				strQuery += " AND"
			}
			strQuery += " h_datelogged <= '" + cleanerConf.RequestLogDateTo + "'"
		}
	}

	espXmlmc.SetParam("database", "swdata")
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("table", table)
	if strQuery != "" {
		espXmlmc.SetParam("where", strQuery)
	}

	browse, err := espXmlmc.Invoke("data", "getRecordCount")
	if err != nil {
		color.Red("Get Record Count API Invoke failed for table: [" + table + "]")
		espLogger("Get Record Count API Invoke failed for table: ["+table+"]", "error")
		return 0
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		color.Red("Get Record Count data unmarshalling failed for table: [" + table + "]")
		espLogger("Get Record Count data unmarshalling failed for table: ["+table+"]", "error")
		return 0
	}
	if xmlRespon.MethodResult != "ok" {
		color.Red("sqlQuery was unsuccessful for table [" + table + "]: " + xmlRespon.State.ErrorRet)
		espLogger("Count sqlQuery was unsuccessful for table ["+table+"]: "+xmlRespon.State.ErrorRet, "error")
		return 0
	}
	return xmlRespon.Params.RecordCount
}

//getMaxRecordsSetting - gets and returns current maxResultsAllowed sys setting value
func getMaxRecordsSetting() int {
	espXmlmc.SetParam("filter", "api.xmlmc.queryExec.maxResultsAllowed")
	browse, err := espXmlmc.Invoke("admin", "sysOptionGet")
	if err != nil {
		color.Red("Call to sysOptionGet for api.xmlmc.queryExec.maxResultsAllowed failed.")
		espLogger("Call to sysOptionGet for api.xmlmc.queryExec.maxResultsAllowed failed.", "error")
		return 0
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		color.Red("Unmarshalling of data for sysOptionGet for api.xmlmc.queryExec.maxResultsAllowed failed.")
		espLogger("Unmarshalling of data for sysOptionGet for api.xmlmc.queryExec.maxResultsAllowed failed.", "error")
		return 0
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("sysOptionGet was unsuccessful: "+xmlRespon.State.ErrorRet, "error")
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

//getLowerInt
func getLowerInt(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

//processEntityClean - iterates through and processes record blocks of size defined in flag BlockSize
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

//getRecordIDs - returns an array of records for deletion
func getRecordIDs(entity string) []string {
	if currentBlock <= totalBlocks {
		fmt.Println("Returning block " + strconv.Itoa(currentBlock) + " of " + strconv.Itoa(totalBlocks) + " blocks of records from " + entity + " entity...")
	} else {
		color.Green("All " + entity + " records processed.")
	}
	if entity == "Requests" {
		//Use a stored query to get request IDs
		espXmlmc.SetParam("application", "com.hornbill.servicemanager")
		espXmlmc.SetParam("queryName", "_listRequestsOfType")
		espXmlmc.OpenElement("queryParams")
		for _, reqType := range cleanerConf.RequestTypes {
			espXmlmc.SetParam("type", reqType)
		}
		espXmlmc.SetParam("limit", configBlockSize)
		for _, reqStatus := range cleanerConf.RequestStatuses {
			espXmlmc.SetParam("status", reqStatus)
		}
		for _, reqService := range cleanerConf.RequestServices {
			espXmlmc.SetParam("service", strconv.Itoa(reqService))
		}
		if cleanerConf.RequestLogDateFrom != "" {
			espXmlmc.SetParam("fromDateTime", cleanerConf.RequestLogDateFrom)
		}
		if cleanerConf.RequestLogDateTo != "" {
			espXmlmc.SetParam("toDateTime", cleanerConf.RequestLogDateTo)
		}
		espXmlmc.CloseElement("queryParams")

		browse, err := espXmlmc.Invoke("data", "queryExec")
		if err != nil {
			espLogger("Call to queryExec ["+entity+"] failed when returning block "+strconv.Itoa(currentBlock), "error")
			color.Red("Call to queryExec [" + entity + "] failed when returning block " + strconv.Itoa(currentBlock))
			return nil
		}
		var xmlRespon xmlmcResponse
		err = xml.Unmarshal([]byte(browse), &xmlRespon)
		if err != nil {
			espLogger("Unmarshal of queryExec ["+entity+"] data failed when returning block "+strconv.Itoa(currentBlock), "error")
			color.Red("Unmarshal of queryExec [" + entity + "] data failed when returning block " + strconv.Itoa(currentBlock))
			return nil
		}
		if xmlRespon.MethodResult != "ok" {
			espLogger("queryExec was unsuccessful: "+xmlRespon.State.ErrorRet, "error")
			color.Red("queryExec was unsuccessful: " + xmlRespon.State.ErrorRet)
			return nil
		}
		return xmlRespon.Params.RequestIDs
	}

	if entity == "Asset" {
		//Use entityBrowseRecords to get asset entity records
		espXmlmc.SetParam("application", "com.hornbill.servicemanager")
		espXmlmc.SetParam("entity", entity)
		espXmlmc.SetParam("maxResults", configBlockSize)
		browse, err := espXmlmc.Invoke("data", "entityBrowseRecords")
		if err != nil {
			espLogger("Call to entityBrowseRecords ["+entity+"] failed when returning block "+strconv.Itoa(currentBlock), "error")
			color.Red("Call to entityBrowseRecords [" + entity + "] failed when returning block " + strconv.Itoa(currentBlock))
			return nil
		}
		var xmlRespon xmlmcResponse
		err = xml.Unmarshal([]byte(browse), &xmlRespon)
		if err != nil {
			espLogger("Unmarshal of entityBrowseRecords ["+entity+"] data failed when returning block "+strconv.Itoa(currentBlock), "error")
			color.Red("Unmarshal of entityBrowseRecords [" + entity + "] data failed when returning block " + strconv.Itoa(currentBlock))
			return nil
		}
		if xmlRespon.MethodResult != "ok" {
			espLogger("entityBrowseRecords was unsuccessful: "+xmlRespon.State.ErrorRet, "error")
			color.Red("entityBrowseRecords was unsuccessful: " + xmlRespon.State.ErrorRet)
			return nil
		}
		return xmlRespon.Params.AssetIDs
	}

	//Use entityBrowseRecords to get assetslinks entity records
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("entity", entity)
	espXmlmc.SetParam("maxResults", configBlockSize)
	browse, err := espXmlmc.Invoke("data", "entityBrowseRecords")
	if err != nil {
		espLogger("Call to entityBrowseRecords ["+entity+"] failed when returning block "+strconv.Itoa(currentBlock), "error")
		color.Red("Call to entityBrowseRecords [" + entity + "] failed when returning block " + strconv.Itoa(currentBlock))
		return nil
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Unmarshal of entityBrowseRecords ["+entity+"] data failed when returning block "+strconv.Itoa(currentBlock), "error")
		color.Red("Unmarshal of entityBrowseRecords [" + entity + "] data failed when returning block " + strconv.Itoa(currentBlock))
		return nil
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("entityBrowseRecords was unsuccessful: "+xmlRespon.State.ErrorRet, "error")
		color.Red("entityBrowseRecords was unsuccessful: " + xmlRespon.State.ErrorRet)
		return nil
	}
	return xmlRespon.Params.AssetLinkIDs
}

//deleteRecords - deletes the records in the array generated by getRecordIDs
func deleteRecords(entity string, records []string) {
	espLogger("Deleting block "+strconv.Itoa(currentBlock)+" of "+strconv.Itoa(totalBlocks)+" blocks of records from "+entity+" entity. Please wait...", "debug")
	fmt.Println("Deleting block " + strconv.Itoa(currentBlock) + " of " + strconv.Itoa(totalBlocks) + " blocks of records from " + entity + " entity. Please wait...")

	if entity == "Requests" {
		//Go through requests, and delete any associated records
		for _, callRef := range records {

			//-- System Timers
			sysTimerIDs := getSystemTimerIDs(callRef)
			if len(sysTimerIDs) != 0 {
				for _, timerID := range sysTimerIDs {
					if timerID != "<nil>" && timerID != "" {
						deleteTimer(timerID)
					}
				}
			}
			//-- Spawned workflow
			requestWorkflow := getRequestWorkflow(callRef)
			if requestWorkflow != "<nil>" && requestWorkflow != "" {
				deleteWorkflow(requestWorkflow)
			}

			//-- Request Tasks
			requestTasks := getRequestTasks(callRef)
			for _, stateMap := range requestTasks {
				for _, taskMap := range stateMap {
					deleteTask(taskMap.TaskID)
				}
			}

			//-- Asset Associations
			requestAssets := getRequestAssetLinks(callRef)
			for _, linkID := range requestAssets {
				if linkID != "<nil>" && linkID != "" {
					deleteAssetLink(linkID)
				}
			}

		}
	}

	//Now delete the block of records
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("entity", entity)
	for _, v := range records {
		espXmlmc.SetParam("keyValue", v)
	}
	deleted, err := espXmlmc.Invoke("data", "entityDeleteRecord")
	if err != nil {
		espLogger("Delete Records failed for entity ["+entity+"], block "+strconv.Itoa(currentBlock), "error")
		color.Red("Delete Records failed for entity [" + entity + "], block " + strconv.Itoa(currentBlock))
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(deleted), &xmlRespon)
	if err != nil {
		espLogger("Delete Records response unmarshall failed for entity ["+entity+"], block "+strconv.Itoa(currentBlock), "error")
		color.Red("Delete Records response unmarshall failed for entity [" + entity + "], block " + strconv.Itoa(currentBlock))
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("entityDeleteRecords was unsuccessful for entity ["+entity+"]: "+xmlRespon.State.ErrorRet, "error")
		color.Red("Could not delete records from " + entity + " entity: " + xmlRespon.State.ErrorRet)
		return
	}
	color.Green("Block " + strconv.Itoa(currentBlock) + " of " + strconv.Itoa(totalBlocks) + " deleted.")
	currentBlock++
	return
}

func deleteUser(strUser string) {
	//Now delete the block of records
	espXmlmc.SetParam("userId", strUser)

	deleted, err := espXmlmc.Invoke("admin", "userDelete")
	if err != nil {
		espLogger("Delete Records failed for user ["+strUser+"]", "error")
		color.Red("Delete Records failed for user [" + strUser + "]")
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(deleted), &xmlRespon)
	if err != nil {
		espLogger("Delete Records response unmarshall failed for user ["+strUser+"]", "error")
		color.Red("Delete Records response unmarshall failed for user [" + strUser + "]")
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("entityDeleteRecords was unsuccessful for user ["+strUser+"]: "+xmlRespon.State.ErrorRet, "error")
		color.Red("Could not delete user [" + strUser + "]: " + xmlRespon.State.ErrorRet)
		return
	}
	color.Green("User [" + strUser + "] deleted.")
	return
}

//getRequestTasks - take a call reference, get all associated request tasks
func getRequestTasks(callRef string) map[string][]taskStruct {
	//First get request task counters so we can set correct state
	espXmlmc.SetParam("objectRefUrn", "urn:sys:entity:"+appServiceManager+":Requests:"+callRef)
	espXmlmc.SetParam("counters", "true")
	getCounters, err := espXmlmc.Invoke("apps/com.hornbill.core/Task", "getEntityTasks")
	if err != nil {
		espLogger("Call to [getEntityTasks] failed for Request "+callRef+" : "+fmt.Sprintf("%s", err), "error")
		color.Red("Call to [getEntityTasks] failed for Request " + callRef)
		return nil
	}
	var xmlResponCount xmlmcTaskResponse
	err = xml.Unmarshal([]byte(getCounters), &xmlResponCount)
	if err != nil {
		espLogger("Unmarshal of [getEntityTasks] data failed for Request "+callRef+" : "+fmt.Sprintf("%s", err), "error")
		color.Red("Unmarshal of [getEntityTasks] data failed for Request " + callRef)
		return nil
	}
	if xmlResponCount.MethodResult != "ok" {
		espLogger("[getEntityTasks] was unsuccessful for Request "+callRef+": "+xmlResponCount.State.ErrorRet, "error")
		color.Red("[getEntityTasks] was unsuccessful for Request " + callRef + ": " + xmlResponCount.State.ErrorRet)
		return nil
	}

	objCounter := make(map[string]interface{})
	json.Unmarshal([]byte(xmlResponCount.Counters), &objCounter)

	if len(objCounter) > 0 {
		//--  get task IDs
		espXmlmc.SetParam("objectRefUrn", "urn:sys:entity:"+appServiceManager+":Requests:"+callRef)

		for k := range objCounter {
			espXmlmc.SetParam("taskStatus", fmt.Sprintf("%s", k))
		}

		browse, errTask := espXmlmc.Invoke("apps/"+appServiceManager+"/Task", "getEntityTasks")
		if errTask != nil {
			espLogger("Call to [getEntityTasks] failed for Request "+callRef+" : "+fmt.Sprintf("%s", errTask), "error")
			color.Red("Call to [getEntityTasks] failed for Request " + callRef)
			return nil
		}
		var xmlRespon xmlmcTaskResponse
		errTask = xml.Unmarshal([]byte(browse), &xmlRespon)
		if errTask != nil {
			espLogger("Unmarshal of [getEntityTasks] data failed for Request "+callRef+" : "+fmt.Sprintf("%s", errTask), "error")
			color.Red("Unmarshal of [getEntityTasks] data failed for Request " + callRef)
			return nil
		}
		if xmlRespon.MethodResult != "ok" {
			espLogger("[getEntityTasks] was unsuccessful for Request "+callRef+": "+xmlRespon.State.ErrorRet, "error")
			color.Red("[getEntityTasks] was unsuccessful for Request " + callRef + ": " + xmlRespon.State.ErrorRet)
			return nil
		}
		//Unmarshall JSON string in to map containing taskStruct slices
		objTasks := make(map[string][]taskStruct)
		json.Unmarshal([]byte(xmlRespon.TasksJSON), &objTasks)
		return objTasks
	}
	return nil
}

func getRequestAssetLinks(callref string) []string {
	//Use entityBrowseRecords to get asset entity records
	callrefURN := "urn:sys:entity:com.hornbill.servicemanager:Requests:" + callref
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("entity", "AssetsLinks")
	espXmlmc.SetParam("matchScope", "any")
	espXmlmc.OpenElement("searchFilter")
	espXmlmc.SetParam("column", "h_fk_id_l")
	espXmlmc.SetParam("value", callrefURN)
	espXmlmc.SetParam("matchType", "exact")
	espXmlmc.CloseElement("searchFilter")
	espXmlmc.OpenElement("searchFilter")
	espXmlmc.SetParam("column", "h_fk_id_r")
	espXmlmc.SetParam("value", callrefURN)
	espXmlmc.SetParam("matchType", "exact")
	espXmlmc.CloseElement("searchFilter")
	browse, err := espXmlmc.Invoke("data", "entityBrowseRecords2")
	if err != nil {
		espLogger("Call to entityBrowseRecords2 failed when returning asset links for "+callref, "error")
		color.Red("Call to entityBrowseRecords2 failed when returning asset links for " + callref)
		return nil
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Unmarshal of entityBrowseRecords2 failed when returning asset links for "+callref, "error")
		color.Red("Unmarshal of entityBrowseRecords2 failed when returning asset links for " + callref)
		return nil
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("entityBrowseRecords2 was unsuccessful: "+xmlRespon.State.ErrorRet, "error")
		color.Red("entityBrowseRecords2 was unsuccessful: " + xmlRespon.State.ErrorRet)
		return nil
	}
	return xmlRespon.Params.AssetLinkIDs
}

//getRequestWorkflow - take a call reference, get all associated rBPM workflow ID
func getRequestWorkflow(callRef string) string {
	returnWorkflowID := ""
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("entity", "Requests")
	espXmlmc.SetParam("keyValue", callRef)
	browse, err := espXmlmc.Invoke("data", "entityGetRecord")
	if err != nil {
		espLogger("Call to entityGetRecord failed when attepmting to return request ["+callRef+"]: "+fmt.Sprintf("%s", err), "error")
		color.Red("Call to entityGetRecord failed when attepmting to return request [" + callRef + "]")
		return ""
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Call to entityGetRecord failed when attepmting to return request ["+callRef+"]: "+fmt.Sprintf("%s", err), "error")
		color.Red("Call to entityGetRecord failed when attepmting to return request [" + callRef + "]")
		return ""
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("Call to entityGetRecord failed when attepmting to return request ["+callRef+"]: "+fmt.Sprintf("%s", err), "error")
		color.Red("Call to entityGetRecord failed when attepmting to return request [" + callRef + "]")
		return ""
	}
	returnWorkflowID = xmlRespon.Params.BPMID
	return returnWorkflowID
}

//getSystemTimerIDs - take call reference, return array of System Timers that are associated with it
func getSystemTimerIDs(callRef string) []string {
	//Use a stored query to get request IDs
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("queryName", "getRequestSystemTimers")
	espXmlmc.OpenElement("queryParams")
	espXmlmc.SetParam("requestId", callRef)
	espXmlmc.CloseElement("queryParams")
	browse, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		espLogger("Call to queryExec [getRequestSystemTimers] failed for Request "+callRef+" : "+fmt.Sprintf("%s", err), "error")
		color.Red("Call to queryExec [getRequestSystemTimers] failed for Request " + callRef)
		return nil
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Unmarshal of queryExec [getRequestSystemTimers] data failed for Request "+callRef+" : "+fmt.Sprintf("%s", err), "error")
		color.Red("Unmarshal of queryExec [getRequestSystemTimers] data failed for Request " + callRef)
		return nil
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("queryExec [getRequestSystemTimers] was unsuccessful: "+xmlRespon.State.ErrorRet, "error")
		color.Red("queryExec [getRequestSystemTimers] was unsuccessful: " + xmlRespon.State.ErrorRet)
		return nil
	}
	return xmlRespon.Params.TimerIDs
}

//deleteTimer - Takes a System Timer ID, sends it to time::timerDelete API for safe deletion
func deleteTimer(timerID string) {
	espXmlmc.SetParam("timerId", timerID)
	browse, err := espXmlmc.Invoke("time", "timerDelete")
	if err != nil {
		espLogger("Deletion of System Timer failed ["+timerID+"]", "error")
		color.Red("Deletion of System Timer failed [" + timerID + "]")
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Unmarshal of response to deletion of System Timer failed ["+timerID+"]", "error")
		color.Red("Unmarshal of response to deletion of System Timer failed [" + timerID + "]")
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("API Call to delete System Timer failed ["+timerID+"] "+xmlRespon.State.ErrorRet, "error")
		color.Red("API Call to delete System Timer failed [" + timerID + "] " + xmlRespon.State.ErrorRet)
		return
	}
	espLogger("Timer instance sucessfully deleted ["+timerID+"] ", "debug")
}

//deleteTask - Takes a Task ID, sends it to task::taskDelete API for safe deletion
func deleteTask(taskID string) {
	espXmlmc.SetParam("taskId", taskID)
	browse, err := espXmlmc.Invoke("task", "taskDelete")
	if err != nil {
		espLogger("Deletion of Task failed ["+taskID+"]", "error")
		color.Red("Deletion of Task failed [" + taskID + "]")
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Unmarshal of response to deletion of Task failed ["+taskID+"]", "error")
		color.Red("Unmarshal of response to deletion of Task failed [" + taskID + "]")
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("API Call to delete Task failed ["+taskID+"] "+xmlRespon.State.ErrorRet, "error")
		color.Red("API Call to delete Task failed [" + taskID + "] " + xmlRespon.State.ErrorRet)
		return
	}
	espLogger("Task instance sucessfully deleted ["+taskID+"] ", "debug")
}

//deleteWorkflow - Takes a Workflow ID, sends it to bpm::processDelete API for safe deletion
func deleteWorkflow(workflowID string) {
	espXmlmc.SetParam("identifier", workflowID)
	browse, err := espXmlmc.Invoke("bpm", "processDelete")
	if err != nil {
		espLogger("Deletion of Workflow failed ["+workflowID+"]", "error")
		color.Red("Deletion of Workflow failed [" + workflowID + "]")
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Unmarshal of response to deletion of Workflow failed ["+workflowID+"]", "error")
		color.Red("Unmarshal of response to deletion of Workflow failed [" + workflowID + "]")
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("API Call to delete Workflow failed ["+workflowID+"] "+xmlRespon.State.ErrorRet, "error")
		color.Red("API Call to delete Workflow failed [" + workflowID + "] " + xmlRespon.State.ErrorRet)
		return
	}
	espLogger("Workflow instance sucessfully deleted ["+workflowID+"] ", "debug")
}

//deleteAssetLink - Takes a Link PK ID, sends it to data::entityDelete API for safe deletion
func deleteAssetLink(linkID string) {
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("entity", "AssetsLinks")
	espXmlmc.SetParam("keyValue", linkID)
	espXmlmc.SetParam("preserveOneToOneData", "false")
	espXmlmc.SetParam("preserveOneToManyData", "false")
	browse, err := espXmlmc.Invoke("data", "entityDeleteRecord")
	if err != nil {
		espLogger("Deletion of Asset Link failed ["+linkID+"]", "error")
		color.Red("Deletion of Asset Link failed [" + linkID + "]")
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Unmarshal of response to deletion of Asset Link failed ["+linkID+"]", "error")
		color.Red("Unmarshal of response to deletion of Asset Link failed [" + linkID + "]")
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("API Call to delete Asset Link failed ["+linkID+"] "+xmlRespon.State.ErrorRet, "error")
		color.Red("API Call to delete Asset Link failed [" + linkID + "] " + xmlRespon.State.ErrorRet)
		return
	}
	espLogger("Asset Link sucessfully deleted ["+linkID+"] ", "debug")
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

//login - Starts a new ESP session
func login() bool {

	espXmlmc = apiLib.NewXmlmcInstance(cleanerConf.URL)
	espXmlmc.SetParam("userId", cleanerConf.UserName)
	espXmlmc.SetParam("password", base64.StdEncoding.EncodeToString([]byte(cleanerConf.Password)))
	XMLLogin, err := espXmlmc.Invoke("session", "userLogon")
	if err != nil {
		color.Red("Error returned when attempting to run Login API call.")
		fmt.Println(err)
		return false
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(XMLLogin), &xmlRespon)
	if err != nil {
		color.Red("Error returned when attempting to unmarshal Login API call response.")
		fmt.Println(err)
		return false
	}
	if xmlRespon.MethodResult != "ok" {
		color.Red("Error returned when attempting to log in to your instance: " + xmlRespon.State.ErrorRet)
		fmt.Println(xmlRespon)
		return false
	}
	return true
}

//logout - Log out of ESP
func logout() {
	espXmlmc.Invoke("session", "userLogoff")
}
