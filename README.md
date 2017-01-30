# IP2Location Go Package (Idiomatic Version)

## What's Different?

Compared to upstream, this fork has a number of changes we feel were necessary:

- No global state.
- Proper error handling (errors are not silently ignored).
- camelCase, instead of snake_case.
- Basically, idiomatic Go (see example at the end).

## Upstream Description

This Go package provides a fast lookup of country, region, city, latitude, longitude, ZIP code, time zone, ISP, domain name, connection type, IDD code, area code, weather station code, station name, mcc, mnc, mobile brand, elevation, and usage type from IP address by using IP2Location database. This package uses a file based database available at IP2Location.com. This database simply contains IP blocks as keys, and other information such as country, region, city, latitude, longitude, ZIP code, time zone, ISP, domain name, connection type, IDD code, area code, weather station code, station name, mcc, mnc, mobile brand, elevation, and usage type as values. It supports both IP address in IPv4 and IPv6.

This package can be used in many types of projects such as:

 - select the geographically closest mirror
 - analyze your web server logs to determine the countries of your visitors
 - credit card fraud detection
 - software export controls
 - display native language and currency
 - prevent password sharing and abuse of service
 - geotargeting in advertisement

The database will be updated in monthly basis for the greater accuracy. Free LITE databases are available at https://lite.ip2location.com/ upon registration.

The paid databases are available at https://www.ip2location.com under Premium subscription package.


## Installation

```
go get github.com/getsocial-rnd/ip2location-go
```

## Example

```go
package main

import (
	"fmt"
	"github.com/getsocial-rnd/ip2location-go"
)

func main() {
	db, err := ip2location.Open("/path/to/db.bin")
	if err != nil {
        // handle error
	}

	results := db.GetAll(ip)
	fmt.Println(results.String())
	db.Close()
}
```

## Dependencies

The complete database is available at http://www.ip2location.com under subscription package.

## Copyright

Copyright (C) 2016 by IP2Location.com, support@ip2location.com
