package main

import (
	"regexp"

	"github.com/hornbill/goApiLib"
)

const (
	toolVer           = "1.5.0"
	appServiceManager = "com.hornbill.servicemanager"
	datetimeFormat    = "2006-01-02 15:04:05"
)

var (
	cleanerConf     cleanerConfStruct
	configFileName  string
	configAPIKey    string
	configInstance  string
	configBlockSize int
	durationRegex   = regexp.MustCompile(`P[0-9]*D[0-9]*H[0-9]*M[0-9]*S`)
	maxResults      int
	resetResults    bool
	currentBlock    int
	totalBlocks     int
	espXmlmc        *apiLib.XmlmcInstStruct
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
