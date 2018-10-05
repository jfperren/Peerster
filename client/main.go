package main

import (
  "fmt"
  "flag"
)

func main () {

  // Define Flags

  uiPort := flag.String("UIPort", "8080", "port for the UI client")
  message := flag.String("msg", "Test message", "message to be sent")

  flag.Parse()

  // Print Flags

  fmt.Printf("port = %v\n", *uiPort)
  fmt.Printf("message = %v\n", *message)
}
