# Customized FOSDEM schedule
A simple application for my custom sorted list of all events in FOSDEM 2021.

The last version saved in this repo can be opened here: https://rawcdn.githack.com/Gys/fosdem/main/fosdem_schedules.html

## Features
* Private: there is no interaction with a server except for (anonymously) loading the page.
* All currently active talks are highlighted. This uses your local time.
* Each talk has a checkbox of which the state is persistent across sessions by keeping them in a cookie. Use these to track your own preferences. The checkboxes state is implemented in pure clientside javascript. 
* All times shown are in Belgium (CET) time.

## Why
I could not find a list of all presentations sorted by time only. Such list makes it easier to see what is presented at any moment in time. I have my favorite talks, but maybe those are not as expected or I want to fill time for which I have nothing planned myself. This list makes it easy to see current alternatives.

To create my list I deciced to parse the online schedule into a new schedule, sorted by time.

The resulting fosdem_schedules.html has more or less the same styling as the original FOSDEM page.

## Code
Make your own list (or process the list otherwise), the code is simple.

First the html of https://fosdem.org/2021/schedule/events/ is loaded. [Goquery]("github.com/PuerkitoBio/goquery") is used to parse the html into a slice with all events. That list is sorted by datetime and finally written to a html file as a table. 
For clientside tracking of preferences https://github.com/js-cookie/js-cookie is used.
## Possible improvements

* Follow each event link to retrieve `#main > div.event-blurb > div.event-abstract` for the event-abstract. To show in dropdown or something.
* Adjust times to local time?
* Add date to start/end for better highlighting.
* Export as csv to easily import in other applications.

