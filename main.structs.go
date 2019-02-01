package main

import (
	"regexp"

	"github.com/hornbill/goApiLib"
)

const (
	toolVer              = "1.8.2"
	appServiceManager    = "com.hornbill.servicemanager"
	datetimeFormat       = "2006-01-02 15:04:05"
	minBoardManagerBuild = 100
)

var (
	boardManagerInstalled = false
	cleanerConf           cleanerConfStruct
	configFileName        string
	configAPIKey          string
	configInstance        string
	configBlockSize       int
	configDryRun          bool
	configSkipPrompts     bool
	durationRegex         = regexp.MustCompile(`P[0-9]*D[0-9]*H[0-9]*M[0-9]*S`)
	maxResults            int
	resetResults          bool
	currentBlock          int
	totalBlocks           int
	espXmlmc              *apiLib.XmlmcInstStruct
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
	RequestID   string `xml:"h_pk_reference"`
	LogDate     string `xml:"h_datelogged"`
	CloseDate   string `xml:"h_dateclosed"`
	AssetID     string `xml:"h_pk_asset_id"`
	AssetName   string `xml:"asset_name"`
	AssetLinkID string `xml:"h_pk_id"`
	CardID      string `xml:"rowData>row>h_id"`
	TimerID     string `xml:"rowData>row>h_pk_tid"`
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

type workflowStruct struct {
	WorkflowID string
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
	CleanUsers            bool
	Users                 []string
}
