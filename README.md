# Hornbill Cleaner

The utility provides a quick and easy method of removing requests, assets or user records from a specified Hornbill instance.

## WARNING

This utility permanently deletes request, asset or user records, and records of entities that are associated to the deleted records. It is intended to be used only by an administrator of a Hornbill instance at the appropriate stage of the switch-on process, to remove demonstration and test data prior to go-live.

## Quick Links

- [Hornbill Cleaner](#Hornbill-Cleaner)
  - [WARNING](#WARNING)
  - [Quick Links](#Quick-Links)
  - [Installation](#Installation)
    - [Windows](#Windows)
  - [Configuration](#Configuration)
  - [Execute](#Execute)

## Installation

### Windows

- Download the OS-specific ZIP archive containing the cleaner executable, configuration file and license;
- Extract the ZIP archive into a folder you would like the application to run from e.g. 'C:\hornbill_cleaner\'.

## Configuration

Example JSON File:

```json
{
      "CleanRequests": true,
      "RequestServices":[
          1,
          3,
          9
      ],
      "RequestStatuses":[
          "status.open",
          "status.cancelled",
          "status.closed",
          "status.resolved"
      ],
      "RequestTypes":[
          "Incident",
          "Service Request"
      ],
      "RequestLogDateFrom":"2016-01-01 00:00:00",
      "RequestLogDateTo":"2018-01-01 00:00:00",
      "RequestClosedDateFrom":"2016-01-01 00:00:00",
      "RequestClosedDateTo":"2018-01-01 00:00:00",
      "RequestReferences":[
          "CHR00000021",
          "INC00000003"
      ],
      "CleanAssets": false,
      "AssetClassID": "",
	    "AssetTypeID": 0,
      "CleanUsers":true,
      "Users":[
          "userIdOne",
          "userIdTwo"
      ]
}
```

- CleanRequests : Set to true to remove all Service Manager Requests (and related entity data) from a Hornbill instance. Filter the requests to be deleted using the following parameters:
  - RequestServices : An array containing Service ID Integers to filter the requests for deletion against. An empty array will remove the Service filter, meaning requests with any or no service associated will be deleted
  - RequestStatuses : An array containing Status strings to filter the requests for deletion against. An empty array will remove the Status filter, meaning requests at any status will be deleted
  - RequestTypes : An array containing Request Type strings to filter the requests for deletion against. An empty array will remove the Type filter, meaning requests of any Type will be deleted
  - RequestLogDateFrom :  A date to filter requests against log date (requests logged after or equal to this date/time). Can take one of the following values:
    - An empty string will remove the Logged From filter.
    - A date string in the format YYYY-MM-DD HH:MM:SS.
    - A duration string, to calculate a new datetime from the current datetime:
      - Example: -P1D2H3M4S - This will subtract 1 day (1D), 2 hours (2H), 3 minutes (3H) and 4 seconds (4S) from the current date & time.
      - See the CalculateTimeDuration function documentation in <https://github.com/hornbill/goHornbillHelpers> for more details
  - RequestLogDateTo : A date to filter requests against log date (requests logged before or equal to this date/time). Can take one of the following values:
    - An empty string will remove the Logged Before filter.
    - A date string in the format YYYY-MM-DD HH:MM:SS.
    - A duration string, to calculate a new datetime from the current datetime:
      - Example: -P1D2H3M4S - This will subtract 1 day (1D), 2 hours (2H), 3 minutes (3H) and 4 seconds (4S) from the current date & time.
      - See the CalculateTimeDuration function documentation in <https://github.com/hornbill/goHornbillHelpers> for more details
  - RequestClosedDateFrom :  A date to filter requests against close date (requests closed after or equal to this date/time). Can take one of the following values:
    - An empty string will remove the closed From filter.
    - A date string in the format YYYY-MM-DD HH:MM:SS.
    - A duration string, to calculate a new datetime from the current datetime:
      - Example: -P1D2H3M4S - This will subtract 1 day (1D), 2 hours (2H), 3 minutes (3H) and 4 seconds (4S) from the current date & time.
      - See the CalculateTimeDuration function documentation in <https://github.com/hornbill/goHornbillHelpers> for more details
  - RequestClosedDateTo : A date to filter requests against close date (requests closed before or equal to this date/time). Can take one of the following values:
    - An empty string will remove the Closed Before filter.
    - A date string in the format YYYY-MM-DD HH:MM:SS.
    - A duration string, to calculate a new datetime from the current datetime:
      - Example: -P1D2H3M4S - This will subtract 1 day (1D), 2 hours (2H), 3 minutes (3H) and 4 seconds (4S) from the current date & time.
      - See the CalculateTimeDuration function documentation in <https://github.com/hornbill/goHornbillHelpers> for more details
  - RequestReferences : An array of Request References to delete. If requests are defined in this array, then ONLY these requests will be deleted. The other parameters above will be ignored. In the example above, requests with reference CHR00000021 and INC00000003 would be deleted, and no other requests would be removed.
- CleanAssets : Set to true to remove all Assets (and related entity data) from a Hornbill instance
- AssetClassID: Filter assets for deletetion by a single asset class ID (basic, computer, computerPeripheral, mobileDevice, printer, software, telecoms)
- AssetTypeID: Filter assets for deletion by a single asset type ID - the primary key value of the asset type from the database. Can also be found in the URL when viewing an asset type, 18 in this example: https://live.hornbill.com/yourinstanceid/servicemanager/asset/type/view/18/ 
- CleanUsers : Set to true to remove all Users listed in the Users array
- Users : Array of strings, contains a list of User IDs to remove from your Hornbill instance

## Execute

Command Line Parameters

- instance : This should be the ID of your instance
- apikey : This should be an API of a user on your instance that has the correct rights to perform the search & deletion of the specified records
- file : This should point to your json configration file and by default looks for a file in the current working directory called conf.json. If this is present you don't need to have the parameter.
- blocksize : This allows you to override the default number of records that should be retrieved and deleted as "blocks". The default is 3, and this should only need to be overridden if your Hornbill instance holds large amounts of records to delete, and you experience errors when running the utility.
- dryrun : Requires Service Manager build >= 1392 to work with request data. This boolean flag allows a "dry run" to be performed - the tool identifies the primary key for all parent records that would have been deleted, and outputs them to the log file without deleting any records. Defaults to false.
- justrun : This boolean flag allows you to skip the confirmation prompts when the tool is run. This allows the tool to be scheduled, with the correct configuration defined to delete request records over a certain age for example. Defaults to false.

'hornbillCleaner.exe -instance=yourinstancename -apikey=yourapikey -file=conf.json'

When you are ready to clear-down your request and/or asset records:

- Open '''conf.json''' and add in the necessary configration;
- Open Command Line Prompt as Administrator;
- Change Directory to the folder with hornbillCleaner executable 'C:\hornbill_cleaner\';
- Run the command:
  - On 32 bit Windows PCs: hornbillCleaner.exe -instance=yourinstancename -apikey=yourapikey
  - On 64 bit Windows PCs: hornbillCleaner.exe -instance=yourinstancename -apikey=yourapikey -dryrun=true
- Follow all on-screen prompts, taking careful note of all prompts and messages provided.
