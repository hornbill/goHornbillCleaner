package main

import (
	"regexp"

	apiLib "github.com/hornbill/goApiLib"
)

const (
	version               = "1.11.0"
	appSM                 = "com.hornbill.servicemanager"
	appBM                 = "com.hornbill.boardmanager"
	appCM                 = "com.hornbill.configurationmanager"
	datetimeFormat        = "2006-01-02 15:04:05"
	minBoardManagerBuild  = 100
	minConfigManagerBuild = 10
)

var (
	boardManagerInstalled  = false
	configManagerInstalled = false
	cleanerConf            cleanerConfStruct
	configFileName         string
	configAPIKey           string
	configInstance         string
	configBlockSize        int
	configDryRun           bool
	configSkipPrompts      bool
	configVersion          bool
	durationRegex          = regexp.MustCompile(`P[0-9]*D[0-9]*H[0-9]*M[0-9]*S`)
	currentBlock           int
	totalBlocks            int
	espXmlmc               *apiLib.XmlmcInstStruct
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
	SessionID   string       `xml:"sessionId"`
	RecordIDs   []dataStruct `xml:"rowData>row"`
	BPMID       string       `xml:"primaryEntityData>record>h_bpm_id"`
	RecordCount int          `xml:"count"`
	MaxResults  int          `xml:"option>value"`
	Application []appsStruct `xml:"application"`
}

type dataStruct struct {
	RequestID         string `xml:"h_pk_reference"`
	LogDate           string `xml:"h_datelogged"`
	CloseDate         string `xml:"h_dateclosed"`
	AssetID           string `xml:"h_pk_asset_id"`
	AssetName         string `xml:"asset_name"`
	AssetLinkID       string `xml:"h_pk_id"`
	AssetImpactID     string `xml:"h_pk_confitemimpactid"`
	AssetPolicyID     string `xml:"h_pk_confiteminpolicyid"`
	AssetDependencyID string `xml:"h_pk_confitemdependencyid"`
	BPMEventID        string `xml:"h_fk_eventid"`
	BPMTimerID        string `xml:"h_fk_timerid"`
	CardID            string `xml:"h_id"`
	TimerID           string `xml:"h_pk_tid"`
}

type taskStruct struct {
	TaskID    string `xml:"h_task_id"`
	TaskTitle string `xml:"h_title"`
}

type appsStruct struct {
	Name   string `xml:"name"`
	Status string `xml:"status"`
	Build  int    `xml:"build"`
}

type cleanerConfStruct struct {
	CleanRequests         bool
	RequestServices       []int
	RequestStatuses       []string
	RequestTypes          []string
	RequestLogDateFrom    string
	RequestLogDateTo      string
	RequestClosedDateFrom string
	RequestClosedDateTo   string
	RequestReferences     []string
	CleanAssets           bool
	AssetClassID          string
	AssetTypeID           int
	CleanUsers            bool
	Users                 []string
}

type queryParamsStruct struct {
	Name  string
	Value string
}

type browseRecordsParamsStruct struct {
	Column    string
	Value     string
	MatchType string
}
