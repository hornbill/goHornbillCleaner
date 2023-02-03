package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/color"
	hornbillHelpers "github.com/hornbill/goHornbillHelpers"
)

func printOnly(r rune) rune {
	if unicode.IsPrint(r) {
		return r
	}
	return -1
}

// getRecordCount - takes a table name, returns the total number of records in the entity
func getRecordCount(table, id, whereCol string) int {
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

		if len(cleanerConf.RequestCatalogItems) > 0 {
			if strQuery != "" {
				strQuery += " AND"
			}
			strQuery += " h_catalog_id IN ("
			firstElement := true
			for _, reqCatalog := range cleanerConf.RequestCatalogItems {
				if !firstElement {
					strQuery += ","
				}
				strQuery += strconv.Itoa(reqCatalog)
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
	}
	if table == "h_cmdb_links" && id != "" && whereCol != "" {
		strQuery += " " + whereCol + " = '" + id + "'"
	}
	espXmlmc.SetParam("database", "swdata")
	espXmlmc.SetParam("application", appSM)
	espXmlmc.SetParam("table", table)
	if strQuery != "" {
		espXmlmc.SetParam("where", strQuery)
	}
	requestXML := espXmlmc.GetParam()
	browse, err := espXmlmc.Invoke("data", "getRecordCount")
	if err != nil {
		espLogger("getRecordCount:Invoke:"+table+":"+strQuery+":"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		color.Red("getRecordCount Invoke failed for " + table + ":" + err.Error())
		return 0
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(strings.Map(printOnly, string(browse))), &xmlRespon)
	if err != nil {
		espLogger("getRecordCount:Unmarshal:"+table+":"+strQuery+":"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("getRecordCount Unmarshal failed for " + table + ":" + err.Error())
		return 0
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("getRecordCount:MethodResult:"+table+":"+strQuery+":"+xmlRespon.State.ErrorRet, "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("getRecordCount MethodResult failed for " + table + ":" + xmlRespon.State.ErrorRet)
		return 0
	}
	return xmlRespon.Params.RecordCount
}

func getEmailCount() (count int) {
	filters := cleanerConf.EmailFilters

	espXmlmc.SetParam("application", appCore)
	espXmlmc.SetParam("queryName", "systemEmailsCleaner")
	espXmlmc.OpenElement("queryParams")
	espXmlmc.SetParam("queryType", "count")
	for _, v := range filters.FolderIDs {
		espXmlmc.SetParam("folderId", strconv.Itoa(v))
	}
	if filters.ReceivedFrom != "" {
		espXmlmc.SetParam("msgDateFrom", filters.ReceivedFrom)
	}
	if filters.ReceivedTo != "" {
		espXmlmc.SetParam("msgDateTo", filters.ReceivedTo)
	}
	if filters.RecipientAddress != "" {
		espXmlmc.SetParam("recipientAddress", filters.RecipientAddress)
		if filters.RecipientClass != "" {
			if class, ok := mailRecipientClasses[filters.RecipientClass]; ok {
				espXmlmc.SetParam("recipientClass", strconv.Itoa(class))
			}
		}
	}
	if filters.Subject != "" {
		espXmlmc.SetParam("subject", filters.Subject)
	}
	espXmlmc.CloseElement("queryParams")
	requestXML := espXmlmc.GetParam()

	browse, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		espLogger("count:queryExec:Invoke:"+appSM+":systemEmailsCleaner:count:"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		color.Red("queryExec Invoke failed to get count for " + appSM + ":systemEmailsCleaner:count:" + err.Error())
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(strings.Map(printOnly, string(browse))), &xmlRespon)
	if err != nil {
		espLogger("queryExec:Unmarshal:"+appSM+":systemEmailsCleaner:count:"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec Unmarshal failed to get count for " + appSM + ":systemEmailsCleaner:count:" + err.Error())
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("queryExec:MethodResult:"+appSM+":systemEmailsCleaner:count:"+xmlRespon.State.ErrorRet, "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec MethodResult failed to get count for " + appSM + ":systemEmailsCleaner:count:" + xmlRespon.State.ErrorRet)
		return
	}
	count = xmlRespon.Params.RecordIDs[0].MessageCount
	return
}

func getAssetCount() int {
	//Use a stored query to get asset IDs
	espXmlmc.SetParam("application", appSM)
	espXmlmc.SetParam("queryName", "Asset.getAssetsFiltered")
	espXmlmc.OpenElement("queryParams")
	espXmlmc.SetParam("resultType", "count")
	if cleanerConf.AssetClassID != "" {
		espXmlmc.SetParam("assetClass", cleanerConf.AssetClassID)
	}
	if len(cleanerConf.AssetFilters) > 0 {
		var filterList []filterStuct
		for _, v := range cleanerConf.AssetFilters {
			filter := filterStuct{}
			filter.ColumnName = v.ColumnName
			filter.ColumnValue = v.ColumnValue
			filter.Operator = v.Operator
			filter.IsGeneralProperty = v.IsGeneralProperty
			filterList = append(filterList, filter)
		}
		filters, err := json.Marshal(filterList)
		if err != nil {
			espLogger("getAssetCount:Filters:Marshal:"+err.Error(), "error")
			color.Red("getAssetCount could not marshal filters into JSON: " + err.Error())
			return 0
		}
		espXmlmc.SetParam("filters", string(filters))
	}
	espXmlmc.CloseElement("queryParams")
	requestXML := espXmlmc.GetParam()
	browse, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		espLogger("count:queryExec:Invoke:"+appSM+":Asset.getAssetsFiltered:count:"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		color.Red("queryExec Invoke failed to get count for " + appSM + ":Asset.getAssetsFiltered:count:" + err.Error())
		return 0
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(strings.Map(printOnly, string(browse))), &xmlRespon)
	if err != nil {
		espLogger("queryExec:Unmarshal:"+appSM+":Asset.getAssetsFiltered:count:"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec Unmarshal failed to get count for " + appSM + ":Asset.getAssetsFiltered:count:" + err.Error())
		return 0
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("queryExec:MethodResult:"+appSM+":Asset.getAssetsFiltered:count:"+xmlRespon.State.ErrorRet, "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec MethodResult failed to get count for " + appSM + ":Asset.getAssetsFiltered:count:" + xmlRespon.State.ErrorRet)
		return 0
	}
	return xmlRespon.Params.RecordIDs[0].Count
}

func getServiceAvailabilityHistoryCount() int {
	//Use a stored query to get service availability record count
	espXmlmc.SetParam("application", appSM)
	espXmlmc.SetParam("queryName", "_listServiceStatusHistory")
	espXmlmc.OpenElement("queryParams")
	for _, v := range cleanerConf.ServiceAvailabilityServiceIDs {
		espXmlmc.SetParam("serviceId", strconv.Itoa(v))
	}
	espXmlmc.CloseElement("queryParams")
	espXmlmc.OpenElement("queryOptions")
	espXmlmc.SetParam("querytype", "count")
	espXmlmc.CloseElement("queryOptions")
	requestXML := espXmlmc.GetParam()
	browse, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		espLogger("count:queryExec:Invoke:"+appSM+":_listServiceStatusHistory:count:"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		color.Red("queryExec Invoke failed to get count for " + appSM + ":_listServiceStatusHistory:count:" + err.Error())
		return 0
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(strings.Map(printOnly, string(browse))), &xmlRespon)
	if err != nil {
		espLogger("queryExec:Unmarshal:"+appSM+":_listServiceStatusHistory:count:"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec Unmarshal failed to get count for " + appSM + ":_listServiceStatusHistory:count:" + err.Error())
		return 0
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("queryExec:MethodResult:"+appSM+":_listServiceStatusHistory:count:"+xmlRespon.State.ErrorRet, "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec MethodResult failed to get count for " + appSM + ":_listServiceStatusHistory:count:" + xmlRespon.State.ErrorRet)
		return 0
	}
	return xmlRespon.Params.RecordIDs[0].Count
}

// getRecordIDs - returns an array of records for deletion
func getRecordIDs(entity string) []dataStruct {
	if currentBlock < totalBlocks {
		fmt.Println("Returning block " + strconv.Itoa(displayBlock) + " of " + strconv.Itoa(totalBlocks) + " blocks of records from " + entity + " entity...")
		if entity == "Requests" {
			return getRequestRecords()
		}

		if entity == "Asset" {
			return getAssetRecords()
		}

		if entity == "ServiceStatusHistory" {
			return getServiceStatusHistoryRecords()
		}

		if entity == "Email" {
			return getEmailRecords()
		}
	}
	color.Green("All " + entity + " records processed.")
	return nil
}

func getEmailRecords() []dataStruct {
	filters := cleanerConf.EmailFilters
	espXmlmc.SetParam("application", appCore)
	espXmlmc.SetParam("queryName", "systemEmailsCleaner")
	espXmlmc.OpenElement("queryParams")
	espXmlmc.SetParam("queryType", "records")
	for _, v := range filters.FolderIDs {
		espXmlmc.SetParam("folderId", strconv.Itoa(v))
	}
	if filters.ReceivedFrom != "" {
		espXmlmc.SetParam("msgDateFrom", filters.ReceivedFrom)
	}
	if filters.ReceivedTo != "" {
		espXmlmc.SetParam("msgDateTo", filters.ReceivedTo)
	}
	if filters.RecipientAddress != "" {
		espXmlmc.SetParam("recipientAddress", filters.RecipientAddress)
		if filters.RecipientClass != "" {
			if class, ok := mailRecipientClasses[filters.RecipientClass]; ok {
				espXmlmc.SetParam("recipientClass", strconv.Itoa(class))
			}
		}
	}
	if filters.Subject != "" {
		espXmlmc.SetParam("subject", filters.Subject)
	}

	if !configDryRun || (configDryRun && currentBlock == 0) {
		espXmlmc.SetParam("rowstart", "0")
	} else {
		espXmlmc.SetParam("rowstart", strconv.Itoa(configBlockSize*currentBlock))
	}
	espXmlmc.SetParam("limit", strconv.Itoa(configBlockSize))
	espXmlmc.CloseElement("queryParams")
	requestXML := espXmlmc.GetParam()
	browse, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		espLogger("queryExec:Invoke:"+appCore+":systemEmailsCleaner:"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		color.Red("queryExec Invoke failed for " + appSM + ":systemEmailsCleaner:" + err.Error())
		return nil
	}
	var xmlRespon xmlmcResponse

	err = xml.Unmarshal([]byte(strings.Map(printOnly, string(browse))), &xmlRespon)
	if err != nil {
		espLogger("queryExec:Unmarshal:"+appCore+":systemEmailsCleaner:"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec Unmarshal failed for " + appCore + ":systemEmailsCleaner:" + err.Error())
		return nil
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("queryExec:MethodResult:"+appCore+":systemEmailsCleaner:"+xmlRespon.State.ErrorRet, "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec MethodResult failed for " + appCore + ":systemEmailsCleaner:" + xmlRespon.State.ErrorRet)
		return nil
	}
	return xmlRespon.Params.RecordIDs
}

func getAssetRecords() []dataStruct {
	//Use a stored query to get asset IDs
	espXmlmc.SetParam("application", appSM)
	espXmlmc.SetParam("queryName", "Asset.getAssetsFiltered")
	espXmlmc.OpenElement("queryParams")
	espXmlmc.SetParam("resultType", "data")
	if cleanerConf.AssetClassID != "" {
		espXmlmc.SetParam("assetClass", cleanerConf.AssetClassID)
	}
	if len(cleanerConf.AssetFilters) > 0 {
		var filterList []filterStuct
		for _, v := range cleanerConf.AssetFilters {
			filter := filterStuct{}
			filter.ColumnName = v.ColumnName
			filter.ColumnValue = v.ColumnValue
			filter.Operator = v.Operator
			filter.IsGeneralProperty = v.IsGeneralProperty
			filterList = append(filterList, filter)
		}
		filters, err := json.Marshal(filterList)
		if err != nil {
			espLogger("getRecordIds:Filters:Marshal:"+err.Error(), "error")
			color.Red("getRecordIds could not marshal filters into JSON: " + err.Error())
			return nil

		}
		espXmlmc.SetParam("filters", string(filters))
	}
	if !configDryRun || (configDryRun && currentBlock == 0) {
		espXmlmc.SetParam("rowstart", "0")
	} else {
		espXmlmc.SetParam("rowstart", strconv.Itoa(configBlockSize*currentBlock))
	}
	espXmlmc.SetParam("limit", strconv.Itoa(configBlockSize))
	espXmlmc.CloseElement("queryParams")
	requestXML := espXmlmc.GetParam()
	browse, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		espLogger("queryExec:Invoke:"+appSM+":Asset.getAssetsFiltered:"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		color.Red("queryExec Invoke failed for " + appSM + ":Asset.getAssetsFiltered:" + err.Error())
		return nil
	}
	var xmlRespon xmlmcResponse

	err = xml.Unmarshal([]byte(strings.Map(printOnly, string(browse))), &xmlRespon)
	if err != nil {
		espLogger("queryExec:Unmarshal:"+appSM+":Asset.getAssetsFiltered:"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec Unmarshal failed for " + appSM + ":Asset.getAssetsFiltered:" + err.Error())
		return nil
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("queryExec:MethodResult:"+appSM+":Asset.getAssetsFiltered:"+xmlRespon.State.ErrorRet, "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec MethodResult failed for " + appSM + ":Asset.getAssetsFiltered:" + xmlRespon.State.ErrorRet)
		return nil
	}
	return xmlRespon.Params.RecordIDs
}

func getAssetLinkIDs(assetURN, direction string) []dataStruct {
	if currentBlock < totalBlocks {
		fmt.Println("Returning block " + strconv.Itoa(displayBlock) + " of " + strconv.Itoa(totalBlocks) + " " + direction + " blocks of records from AssetLinks entity...")
	} else {
		color.Green("All " + direction + " AssetLinks records processed.")
	}
	//Use queryExec to get assetslinks entity records
	espXmlmc.SetParam("application", appSM)
	espXmlmc.SetParam("queryName", "assetLinks")
	espXmlmc.OpenElement("queryParams")
	if direction == "left" {
		espXmlmc.SetParam("leftId", assetURN)
	} else {
		espXmlmc.SetParam("rightId", assetURN)
	}

	if !configDryRun || (configDryRun && currentBlock == 0) {
		espXmlmc.SetParam("rowstart", "0")
	} else {
		espXmlmc.SetParam("rowstart", strconv.Itoa(configBlockSize*currentBlock))
	}
	espXmlmc.SetParam("limit", strconv.Itoa(configBlockSize))
	espXmlmc.CloseElement("queryParams")
	requestXML := espXmlmc.GetParam()
	browse, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		espLogger("Call to queryExec [assetLinks] failed when returning block "+strconv.Itoa(displayBlock), "error")
		espLogger("Request XML: "+requestXML, "debug")
		color.Red("Call to queryExec [assetLinks] failed when returning block " + strconv.Itoa(displayBlock))
		return nil
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(strings.Map(printOnly, string(browse))), &xmlRespon)
	if err != nil {
		espLogger("Unmarshal of queryExec [assetLinks] data failed when returning block "+strconv.Itoa(displayBlock), "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("Unmarshal of queryExec [assetLinks] data failed when returning block " + strconv.Itoa(displayBlock))
		return nil
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("AssetLinks queryExec was unsuccessful: "+xmlRespon.State.ErrorRet, "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("AssetLinks queryExec was unsuccessful: " + xmlRespon.State.ErrorRet)
		return nil
	}
	return xmlRespon.Params.RecordIDs
}

func getRequestRecords() []dataStruct {
	//Use a stored query to get request IDs
	espXmlmc.SetParam("application", appSM)
	espXmlmc.SetParam("queryName", "_listRequestsOfType")
	espXmlmc.OpenElement("queryParams")
	for _, reqType := range cleanerConf.RequestTypes {
		espXmlmc.SetParam("type", reqType)
	}
	if !configDryRun || (configDryRun && currentBlock == 0) {
		espXmlmc.SetParam("rowstart", "0")
	} else {
		espXmlmc.SetParam("rowstart", strconv.Itoa(configBlockSize*currentBlock))
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

	for _, reqCatalog := range cleanerConf.RequestCatalogItems {
		espXmlmc.SetParam("catalog", strconv.Itoa(reqCatalog))
	}

	espXmlmc.CloseElement("queryParams")
	requestXML := espXmlmc.GetParam()
	browse, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		espLogger("queryExec:Invoke:"+appSM+":_listRequestsOfType:"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		color.Red("queryExec Invoke failed for " + appSM + ":_listRequestsOfType:" + err.Error())
		return nil
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(strings.Map(printOnly, string(browse))), &xmlRespon)
	if err != nil {
		espLogger("queryExec:Unmarshal:"+appSM+":_listRequestsOfType:"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec Unmarshal failed for " + appSM + ":_listRequestsOfType:" + err.Error())
		return nil
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("queryExec:MethodResult:"+appSM+":_listRequestsOfType:"+xmlRespon.State.ErrorRet, "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec MethodResult failed for " + appSM + ":_listRequestsOfType:" + xmlRespon.State.ErrorRet)
		return nil
	}
	return xmlRespon.Params.RecordIDs
}

// getRequestTasks - take a call reference, get all associated request tasks
func getRequestTasks(callRef string) map[string][]taskStruct {
	//First get request task counters so we can set correct state
	espXmlmc.SetParam("objectRefUrn", "urn:sys:entity:"+appSM+":Requests:"+callRef)
	espXmlmc.SetParam("counters", "true")
	requestXML := espXmlmc.GetParam()
	getCounters, err := espXmlmc.Invoke("apps/com.hornbill.core/Task", "getEntityTasks")
	if err != nil {
		espLogger("getEntityTasks:Invoke:Request:"+callRef+":"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		color.Red("getEntityTasks Invoke failed for Request:" + callRef + ":" + err.Error())
		return nil
	}
	var xmlResponCount xmlmcTaskResponse
	err = xml.Unmarshal([]byte(strings.Map(printOnly, string(getCounters))), &xmlResponCount)
	if err != nil {
		espLogger("getEntityTasks:Unmarshal:Request:"+callRef+":"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+getCounters, "debug")
		color.Red("getEntityTasks Unmarshal failed for Request:" + callRef + ":" + err.Error())
		return nil
	}
	if xmlResponCount.MethodResult != "ok" {
		espLogger("getEntityTasks:MethodResult:Request:"+callRef+":"+xmlResponCount.State.ErrorRet, "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+getCounters, "debug")
		color.Red("getEntityTasks MethodResult failed for Request:" + callRef + ":" + xmlResponCount.State.ErrorRet)
		return nil
	}

	objCounter := make(map[string]interface{})
	json.Unmarshal([]byte(xmlResponCount.Counters), &objCounter)

	if len(objCounter) > 0 {
		//--  get task IDs
		espXmlmc.SetParam("objectRefUrn", "urn:sys:entity:"+appSM+":Requests:"+callRef)

		for k := range objCounter {
			espXmlmc.SetParam("taskStatus", k)
		}
		requestXML := espXmlmc.GetParam()
		browse, errTask := espXmlmc.Invoke("apps/com.hornbill.core/Task", "getEntityTasks")
		if errTask != nil {
			espLogger("getEntityTasks:taskStatus:Invoke:Request:"+callRef+":"+errTask.Error(), "error")
			espLogger("Request XML: "+requestXML, "debug")
			color.Red("getEntityTasks Invoke failed for Request:" + callRef + ":" + errTask.Error())
			return nil
		}
		var xmlRespon xmlmcTaskResponse
		errTask = xml.Unmarshal([]byte(strings.Map(printOnly, string(browse))), &xmlRespon)
		if errTask != nil {
			espLogger("getEntityTasks:taskStatus:Unmarshal:Request:"+callRef+":"+errTask.Error(), "error")
			espLogger("Request XML: "+requestXML, "debug")
			espLogger("Response XML: "+browse, "debug")
			color.Red("getEntityTasks Unmarshal failed for Request:" + callRef + ":" + errTask.Error())
			return nil
		}
		if xmlRespon.MethodResult != "ok" {
			espLogger("getEntityTasks:taskStatus:MethodResult:Request:"+callRef+":"+xmlRespon.State.ErrorRet, "error")
			espLogger("Request XML: "+requestXML, "debug")
			espLogger("Response XML: "+browse, "debug")
			color.Red("getEntityTasks MethodResult failed for Request:" + callRef + ":" + xmlRespon.State.ErrorRet)
			return nil
		}
		//Unmarshall JSON string in to map containing taskStruct slices
		objTasks := make(map[string][]taskStruct)
		json.Unmarshal([]byte(xmlRespon.TasksJSON), &objTasks)
		return objTasks
	}
	return nil
}

// getRequestWorkflow - take a call reference, get all associated rBPM workflow ID
func getRequestWorkflow(callRef string) string {
	returnWorkflowID := ""
	espXmlmc.SetParam("application", appSM)
	espXmlmc.SetParam("entity", "Requests")
	espXmlmc.SetParam("keyValue", callRef)
	requestXML := espXmlmc.GetParam()
	browse, err := espXmlmc.Invoke("data", "entityGetRecord")
	if err != nil {
		espLogger("entityGetRecord:Invoke:Requests:"+callRef+":"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		color.Red("entityGetRecord Invoke failed for Requests:" + callRef + ":" + err.Error())
		return ""
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(strings.Map(printOnly, string(browse))), &xmlRespon)
	if err != nil {
		espLogger("entityGetRecord:Unmarshal:Requests:"+callRef+":"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("entityGetRecord Unmarshal failed for Requests:" + callRef + ":" + err.Error())
		return ""
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("entityGetRecord:Unmarshal:Requests:"+callRef+":"+xmlRespon.State.ErrorRet, "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("entityGetRecord Unmarshal failed for Requests:" + callRef + ":" + xmlRespon.State.ErrorRet)
		return ""
	}
	returnWorkflowID = xmlRespon.Params.BPMID
	return returnWorkflowID
}

func getServiceStatusHistoryRecords() []dataStruct {
	//Use a stored query to get asset IDs
	espXmlmc.SetParam("application", appSM)
	espXmlmc.SetParam("queryName", "_listServiceStatusHistory")
	espXmlmc.OpenElement("queryParams")
	for _, v := range cleanerConf.ServiceAvailabilityServiceIDs {
		espXmlmc.SetParam("serviceId", strconv.Itoa(v))
	}
	if !configDryRun || (configDryRun && currentBlock == 0) {
		espXmlmc.SetParam("rowstart", "0")
	} else {
		espXmlmc.SetParam("rowstart", strconv.Itoa(configBlockSize*currentBlock))
	}
	espXmlmc.SetParam("limit", strconv.Itoa(configBlockSize))
	espXmlmc.CloseElement("queryParams")
	espXmlmc.OpenElement("queryOptions")
	espXmlmc.SetParam("querytype", "records")
	espXmlmc.CloseElement("queryOptions")
	requestXML := espXmlmc.GetParam()
	browse, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		espLogger("queryExec:Invoke:"+appSM+":_listServiceStatusHistory:"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		color.Red("queryExec Invoke failed for " + appSM + ":_listServiceStatusHistory:" + err.Error())
		return nil
	}
	var xmlRespon xmlmcResponse

	err = xml.Unmarshal([]byte(strings.Map(printOnly, string(browse))), &xmlRespon)
	if err != nil {
		espLogger("queryExec:Unmarshal:"+appSM+":_listServiceStatusHistory:"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec Unmarshal failed for " + appSM + ":_listServiceStatusHistory:" + err.Error())
		return nil
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("queryExec:MethodResult:"+appSM+":_listServiceStatusHistory:"+xmlRespon.State.ErrorRet, "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec MethodResult failed for " + appSM + ":_listServiceStatusHistory:" + xmlRespon.State.ErrorRet)
		return nil
	}
	return xmlRespon.Params.RecordIDs
}

func getAppList() ([]appsStruct, bool) {
	var returnArray []appsStruct
	returnBool := false
	espXmlmc.SetTimeout(180)
	apps, err := espXmlmc.Invoke("admin", "getApplicationList")
	if err != nil {
		espLogger("getApplicationList:Invoke:"+err.Error(), "error")
		color.Red("getApplicationList Invoke failed:" + err.Error())
		return returnArray, returnBool
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(strings.Map(printOnly, string(apps))), &xmlRespon)
	if err != nil {
		espLogger("getApplicationList:Unmarshal:"+err.Error(), "error")
		espLogger("Response XML: "+apps, "debug")
		color.Red("getApplicationList Unmarshal failed:" + err.Error())
		return returnArray, returnBool
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("getApplicationList:MethodResult:"+xmlRespon.State.ErrorRet, "error")
		espLogger("Response XML: "+apps, "debug")
		color.Red("getApplicationList MethodResult failed:" + xmlRespon.State.ErrorRet)
		return returnArray, returnBool
	}
	return xmlRespon.Params.Application, true
}

func entityBrowseRecords(application, entity, matchScope string, searchFilters []browseRecordsParamsStruct) []dataStruct {
	espXmlmc.SetParam("application", application)
	espXmlmc.SetParam("entity", entity)
	if matchScope != "" {
		espXmlmc.SetParam("matchScope", matchScope)
	}
	var logSearchFilter []string
	for _, v := range searchFilters {
		espXmlmc.OpenElement("searchFilter")
		espXmlmc.SetParam("column", v.Column)
		espXmlmc.SetParam("value", v.Value)
		espXmlmc.SetParam("matchType", v.MatchType)
		espXmlmc.CloseElement("searchFilter")
		logSearchFilter = append(logSearchFilter, v.Column+" = '"+v.Value+"' ("+v.MatchType+")")
	}
	logFilter := strings.Join(logSearchFilter[:], " AND ")
	requestXML := espXmlmc.GetParam()
	browse, err := espXmlmc.Invoke("data", "entityBrowseRecords2")
	if err != nil {
		espLogger("entityBrowseRecords2:Invoke:"+application+":"+entity+":"+logFilter+":"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		color.Red("entityBrowseRecords2 Invoke failed for " + application + ":" + entity + ":" + err.Error())
		return nil
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(strings.Map(printOnly, string(browse))), &xmlRespon)
	if err != nil {
		espLogger("entityBrowseRecords2:Unmarshal:"+application+":"+entity+":"+logFilter+":"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("entityBrowseRecords2 Unmarshal failed for " + application + ":" + entity + ":" + err.Error())
		return nil
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("entityBrowseRecords2:MethodResult:"+application+":"+entity+":"+logFilter+":"+xmlRespon.State.ErrorRet, "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("entityBrowseRecords2 MethodResult failed for " + application + ":" + entity + ":" + xmlRespon.State.ErrorRet)
		return nil
	}
	return xmlRespon.Params.RecordIDs
}

// queryExec -
func queryExec(application, queryName string, queryParams []queryParamsStruct) []dataStruct {
	//Use a stored query to get timer IDs
	espXmlmc.SetParam("application", application)
	espXmlmc.SetParam("queryName", queryName)
	espXmlmc.OpenElement("queryParams")
	var queryKeyVal []string
	for _, param := range queryParams {
		espXmlmc.SetParam(param.Name, param.Value)
		queryKeyVal = append(queryKeyVal, param.Name+":"+param.Value)
	}

	logKeyVals := strings.Join(queryKeyVal[:], "|")

	espXmlmc.CloseElement("queryParams")
	requestXML := espXmlmc.GetParam()
	browse, err := espXmlmc.Invoke("data", "queryExec")
	if err != nil {
		espLogger("queryExec:Invoke:"+application+":"+queryName+":"+logKeyVals+":"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		color.Red("queryExec Invoke failed for " + application + ":" + queryName + ":" + err.Error())
		return nil
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(strings.Map(printOnly, string(browse))), &xmlRespon)
	if err != nil {
		espLogger("queryExec:Unmarshal:"+application+":"+queryName+":"+logKeyVals+":"+err.Error(), "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec Unmarshal failed for " + application + ":" + queryName + ":" + err.Error())
		return nil
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("queryExec:MethodResult:"+application+":"+queryName+":"+logKeyVals+":"+xmlRespon.State.ErrorRet, "error")
		espLogger("Request XML: "+requestXML, "debug")
		espLogger("Response XML: "+browse, "debug")
		color.Red("queryExec MethodResult failed for " + application + ":" + queryName + ":" + xmlRespon.State.ErrorRet)
		return nil
	}
	return xmlRespon.Params.RecordIDs
}
