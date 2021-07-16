# CHANGELOG

## 1.17.0 (July 16th, 2021)

Features:

- Added support to delete Supplier records from Supplier Manager 
- Added support to delete Supplier Contract records from Supplier Manager 

## 1.16.2 (July 6th, 2021)

Change:

- Rebuilt using latest version of goApiLib, to fix possible issue witt connections via a proxy

## 1.16.1 (June 30th, 2021)

Fix:

- Issue returning count of Service Status History records

## 1.16.0 (June 17th 2021)

Changes:

- Updates to cater for move of Configuration Manager into Service Manager

Features:

- Added support to filter Requests for deletion by Catalog ID
- Added support to delete Service Availability records, filtered by Service ID(s)
- Added support to delete Contact records
- Added support to delete Organization records
- Added version check against Github repo 
- Enhanced debug logging, now logs request and response payloads when an API call into Hornbill fails

## 1.15.1 (May 6th, 2021)

Fix:

- Tweaked a little code in order to prevent a non-entity from being deleted.

## 1.15.0 (April 26th, 2021)

Change:

- Added ability to cancel request tasks and BP - see KeepRequestsCancelBPTasks.

## 1.14.0 (March 10th, 2021)

Change:

- Removal of unsupported unicode characters from response XML

## 1.13.1 (March 4th, 2021)

Change:

- Pushed CLI errors to instance log to aid in debugging

## 1.13.0 (March 1st, 2021)

Change:

- Additional logging around returning records for deletion
- Added string sanitizer to remove illegal character codes

## 1.12.1 (February 24th, 2021)

Change:

- Upped the request timeout from 30 to 180 seconds

## 1.12.0 (October 16th, 2020)

Change:

- Added support to provide complex filters for deletion of assets

## 1.11.2 (September 2nd, 2020)

Fix:

- Fixed issue where request tasks were not being deleted

## 1.11.1 (February 12th, 2020)

Change:

- Refactored system & BPM event timer delete order

## 1.11.0 (September 27th, 2019)

Features:

- Added support to clean-up Configuration Manager Dependency, Impact and Policy records when Assets are deleted
- Refactored to remove duplicate code
- Improved log output consistency

## 1.10.0 (July 12th, 2019)

Features:

- Added support to filter assets for deletion by class and type

## 1.9.0 (March 15th, 2019)

Features:

- Replaced large data set calls to entityBrowseRecords2 with paginated queries
- Removed calls to sysOptionGet and sysOptionSet
- Added support to remove additional Request BPM timers and events
- Added support to remove Request SLM events

## 1.8.2 (February 1st, 2019)

Defect Fix:

- Fixed issue with AssetsLinks records not being cleared

## 1.8.1 (January 28th, 2019)

Defect Fix:

- Fixed issue with RequestClosedDateFrom and RequestClosedDateTo returning a query error

## 1.8.0 (December 5th, 2018)

Features:

- Added new filters for requests for deletion:
  - RequestClosedDateFrom : Delete requests that were closed on or after this date/time
  - RequestClosedDateTo: Delete requests that were closed before or at this date/time

## 1.7.0 (November 29th, 2018)

Features:

- Added a "dryrun" CLI input parameter, which when enabled allows users to run the tool without deleting any records, and the primary key for each record that would have been deleted is output to logs for review. Requires Service Manager build 1392 or above to work with request data
- Added a "justrun" CLI input parameter, which when enabled will skip the initial "do you want to delete the data" prompts. This allows the tool to be run on a schedule
- Improved logging output:
  - Logs everything on both client and server side now, instead of just server side, to aid in the reviewing of logs
  - At the start of the log, all configuration options (the CLI params AND all options & filters from the config JSON) are now logged for auditing purposes

## 1.6.0 (November 1st, 2018)

Features:

- Added support to delete Board Manager cards when parent Requests are deleted if Board Manager build >= 100 is present
- Improved logging output

## 1.5.1 (October 10th, 2018)

Features:

- Outputs relevant error message if instance is not found
- Improved performance and better sharing of HTTP sessions

Defect fixes:

- Fixed memory leak

## 1.5.0 (September 9th, 2018)

Features:

- Added support for supplying a duration string, as well as the existing hard-coded datetime string in the RequestLogDateFrom and RequestLogDateTo configuration parameters. This allows for the calculation of datetimes using the runtime datetime.
- General tidy-up of the code, split code in to separate Go files for ease of maintenance

## 1.4.0 (June 4th, 2018)

Feature:

- Replaced Username & Password session authentication with API key
- Replaced stored username, password and instance URL with command line inputs for instance ID and API Key

## 1.3.0 (February 1st, 2018)

Feature:

- When requests are being deleted, any asset links records are now also deleted.
- Added ability to delete User records

## 1.2.0 (November 24th, 2017)

Feature:

- Added ability to delete specific requests using their reference numbers.

## 1.1.0 (September 1st, 2017)

Feature:

- Requests to be deleted can now be filtered by:
  - Multiple Service IDs
  - Multiple Statuses
  - Multiple Types
  - Requests logged after a specific date & time
  - Requests logged before a specific date & time
- NOTE - this version requires Hornbill Service Manager Update 1048 or above.

## 1.0.6 (July 17th, 2017)

Feature:

- Now supports the deletion of Asset CMDB links when clearing down asset records

## 1.0.5 (February 1st, 2017)

Defect fix:

- Changed the order in which request extended information is deleted, so that workflow tasks can be deleted successfully

## 1.0.4 (January 12th, 2017)

NOTE! Removing requests using this version of the Hornbill Cleaner requires Service Manager v 2.38 or above!

Features:

- Added code to process the deletion of:
  - Request Workflow instances
  - Request Activities
  - Request timer events

## 1.0.3 (August 3rd, 2016)

Features:

- Added parameter within configuration file, to specify class of requests to delete  
- Improved performance of request deletion
- Improved error output to display

## 1.0.2 (June 8th, 2016)

Features:

- Reduced record block size default down to 3
- Improved logging output

## 1.0.1 (May 12, 2016)

Features:

- Reduced record block size default down to 20
- Added flag to allow default block-size to be overridden at runtime

## 1.0.0 (March 10, 2016)

Features:

- Initial Release
