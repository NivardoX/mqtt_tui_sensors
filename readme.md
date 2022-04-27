# MQTT TUI Sensors
This a project involving mqtt with a Terminal User Interface.
## Installing

---

You'll need a shell and `go`

- simply start the mqtt broker 
`docker-compose up`
- `go run .`

## How to run

---

For every instance, one mock sensor is created. 
You'll need to inform a identifier and threshold values. 
> The limits will be used to mock sensor reads with an interval of `[min-10,max+10]`

As this is terminal based, to exit simply press `ctrl+c`

Author: `Nivardo Albuquerque Leitão de Castro` Find me at `nivardo00@gmail.com` or `github.com/nivardox`.