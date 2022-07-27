# Raw bluetooth for Go

This is a raw bluetooth library which connects to a MAC address by using syscall on Linux and Windows.

### Features

- Implements a simple Read/Write interface.
- Same behavior on Linux and Windows (later Darwin).
- Zero-dependency (expect Go's x/sys)

### Usage

```go
package main

import (
	"fmt"
	"github.com/infiniteloopcloud/bluetooth-go"
)

func main() {
	// Connect
	conn, err := bluetooth.Connect(bluetooth.Params{
		Address: "58:CF:0A:BB:28:FC",
	})
	if err != nil {
		// handle err
	}

	// Write data into the connection
	_, err = conn.Write([]byte("data"))
	if err != nil {
		// handle err
	}

	// Read from connection
	_, data, err := conn.Read(500)
	if err != nil {
		// handle err
	}
	fmt.Println(string(data))
	
	// Close connection
	err = conn.Close()
	if err != nil {
		// handle err
	}
}
```

### Implement custom logger

```go
package main

import (
	"fmt"
	"github.com/infiniteloopcloud/bluetooth-go"
)

type Logger struct{}

func (Logger) Print(a ...interface{}) {
	fmt.Println(a...)
}

func main() {
	// Connect
	_, err := bluetooth.Connect(bluetooth.Params{
		Address: "58:CF:0A:BB:28:FC",
		Log: Logger{},
	})
	if err != nil {
		// handle err
	}
	//....
}
```