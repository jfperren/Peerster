package main

import (
  "fmt"
  "flag"
)


func main () {

  // Define Flags

  uiPort := flag.String("UIPort", "8080", "port for the UI client")
  message := flag.String("msg", "127.0.0.1:5000", "message to be sent")

  flag.Parse()

  // Print Flags

  fmt.Printf("port = %v\n", *uiPort)
  fmt.Printf("message = %v\n", *message)
}
