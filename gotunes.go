package main

import (
  "fmt"

  "github.com/bernhardfritz/gotunes/greeter"
  "rsc.io/quote"
)

func main() {
  fmt.Println(greeter.Greet("Bernhard"))
  fmt.Println(quote.Hello())
}
