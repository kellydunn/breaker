# breaker

![](https://img1.etsystatic.com/051/0/9689471/il_570xN.666236333_1w1d.jpg)

> Dr. Frederick Frankenstein: Throw... the third switch!
> Igor: [shocked] Not the *third switch*! 

The `breaker` package provides a very simple implementation of the [Circuit Breaker](http://martinfowler.com/bliki/CircuitBreaker.html) design pattern.  Circuit Breakers are a classic pattern used in backend systems, and I just made this out of necessity at work one day.  There's a few other implementations out there, but I figured I'd put mine up on github.

## features

This implementation is relatively barebones; it's intended to be tripped and reset manually, but it also has the ability to "wrap" a function call with a timeout if that's your sort of thing.

## example

```golang
pacakge main

import (
       "github.com/kellydunn/breaker"
       "time"
)

func main() {

     // Lets make a breaker that will trip after 5
     // consecutive errors are counted
     b := breaker.NewBreaker(5)

     // We will now try trigger The breaker
     // once a second to simulate an error.
     // This will exit in about 5 seconds.
     for {
         sleep(time.Second)

         // Lets simulate an error here
         if true {
            b.Trip()
         }
     }
     
     fmt.Printf("Error!: %v", err)

     // This breaker is a bit more forgiving;
     // it'll be thrown after 10 consecutive errors
     b2 := breaker.NewBreaker(10)

     // Here, we'll set up a function that sleeps for
     // 2 seconds, but set a timeout of 1 second.
     // Since it will take 1 second to trip the breaker,
     // This fucntion should exit in about 10 seconds.
     err := b2.Do(func() error {
     
         sleep(2 * time.Second)
         
     }, time.Second)      

     fmt.Printf("Error!: %v", err)
}    
```

Enjoy!

## related work

- [rubyist/circuitbreaker](https://github.com/rubyist/circuitbreaker) - Another go implementation
- [CircuitBreaker](http://martinfowler.com/bliki/CircuitBreaker.html) - Martin Fowler