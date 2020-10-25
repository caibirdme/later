### Later
This library is an implementation of `Hierarchical Timing Wheels`. To learn more detailed information about it, see [here](http://www.cs.columbia.edu/~nahum/w6998/papers/ton97-timing-wheels.pdf).

### Install
`go get -u github.com/caibirdme/later`

### Quick Start
```go
package main

import (
 "github.com/caibirdme/later"
 "time"
 "fmt"
)

func main() {
    timer := later.NewSecondTimeWheel()
    timer.Start() // start the timer

    timer.After(10*time.Second, func() {
        fmt.Println("print this 10s later")            
    })
    timer.Every(2*time.Second, func() {
        fmt.Println("print this every 2s")
    })
    // ...
    timer.Stop() // stop the timer
}
```
