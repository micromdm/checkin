[![Build Status](https://travis-ci.org/micromdm/checkin.svg?branch=master)](https://travis-ci.org/micromdm/checkin)
[![GoDoc](https://godoc.org/github.com/micromdm/checkin?status.svg)](http://godoc.org/github.com/micromdm/checkin)

The Checkin Service implements the MDM Check-in protocol.
It responds to device requests sending `Authenticate`, `TokenUpdate` and `CheckOut` commands.

The Checkin Service can be used as both a library and a standalone service.
The current implementation of the checkin service uses [BoltDB](https://github.com/boltdb/bolt#bolt---) to archive events and [NSQ](http://nsq.io/overview/design.html) as the message queue, both of which can be embeded in a larger standalone program. 

# Architecture Diagram
![mdm checkinservice](https://cloud.githubusercontent.com/assets/1526945/20739401/4c4304c2-b688-11e6-97d0-1d369bbc63e7.png)
