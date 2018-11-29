package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/hornbill/color"
	"github.com/hornbill/goHornbillHelpers"
)

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

			logDateFrom := cleanerConf.RequestLogDateFrom
			boolIsDuration := durationRegex.MatchString(logDateFrom)
			if boolIsDuration {
				fromTime, _ := hornbillHelpers.CalculateTimeDuration(time.Now(), logDateFrom)
				logDateFrom = fromTime.UTC().Format(datetimeFormat)
			}
			strQuery += " h_datelogged >= '" + logDateFrom + "'"
		}
		if cleanerConf.RequestLogDateTo != "" {
			if strQuery != "" {
				strQuery += " AND"
			}
			logDateTo := cleanerConf.RequestLogDateTo
			boolIsDuration := durationRegex.MatchString(logDateTo)
			if boolIsDuration {
				toTime, _ := hornbillHelpers.CalculateTimeDuration(time.Now(), logDateTo)
				logDateTo = toTime.UTC().Format(datetimeFormat)
			}
			strQuery += " h_datelogged <= '" + logDateTo + "'"
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

//getRecordIDs - returns an array of records for deletion
func getRecordIDs(entity string) []dataStruct {
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
		if configDryRun {
			if currentBlock == 1 {
				espXmlmc.SetParam("rowstart", "0")
			} else {
				espXmlmc.SetParam("rowstart", strconv.Itoa((configBlockSize*currentBlock)-1))
			}
		}
		espXmlmc.SetParam("limit", strconv.Itoa(configBlockSize))
		for _, reqStatus := range cleanerConf.RequestStatuses {
			espXmlmc.SetParam("status", reqStatus)
		}
		for _, reqService := range cleanerConf.RequestServices {
			espXmlmc.SetParam("service", strconv.Itoa(reqService))
		}
		if cleanerConf.RequestLogDateFrom != "" {
			logDateFrom := cleanerConf.RequestLogDateFrom
			boolIsDuration := durationRegex.MatchString(logDateFrom)
			if boolIsDuration {
				fromTime, _ := hornbillHelpers.CalculateTimeDuration(time.Now(), logDateFrom)
				logDateFrom = fromTime.UTC().Format(datetimeFormat)
			}
			espXmlmc.SetParam("fromDateTime", logDateFrom)
		}
		if cleanerConf.RequestLogDateTo != "" {
			logDateTo := cleanerConf.RequestLogDateTo
			boolIsDuration := durationRegex.MatchString(logDateTo)
			if boolIsDuration {
				toTime, _ := hornbillHelpers.CalculateTimeDuration(time.Now(), logDateTo)
				logDateTo = toTime.UTC().Format(datetimeFormat)
			}
			espXmlmc.SetParam("toDateTime", logDateTo)
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
		return xmlRespon.Params.RecordIDs
	}

	if entity == "Asset" {
		//Use a stored query to get asset IDs
		espXmlmc.SetParam("application", "com.hornbill.servicemanager")
		espXmlmc.SetParam("queryName", "getAssetsList")
		espXmlmc.OpenElement("queryParams")
		if configDryRun {
			if currentBlock == 1 {
				espXmlmc.SetParam("rowstart", "0")
			} else {
				espXmlmc.SetParam("rowstart", strconv.Itoa((configBlockSize*currentBlock)-1))
			}
		}
		espXmlmc.SetParam("limit", strconv.Itoa(configBlockSize))
		espXmlmc.CloseElement("queryParams")
		espXmlmc.OpenElement("queryOptions")
		espXmlmc.SetParam("queryType", "records")
		espXmlmc.CloseElement("queryOptions")
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
		return xmlRespon.Params.RecordIDs
	}

	//Use entityBrowseRecords to get assetslinks entity records
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("entity", entity)
	espXmlmc.SetParam("maxResults", strconv.Itoa(configBlockSize))
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
	return xmlRespon.Params.RecordIDs
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

func getRequestAssetLinks(callref string) []dataStruct {
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
	return xmlRespon.Params.RecordIDs
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
func getSystemTimerIDs(callRef string) []dataStruct {
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
	return xmlRespon.Params.RecordIDs
}

func getAppList() ([]appsStruct, bool) {
	var returnArray []appsStruct
	returnBool := false
	apps, err := espXmlmc.Invoke("admin", "getApplicationList")
	if err != nil {
		espLogger("Call to admin::getApplicationList error: "+fmt.Sprintf("%v", err), "error")
		color.Red("Call to admin::getApplicationList error:", err)
		return returnArray, returnBool
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(apps), &xmlRespon)
	if err != nil {
		espLogger("Unmarshal of admin::getApplicationList response error: "+fmt.Sprintf("%v", err), "error")
		color.Red("Unmarshal of admin::getApplicationList response error:", err)
		return returnArray, returnBool
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("Response from admin::getApplicationList not ok: "+xmlRespon.State.ErrorRet, "error")
		color.Red("Response from admin::getApplicationList not ok:", xmlRespon.State.ErrorRet)
		return returnArray, returnBool
	}

	return xmlRespon.Params.Application, true
}

func getRequestCards(callref string) []dataStruct {
	//Use entityBrowseRecords to get asset entity records
	espXmlmc.SetParam("application", "com.hornbill.boardmanager")
	espXmlmc.SetParam("entity", "Card")
	espXmlmc.OpenElement("searchFilter")
	espXmlmc.SetParam("column", "h_key")
	espXmlmc.SetParam("value", callref)
	espXmlmc.SetParam("matchType", "exact")
	espXmlmc.CloseElement("searchFilter")
	browse, err := espXmlmc.Invoke("data", "entityBrowseRecords2")
	if err != nil {
		espLogger("Call to entityBrowseRecords2 failed when returning Card IDs for "+callref, "error")
		color.Red("Call to entityBrowseRecords2 failed when returning Card IDs for " + callref)
		return nil
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Unmarshal of entityBrowseRecords2 failed when returning Card IDs for "+callref, "error")
		color.Red("Unmarshal of entityBrowseRecords2 failed when returning Card IDs for " + callref)
		return nil
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("entityBrowseRecords2 was unsuccessful: "+xmlRespon.State.ErrorRet, "error")
		color.Red("entityBrowseRecords2 was unsuccessful: " + xmlRespon.State.ErrorRet)
		return nil
	}
	return xmlRespon.Params.RecordIDs
}
