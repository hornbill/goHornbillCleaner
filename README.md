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
        "CleanRequests": false,
        "CleanAssets": false,
        "RequestClass": "Change Request"
}
```

* UserName: This is the username which will be used to connect to your Hornbill instance. This user should have the appropriate roles associated to it, to be able to remove request and asset entity records.
* Password: This is the password for the supplied username.
* URL : This is the url of the API endpoint of your instance. It must be https, and include /xmlmc/ at the end of the url. You must replace "instancename" with the name of your Hornbill instance.
* CleanRequests : Set to true to remove all Service Manager Requests (and related entity data) from a Hornbill instance
* CleanAssets : Set to true to remove all Assets (and related entity data) from a Hornbill instance  
* RequestClass : Specify the class of the requests you wish to delete

## Execute
Command Line Parameters

- file
This should point to your json configration file and by default looks for a file in the current working directory called conf.json. If this is present you don't need to have the parameter.

'hornbillCleaner.exe -file=conf.json'

- blocksize
This allows you to override the default number of records that should be retrieved and deleted as "blocks". The default is 3, and this should only need to be overridden if your Hornbill instance holds large amounts of records to delete, and you experience errors when running the utility.

'hornbillCleaner.exe -blocksize=1'

When you are ready to clear-down your request and/or asset records:

* Open '''conf.json''' and add in the necessary configration;
* Open Command Line Prompt as Administrator;
* Change Directory to the folder with hornbillCleaner.exe 'C:\hornbill_cleaner\';
* Run the command: hornbillCleaner.exe
* Follow all on-screen prompts, taking careful note of all prompts and messages provided.
