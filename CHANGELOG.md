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

## 1.0.4 (January 12th, 2017)

NOTE! Removing requests using this version of the Hornbill Cleaner requires Service Manager v 2.38 or above!

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
