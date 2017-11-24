# Hornbill Cleaner

The utility provides a quick and easy method of removing all request and/or asset records from a specified Hornbill instance.

## WARNING

This utility permanently deletes request and asset records, and records of entities that are associated to the deletes requests and assets. It is intended to be used only by an administrator of a Hornbill instance at the appropriate stage of the switch-on process, to remove demonstration and test data prior to go-live.

## Quick Links
- [Installation](#installation)
- [Configuration](#configuration)
- [Execute](#execute)

## Installation

#### Windows
* Download the ZIP archive containing the cleaner executable, configuration file and license;
* Extract the ZIP archive into a folder you would like the application to run from e.g. 'C:\hornbill_cleaner\'.

## Configuration

Example JSON File:

```
{
        "UserName": "admin",
        "Password": "password",
        "URL": "https://eurapi.hornbill.com/instancename/xmlmc/"
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
        "RequestReferences":[
                "CHR00000021",
                "INC00000003"
        ],
        "CleanAssets": false
}
```

- UserName: This is the username which will be used to connect to your Hornbill instance. This user should have the appropriate roles associated to it, to be able to remove request and asset entity records.
- Password: This is the password for the supplied username.
- URL : This is the url of the API endpoint of your instance. It must be https, and include /xmlmc/ at the end of the url. You must replace "instancename" with the name of your Hornbill instance.
- CleanRequests : Set to true to remove all Service Manager Requests (and related entity data) from a Hornbill instance. Filter the requests to be deleted using the following parameters:
  - RequestServices : An array containing Service ID Integers to filter the requests for deletion against. An empty array will remove the Service filter, meaning requests with any or no service associated will be deleted
  - RequestStatuses : An array containing Status strings to filter the requests for deletion against. An empty array will remove the Status filter, meaning requests at any status will be deleted
  - RequestTypes : An array containing Request Type strings to filter the requests for deletion against. An empty array will remove the Type filter, meaning requests of any Type will be deleted
  - RequestLogDateFrom : A date string to filter requests against log date (requests logged after or equal to this date/time), in the format YYYY-MM-DD HH:MM:SS. An empty string will remove the Logged After filter. 
  - RequestLogDateTo : A date string to filter requests against log date (requests logged before or equal to this date/time), in the format YYYY-MM-DD HH:MM:SS. An empty string will remove the Logged Before filter. 
  - RequestReferences : An array of Request References to delete. If requests are defined in this array, then ONLY these requests will be deleted. The other parameters above will be ignored. In the example above, requests with reference CHR00000021 and INC00000003 would be deleted, and no other requests would be removed.
- CleanAssets : Set to true to remove all Assets (and related entity data) from a Hornbill instance  

## Execute
Command Line Parameters

- file
This should point to your json configration file and by default looks for a file in the current working directory called conf.json. If this is present you don't need to have the parameter.

'hornbillCleaner_x64.exe -file=conf.json'

- blocksize
This allows you to override the default number of records that should be retrieved and deleted as "blocks". The default is 3, and this should only need to be overridden if your Hornbill instance holds large amounts of records to delete, and you experience errors when running the utility.

'hornbillCleaner_x64.exe -blocksize=1'

When you are ready to clear-down your request and/or asset records:

- Open '''conf.json''' and add in the necessary configration;
- Open Command Line Prompt as Administrator;
- Change Directory to the folder with hornbillCleaner_* executables 'C:\hornbill_cleaner\';
- Run the command: 
        - On 32 bit Windows PCs: hornbillCleaner_x86.exe
        - On 64 bit Windows PCs: hornbillCleaner_x64.exe
- Follow all on-screen prompts, taking careful note of all prompts and messages provided.
