package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/hornbill/color"
	hornbillHelpers "github.com/hornbill/goHornbillHelpers"
)

//getRecordCount - takes a table name, returns the total number of records in the entity
func getRecordCount(table string) int {
	strQuery := ""
	if table == "h_itsm_requests" {
		if len(cleanerConf.RequestTypes) > 0 {
			strQuery += " h_requesttype IN ("
			firstElement := true
			for _, reqType := range cleanerConf.RequestTypes {
				if !firstElement {
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
				if !firstElement {
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
				if !firstElement {
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

		if cleanerConf.RequestClosedDateFrom != "" {
			if strQuery != "" {
				strQuery += " AND"
			}

			closeDateFrom := cleanerConf.RequestClosedDateFrom
			boolIsDuration := durationRegex.MatchString(closeDateFrom)
			if boolIsDuration {
				fromTime, _ := hornbillHelpers.CalculateTimeDuration(time.Now(), closeDateFrom)
				closeDateFrom = fromTime.UTC().Format(datetimeFormat)
			}
			strQuery += " h_dateclosed >= '" + closeDateFrom + "'"
		}
		if cleanerConf.RequestClosedDateTo != "" {
			if strQuery != "" {
				strQuery += " AND"
			}
			closeDateTo := cleanerConf.RequestClosedDateTo
			boolIsDuration := durationRegex.MatchString(closeDateTo)
			if boolIsDuration {
				toTime, _ := hornbillHelpers.CalculateTimeDuration(time.Now(), closeDateTo)
				closeDateTo = toTime.UTC().Format(datetimeFormat)
			}
			strQuery += " h_dateclosed <= '" + closeDateTo + "'"
		}
	}

	if table == "h_cmdb_assets" {
		if cleanerConf.AssetClassID != "" {
			strQuery += " h_class = '" + cleanerConf.AssetClassID + "'"
		}

		if cleanerConf.AssetTypeID > 0 {
			if strQuery != "" {
				strQuery += " AND"
			}
			strQuery += " h_type = " + strconv.Itoa(cleanerConf.AssetTypeID)
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
		if !configDryRun || (configDryRun && currentBlock == 1) {
			espXmlmc.SetParam("rowstart", "0")
		} else {
			espXmlmc.SetParam("rowstart", strconv.Itoa((configBlockSize*currentBlock)-1))
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

		if cleanerConf.RequestClosedDateFrom != "" {
			closeDateFrom := cleanerConf.RequestClosedDateFrom
			boolIsDuration := durationRegex.MatchString(closeDateFrom)
			if boolIsDuration {
				fromTime, _ := hornbillHelpers.CalculateTimeDuration(time.Now(), closeDateFrom)
				closeDateFrom = fromTime.UTC().Format(datetimeFormat)
			}
			espXmlmc.SetParam("closedFromDateTime", closeDateFrom)
		}
		if cleanerConf.RequestClosedDateTo != "" {
			closeDateTo := cleanerConf.RequestClosedDateTo
			boolIsDuration := durationRegex.MatchString(closeDateTo)
			if boolIsDuration {
				toTime, _ := hornbillHelpers.CalculateTimeDuration(time.Now(), closeDateTo)
				closeDateTo = toTime.UTC().Format(datetimeFormat)
			}
			espXmlmc.SetParam("closedToDateTime", closeDateTo)
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
			espLogger("Requests queryExec was unsuccessful: "+xmlRespon.State.ErrorRet, "error")
			color.Red("Requests queryExec was unsuccessful: " + xmlRespon.State.ErrorRet)
			return nil
		}
		return xmlRespon.Params.RecordIDs
	}

	if entity == "Asset" {
		//Use a stored query to get asset IDs
		espXmlmc.SetParam("application", "com.hornbill.servicemanager")
		espXmlmc.SetParam("queryName", "Asset.getAssetsFiltered")
		espXmlmc.OpenElement("queryParams")
		if cleanerConf.AssetClassID != "" {
			espXmlmc.SetParam("assetClass", cleanerConf.AssetClassID)
		}
		if cleanerConf.AssetTypeID > 0 {
			espXmlmc.SetParam("assetType", strconv.Itoa(cleanerConf.AssetTypeID))
		}
		if !configDryRun || (configDryRun && currentBlock == 1) {
			espXmlmc.SetParam("rowstart", "0")
		} else {
			espXmlmc.SetParam("rowstart", strconv.Itoa((configBlockSize*currentBlock)-1))
		}
		espXmlmc.SetParam("limit", strconv.Itoa(configBlockSize))
		espXmlmc.CloseElement("queryParams")
		//espXmlmc.OpenElement("queryOptions")
		//espXmlmc.SetParam("queryType", "records")
		//espXmlmc.CloseElement("queryOptions")
		browse, err := espXmlmc.Invoke("data", "queryExec")
		if err != nil {
			espLogger("Call to queryExec [Assets] failed when returning block "+strconv.Itoa(currentBlock), "error")
			color.Red("Call to queryExec [Assets] failed when returning block " + strconv.Itoa(currentBlock))
			return nil
		}
		var xmlRespon xmlmcResponse
		err = xml.Unmarshal([]byte(browse), &xmlRespon)
		if err != nil {
			espLogger("Unmarshal of queryExec [Assets] data failed when returning block "+strconv.Itoa(currentBlock), "error")
			color.Red("Unmarshal of queryExec [Assets] data failed when returning block " + strconv.Itoa(currentBlock))
			return nil
		}
		if xmlRespon.MethodResult != "ok" {
			espLogger("Asset queryExec was unsuccessful: "+xmlRespon.State.ErrorRet, "error")
			color.Red("Asset queryExec was unsuccessful: " + xmlRespon.State.ErrorRet)
			return nil
		}
		return xmlRespon.Params.RecordIDs
	}

	//Use queryExec to get assetslinks entity records
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("queryName", "assetLinks")
	espXmlmc.OpenElement("queryParams")
	if !configDryRun || (configDryRun && currentBlock == 1) {
		espXmlmc.SetParam("rowstart", "0")
	} else {
		espXmlmc.SetParam("rowstart", strconv.Itoa((configBlockSize*currentBlock)-1))
	}
	espXmlmc.SetParam("limit", strconv.Itoa(configBlockSize))
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
		espLogger("AssetLinks queryExec was unsuccessful: "+xmlRespon.State.ErrorRet, "error")
		color.Red("AssetLinks queryExec was unsuccessful: " + xmlRespon.State.ErrorRet)
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
			espXmlmc.SetParam("taskStatus", k)
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
	//Use entityBrowseRecords2 to get asset links entity records
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
	//Use a stored query to get timer IDs
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

//getRequestBPMEvents - take call reference, return array of System Timers that are associated with it
func getRequestBPMEvents(callRef string) []dataStruct {
	//Use a stored query to get request IDs
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("queryName", "getRequestBPMEvents")
	espXmlmc.OpenElement("queryParams")
	espXmlmc.SetParam("inRequestId", callRef)
	espXmlmc.CloseElement("queryParams")
	browse, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		espLogger("Call to queryExec [getRequestBPMEvents] failed for Request "+callRef+" : "+fmt.Sprintf("%s", err), "error")
		color.Red("Call to queryExec [getRequestBPMEvents] failed for Request " + callRef)
		return nil
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Unmarshal of queryExec [getRequestBPMEvents] data failed for Request "+callRef+" : "+fmt.Sprintf("%s", err), "error")
		color.Red("Unmarshal of queryExec [getRequestBPMEvents] data failed for Request " + callRef)
		return nil
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("queryExec [getRequestBPMEvents] was unsuccessful: "+xmlRespon.State.ErrorRet, "error")
		color.Red("queryExec [getRequestBPMEvents] was unsuccessful: " + xmlRespon.State.ErrorRet)
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
	//Use entityBrowseRecords2 to get Board Manager cards against requests
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

func getRequestSLMEvents(callref string) []dataStruct {
	//Use entityBrowseRecords2 to get Request SLM Events entity records
	espXmlmc.SetParam("application", "com.hornbill.servicemanager")
	espXmlmc.SetParam("entity", "RequestSLMEvt")
	espXmlmc.OpenElement("searchFilter")
	espXmlmc.SetParam("column", "h_request_id")
	espXmlmc.SetParam("value", callref)
	espXmlmc.SetParam("matchType", "exact")
	espXmlmc.CloseElement("searchFilter")
	browse, err := espXmlmc.Invoke("data", "entityBrowseRecords2")
	if err != nil {
		espLogger("Call to entityBrowseRecords2 failed when returning SLM Events for "+callref, "error")
		color.Red("Call to entityBrowseRecords2 failed when returning SLM Events for " + callref)
		return nil
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("Unmarshal of entityBrowseRecords2 failed when returning SLM Events for "+callref, "error")
		color.Red("Unmarshal of entityBrowseRecords2 failed when returning SLM Events for " + callref)
		return nil
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("entityBrowseRecords2 was unsuccessful for SLM Events: "+xmlRespon.State.ErrorRet, "error")
		color.Red("entityBrowseRecords2 was unsuccessful for SLM Events: " + xmlRespon.State.ErrorRet)
		return nil
	}
	return xmlRespon.Params.RecordIDs
}
