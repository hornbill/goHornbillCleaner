package main

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

//deleteRecords - deletes the records in the array generated by getRecordIDs
func deleteRecords(entity string, records []dataStruct) {

	if configDryRun {
		for _, v := range records {
			id := ""
			description := ""
			if entity == "Requests" {
				id = v.RequestID
				description = "Logged On: " + v.LogDate
				if v.CloseDate != "" && v.CloseDate != "<nil>" {
					description += ", Closed On: " + v.CloseDate
				}
			} else if entity == "Asset" {
				id = v.AssetID
				description = "Asset Name: " + v.AssetName
			}
			espLogger("["+strings.ToUpper(entity)+"] ID:"+id+" "+description, "info")
		}
		currentBlock++
		return
	}

	fmt.Println("Deleting block " + strconv.Itoa(currentBlock) + " of " + strconv.Itoa(totalBlocks) + " blocks of records from " + entity + " entity. Please wait...")

	if entity == "Requests" {
		//Go through requests, and delete any associated records
		for _, callRef := range records {

			//-- BPM Events
			var bpmQueryParams []queryParamsStruct
			bpmQueryParams = append(bpmQueryParams, queryParamsStruct{Name: "inRequestId", Value: callRef.RequestID})
			bpmTimerIDs := queryExec(appSM, "getRequestBPMEvents", bpmQueryParams)
			if len(bpmTimerIDs) != 0 {
				for _, timerRecord := range bpmTimerIDs {
					if timerRecord.BPMEventID != "<nil>" && timerRecord.BPMEventID != "" {
						deleteEvent(timerRecord.BPMEventID)
					}
					if timerRecord.BPMTimerID != "<nil>" && timerRecord.BPMTimerID != "" {
						deleteTimer(timerRecord.BPMTimerID)
					}
				}
			}

			//-- System Timers
			var stQueryParams []queryParamsStruct
			stQueryParams = append(stQueryParams, queryParamsStruct{Name: "requestId", Value: callRef.RequestID})
			sysTimerIDs := queryExec(appSM, "getRequestSystemTimers", stQueryParams)
			if len(sysTimerIDs) != 0 {
				for _, timerID := range sysTimerIDs {
					if timerID.TimerID != "<nil>" && timerID.TimerID != "" {
						deleteTimer(timerID.TimerID)
					}
				}
			}

			//-- SLM Timer Events
			var slmParams []browseRecordsParamsStruct
			slmParams = append(slmParams, browseRecordsParamsStruct{Column: "h_request_id", Value: callRef.RequestID, MatchType: "exact"})
			slmEventIDs := entityBrowseRecords(appSM, "RequestSLMEvt", "", slmParams)
			if len(slmEventIDs) != 0 {
				for _, eventRecord := range slmEventIDs {
					if eventRecord.BPMEventID != "<nil>" && eventRecord.BPMEventID != "" {
						deleteEvent(eventRecord.BPMEventID)
					}
				}
			}

			//-- Spawned workflow
			requestWorkflow := getRequestWorkflow(callRef.RequestID)
			if requestWorkflow != "<nil>" && requestWorkflow != "" {
				deleteWorkflow(requestWorkflow)
			}

			//-- Request Tasks
			requestTasks := getRequestTasks(callRef.RequestID)
			for _, stateMap := range requestTasks {
				for _, taskMap := range stateMap {
					if taskMap.TaskID != "<nil>" && taskMap.TaskID != "" {
						deleteTask(taskMap.TaskID)
					}
				}
			}

			//-- Asset Associations
			var assetLinksParams []browseRecordsParamsStruct
			callrefURN := "urn:sys:entity:" + appSM + ":Requests:" + callRef.RequestID
			assetLinksParams = append(assetLinksParams, browseRecordsParamsStruct{Column: "h_fk_id_l", Value: callrefURN, MatchType: "exact"})
			assetLinksParams = append(assetLinksParams, browseRecordsParamsStruct{Column: "h_fk_id_r", Value: callrefURN, MatchType: "exact"})
			requestAssets := entityBrowseRecords(appSM, "AssetsLinks", "any", assetLinksParams)
			for _, linkID := range requestAssets {
				if linkID.AssetLinkID != "<nil>" && linkID.AssetLinkID != "" {
					entityDeleteRecords(appSM, "AssetsLinks", []string{linkID.AssetLinkID}, false, false)
				}
			}

			//Board Manager Cards
			if boardManagerInstalled {
				var cardParams []browseRecordsParamsStruct
				cardParams = append(cardParams, browseRecordsParamsStruct{Column: "h_key", Value: callRef.RequestID, MatchType: "exact"})
				requestCards := entityBrowseRecords(appBM, "Card", "", cardParams)
				for _, cardID := range requestCards {
					if cardID.CardID != "<nil>" && cardID.CardID != "" {
						deleteCard(cardID.CardID)
					}
				}
			}

		}
	}

	if entity == "Asset" && configManagerInstalled {
		//Go through assets, delete any associated CM records
		for _, asset := range records {
			//Process Impacts
			var impQueryParams []queryParamsStruct
			impQueryParams = append(impQueryParams, queryParamsStruct{Name: "entityId", Value: asset.AssetID})
			ciImpacts := queryExec(appCM, "getImpacts", impQueryParams)
			for _, impact := range ciImpacts {
				if impact.AssetImpactID != "<nil>" && impact.AssetImpactID != "" {
					entityDeleteRecords(appCM, "ConfigurationItemsImpact", []string{impact.AssetImpactID}, false, false)
				}
			}

			//Process Dependencies
			var depQueryParams []queryParamsStruct
			depQueryParams = append(depQueryParams, queryParamsStruct{Name: "entityId", Value: asset.AssetID})
			ciDependencies := queryExec(appCM, "getDependencies", depQueryParams)
			for _, dependency := range ciDependencies {
				if dependency.AssetDependencyID != "<nil>" && dependency.AssetDependencyID != "" {
					entityDeleteRecords(appCM, "ConfigurationItemsDependency", []string{dependency.AssetDependencyID}, false, false)
				}
			}

			//Process Policies
			var polQueryParams []queryParamsStruct
			polQueryParams = append(polQueryParams, queryParamsStruct{Name: "entityId", Value: asset.AssetID})
			ciPolicies := queryExec(appCM, "getItemsInPolicy", polQueryParams)
			for _, dependency := range ciPolicies {
				if dependency.AssetPolicyID != "<nil>" && dependency.AssetPolicyID != "" {
					entityDeleteRecords(appCM, "ConfigurationItemsInPolicy", []string{dependency.AssetPolicyID}, false, false)
				}
			}
		}
	}
	//Now delete the block of records
	var idsToDelete []string
	for _, v := range records {
		id := ""
		if entity == "Requests" {
			id = v.RequestID
		} else if entity == "Asset" {
			id = v.AssetID
		} else if entity == "AssetsLinks" {
			id = v.AssetLinkID
		}
		idsToDelete = append(idsToDelete, id)
	}
	entityDeleteRecords(appSM, entity, idsToDelete, false, false)
	color.Green("Block " + strconv.Itoa(currentBlock) + " of " + strconv.Itoa(totalBlocks) + " deleted.")
	currentBlock++
}

func deleteUser(strUser string) {
	//Now delete the user
	espXmlmc.SetParam("userId", strUser)

	deleted, err := espXmlmc.Invoke("admin", "userDelete")
	if err != nil {
		espLogger("userDelete:Invoke:"+strUser+":"+err.Error(), "error")
		color.Red("userDelete Invoke failed for " + strUser + ":" + err.Error())
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(deleted), &xmlRespon)
	if err != nil {
		espLogger("userDelete:Unmarshal:"+strUser+":"+err.Error(), "error")
		color.Red("userDelete Unmarshal failed for " + strUser + ":" + err.Error())
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("userDelete:MethodResult:"+strUser+":"+xmlRespon.State.ErrorRet, "error")
		color.Red("userDelete MethodResult failed for " + strUser + ":" + xmlRespon.State.ErrorRet)
		return
	}
	espLogger("User "+strUser+" deleted.", "notice")
}

//deleteTimer - Takes a System Timer ID, sends it to time::timerDelete API for safe deletion
func deleteTimer(timerID string) {
	espXmlmc.SetParam("timerId", timerID)
	browse, err := espXmlmc.Invoke("time", "timerDelete")
	if err != nil {
		espLogger("timerDelete:Invoke:"+timerID+":"+err.Error(), "error")
		color.Red("timerDelete Invoke failed for " + timerID + ":" + err.Error())
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("timerDelete:Unmarshal:"+timerID+":"+err.Error(), "error")
		color.Red("timerDelete Unmarshal failed for " + timerID + ":" + err.Error())
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("timerDelete:MethodResult:"+timerID+":"+xmlRespon.State.ErrorRet, "error")
		color.Red("timerDelete MethodResult failed for " + timerID + ":" + xmlRespon.State.ErrorRet)
		return
	}
	espLogger("Timer "+timerID+" deleted", "notice")
}

//deleteEvent - Takes a System Timer Event ID, sends it to time::timerDelete API for safe deletion
func deleteEvent(eventID string) {
	espXmlmc.SetParam("eventId", eventID)
	browse, err := espXmlmc.Invoke("time", "timerEventDelete")
	if err != nil {
		espLogger("timerEventDelete:Invoke:"+eventID+":"+err.Error(), "error")
		color.Red("timerEventDelete Invoke failed for " + eventID + ":" + err.Error())
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("timerEventDelete:Unmarshal:"+eventID+":"+err.Error(), "error")
		color.Red("timerEventDelete Unmarshal failed for " + eventID + ":" + err.Error())
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("timerEventDelete:MethodResult:"+eventID+":"+xmlRespon.State.ErrorRet, "error")
		color.Red("timerEventDelete MethodResult failed for " + eventID + ":" + xmlRespon.State.ErrorRet)
		return
	}
	espLogger("Timer Event "+eventID+" deleted", "notice")
}

//deleteTask - Takes a Task ID, sends it to task::taskDelete API for safe deletion
func deleteTask(taskID string) {
	espXmlmc.SetParam("taskId", taskID)
	browse, err := espXmlmc.Invoke("task", "taskDelete")
	if err != nil {
		espLogger("taskDelete:Invoke:"+taskID+":"+err.Error(), "error")
		color.Red("taskDelete Invoke failed for " + taskID + ":" + err.Error())
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("taskDelete:Unmarshal:"+taskID+":"+err.Error(), "error")
		color.Red("taskDelete Unmarshal failed for " + taskID + ":" + err.Error())
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("taskDelete:MethodResult:"+taskID+":"+xmlRespon.State.ErrorRet, "error")
		color.Red("taskDelete MethodResult failed for " + taskID + ":" + xmlRespon.State.ErrorRet)
		return
	}
	espLogger("Task "+taskID+" deleted", "notice")
}

//deleteWorkflow - Takes a Workflow ID, sends it to bpm::processDelete API for safe deletion
func deleteWorkflow(workflowID string) {
	espXmlmc.SetParam("identifier", workflowID)
	browse, err := espXmlmc.Invoke("bpm", "processDelete")
	if err != nil {
		espLogger("processDelete:Invoke:"+workflowID+":"+err.Error(), "error")
		color.Red("processDelete Invoke failed for " + workflowID + ":" + err.Error())
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("processDelete:Unmarshal:"+workflowID+":"+err.Error(), "error")
		color.Red("processDelete Unmarshal failed for " + workflowID + ":" + err.Error())
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("processDelete:MethodResult:"+workflowID+":"+xmlRespon.State.ErrorRet, "error")
		color.Red("processDelete MethodResult failed for " + workflowID + ":" + xmlRespon.State.ErrorRet)
		return
	}
	espLogger("Workflow "+workflowID+" deleted", "notice")
}

//deleteCard - Takes a Card PK ID, sends it to data::entityDelete API for safe deletion
func deleteCard(cardID string) {
	espXmlmc.SetParam("h_card_id", cardID)
	espXmlmc.SetParam("hardDelete", "true")
	browse, err := espXmlmc.Invoke("apps/"+appBM+"/Card", "removeCard")
	if err != nil {
		espLogger("removeCard:Invoke:"+cardID+":"+err.Error(), "error")
		color.Red("removeCard Invoke failed for " + cardID + ":" + err.Error())
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("removeCard:Unmarshal:"+cardID+":"+err.Error(), "error")
		color.Red("removeCard Unmarshal failed for " + cardID + ":" + err.Error())
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("removeCard:MethodResult:"+cardID+":"+xmlRespon.State.ErrorRet, "error")
		color.Red("removeCard MethodResult failed for " + cardID + ":" + xmlRespon.State.ErrorRet)
		return
	}
	espLogger("Board Manager Card "+cardID+" deleted", "notice")
}

func entityDeleteRecords(application, entity string, keyValues []string, preserveOneToOneData, preserveOneToManyData bool) {
	var deletingKeys []string
	espXmlmc.SetParam("application", application)
	espXmlmc.SetParam("entity", entity)
	for _, keyValue := range keyValues {
		espXmlmc.SetParam("keyValue", keyValue)
		deletingKeys = append(deletingKeys, keyValue)
	}
	logKeys := strings.Join(deletingKeys[:], ";")
	espXmlmc.SetParam("preserveOneToOneData", strconv.FormatBool(preserveOneToOneData))
	espXmlmc.SetParam("preserveOneToManyData", strconv.FormatBool(preserveOneToManyData))
	browse, err := espXmlmc.Invoke("data", "entityDeleteRecord")
	if err != nil {
		espLogger("entityDeleteRecord:Invoke:"+application+":"+entity+":"+logKeys+":"+err.Error(), "error")
		color.Red("entityDeleteRecord Invoke failed for " + application + ":" + entity + ":" + logKeys + ":" + err.Error())
		return
	}
	var xmlRespon xmlmcResponse
	err = xml.Unmarshal([]byte(browse), &xmlRespon)
	if err != nil {
		espLogger("entityDeleteRecord:Unmarshal:"+application+":"+entity+":"+logKeys+":"+err.Error(), "error")
		color.Red("entityDeleteRecord Unmarshal failed for " + application + ":" + entity + ":" + logKeys + ":" + err.Error())
		return
	}
	if xmlRespon.MethodResult != "ok" {
		espLogger("entityDeleteRecord:MethodResult:"+application+":"+entity+":"+logKeys+":"+xmlRespon.State.ErrorRet, "error")
		color.Red("entityDeleteRecord MethodResult failed for " + application + ":" + entity + ":" + logKeys + ":" + xmlRespon.State.ErrorRet)
		return
	}
	for _, val := range deletingKeys {
		espLogger(strings.ToUpper(entity)+" "+val+" deleted", "notice")
	}
}
