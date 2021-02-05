# Customized FOSDEM schedule
A simple application for my custom sorted list of all events in FOSDEM 2021.

I could not find a list of all presentations sorted by time only. Such list makes it easier to see what is presented at any moment in time. I have my favorite talks, but maybe those are not as expected or I want to fill time for which I have nothing planned myself. This list makes it easy to see current alternatives.

To create my list I deciced to parse the online schedule into a new schedule, sorted by time.

If you also want such schedule then just download the html file and type `open fosdem_schedule.html`. Times are in local Brussels time.

The resulting fosdem_schedule.html has more or less the same styling as the original FOSDEM page.

If you want to make your own list, the code is simple:

First the html of https://fosdem.org/2021/schedule/events/ is loaded. [Goquery]("github.com/PuerkitoBio/goquery") is used to parse the html into a list of all events. That list is sorted by datetime and finally written to a html file as a table. 