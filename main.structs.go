package main

import (
	"regexp"

	apiLib "github.com/hornbill/goApiLib"
)

const (
	version              = "1.22.0"
	appName              = "goHornbillCleaner"
	appSM                = "com.hornbill.servicemanager"
	appSuppM             = "com.hornbill.suppliermanager"
	appBM                = "com.hornbill.boardmanager"
	appCore              = "com.hornbill.core"
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
	configVersion         bool
	currentBlock          int
	displayBlock          int
	totalBlocks           int
	durationRegex         = regexp.MustCompile(`[-]{0,1}P[0-9]*D[0-9]*H[0-9]*M[0-9]*S`)
	espXmlmc              *apiLib.XmlmcInstStruct

	assetsDeleted []string

	mailRecipientClasses = map[string]int{
		"unknown":         0,
		"to":              1,
		"cc":              2,
		"bcc":             3,
		"from":            4,
		"replyTo":         5,
		"returnReceiptTo": 6,
	}
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
	SessionID        string       `xml:"sessionId"`
	RecordIDs        []dataStruct `xml:"rowData>row"`
	BPMID            string       `xml:"primaryEntityData>record>h_bpm_id"`
	RecordCount      int          `xml:"count"`
	MaxResults       int          `xml:"option>value"`
	Application      []appsStruct `xml:"application"`
	ChatSessionCount int          `xml:"chatSessionCount"`
	Outcome          string       `xml:"outcome"`
	TotalRecordCount int          `xml:"totalRecordCount"`
	ChatSessions     string       `xml:"chatSessions"` // This is a string type property containing a JSON array of chat sessions
}

type dataStruct struct {
	RequestID         string `xml:"h_pk_reference"`
	LogDate           string `xml:"h_datelogged"`
	CloseDate         string `xml:"h_dateclosed"`
	AssetID           string `xml:"h_pk_asset_id"`
	AssetName         string `xml:"asset_name"`
	AssetLinkID       string `xml:"h_pk_id"`
	AssetLinkLID      string `xml:"h_fk_id_l"`
	AssetLinkRID      string `xml:"h_fk_id_r"`
	AssetImpactID     string `xml:"h_pk_confitemimpactid"`
	AssetPolicyID     string `xml:"h_pk_confiteminpolicyid"`
	AssetDependencyID string `xml:"h_pk_confitemdependencyid"`
	BPMEventID        string `xml:"h_fk_eventid"`
	BPMTimerID        string `xml:"h_fk_timerid"`
	HID               string `xml:"h_id"`
	TimerID           string `xml:"h_pk_tid"`
	Count             int    `xml:"count"`
	ContactID         int
	OrgID             int
	SuppID            int
	SuppConID         string
	MessageCount      int    `xml:"h_count"`
	MessageID         int    `xml:"h_msg_id"`
	MessageDate       string `xml:"h_msg_date"`
	ChatSessionID     string // from the string JSON array returned by getChatSessions
}

type chatSessionObject struct {
	SessionID string `json:"h_session_id"`
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
	CleanRequests                   bool
	KeepRequestsCancelBPTasks       bool
	RequestServices                 []int
	RequestCatalogItems             []int
	RequestStatuses                 []string
	RequestTypes                    []string
	RequestLogDateFrom              string
	RequestLogDateTo                string
	RequestClosedDateFrom           string
	RequestClosedDateTo             string
	RequestReferences               []string
	CleanAssets                     bool
	AssetClassID                    string
	AssetFilters                    []assetFilterStuct
	AssetIDs                        []string
	CleanUsers                      bool
	Users                           []string
	CleanServiceAvailabilityHistory bool
	ServiceAvailabilityServiceIDs   []int
	CleanContacts                   bool
	ContactIDs                      []int
	CleanOrganisations              bool
	OrganisationIDs                 []int
	CleanSuppliers                  bool
	SupplierIDs                     []int
	CleanSupplierContracts          bool
	SupplierContractIDs             []string
	CleanEmails                     bool
	EmailFilters                    struct {
		RecipientAddress string
		RecipientClass   string
		FolderIDs        []int
		ReceivedFrom     string
		ReceivedTo       string
		Subject          string
	}
	CleanReports      bool
	ReportIDs         []int
	CleanChatSessions bool
	ChatSessionIDs    []string
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

type assetFilterStuct struct {
	ColumnName        string
	ColumnValue       string
	Operator          string
	IsGeneralProperty bool
}

type filterStuct struct {
	ColumnName        string `json:"column_name"`
	ColumnValue       string `json:"column_value"`
	Operator          string `json:"operator"`
	IsGeneralProperty bool   `json:"isGeneralProperty"`
}
