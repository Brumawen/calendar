# calendar

Calendar is a microservice that can be used to retrieve upcoming events for one or more calendars.

## Installation

Download the latest release file for your operating system from https://github.com/Brumawen.com/calendar/releases

Extract the files and subfolders to a folder and run the following from the command line

        calendar -service install
        calendar -service start

This will install and run the calendar microservice as a background service on your machine.

If you are using Google Calendars:
* Navigate to https://developers.google.com/calendar/quickstart/go in a browser.
* Click the Enable The Google Calendar API button.
* Create a new project and give it a name.
* Click Next
* Click Download Client Configuration.
* Save the credentials.json file in the same folder as the calendar executable file.

## Configuration

Once the microservice is running, navigate to http://localhost:20513/config.html in a web browser.

To Add a new calendar
* Click the top-left Plus button.
* Enter the Name of the calendar.
* Select the Calender colour.
* Select the Calendar provider from the list.
* Configure the calendar as per the section below.
* Click the Create Calendar button.

To Edit a calendar
* Click the Edit button to the right of the calendar you wish to edit.
* Change the Name of the calendar and/or the Colour.
* Click the Update Calendar button.

To Remove a calendar
* Click the Remove button to the right of the calendar you wish to remove.
* Click the OK button on the confirmation dialog.

### Configuring a Google Calendar

* Click Select Google Calendar button.
* Choose or add a Google Account in the popup window.
* When asked if you want to trust the application, click the Allow button.
* Copy the code text.
* Switch back to the configuration page.
* Paste the code text into the Authentication Code text box.

### Configuring a iCal Public Feed

* Copy and paste the iCal feed URL into the iCal Feed URL text box.
