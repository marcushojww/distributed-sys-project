package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type Item struct {
	Name string `json:"name"`
	Price string `json:"price"`
	Desc string `json:"desc"`
}

type Node struct {
	id int
	httpServer http.Server
}

var Items []Item

func createServer( name string, port int ) *http.Server {

    // // create `ServerMux`
    // mux := http.NewServeMux()

    // // create a default route handler
    // mux.HandleFunc( "/", func( res http.ResponseWriter, req *http.Request ) {
    //     fmt.Fprint( res, "Hello: " + name )
    // } )

	// mux.HandleFunc("/items", func( res http.ResponseWriter, req *http.Request ) {
    //     json.NewEncoder(res).Encode(Items)
    // })

	// creates a new instance of a mux router
    myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc( "/", func( res http.ResponseWriter, req *http.Request ) {
        fmt.Fprint( res, "Hello: " + name )
    } )

	myRouter.HandleFunc("/items", func( res http.ResponseWriter, req *http.Request ) {
        json.NewEncoder(res).Encode(Items)
    })

    // create new server
    server := http.Server {
        Addr: fmt.Sprintf( ":%v", port ), // :{port}
        Handler: myRouter,
    }

    // return new server (pointer)
    return &server
}

func main() {

	Items = []Item{
		Item{Name: "Comb", Desc:"Make your hair look neat with this", Price:"$1.00"},
		Item{Name:"Pokka Green Tea", Desc:"Jasmine green tea", Price:"$2.00"},
	}
    // create a WaitGroup
    wg := new(sync.WaitGroup)

    // add two goroutines to `wg` WaitGroup
    wg.Add(2)

    // goroutine to launch a server on port 9000
    go func() {
        server := createServer( "ONE", 9000 )
        fmt.Println( server.ListenAndServe() )
        wg.Done()
    }()

    // goroutine to launch a server on port 9001
    go func() {
        server := createServer( "TWO", 9001 )
        fmt.Println( server.ListenAndServe() )
        wg.Done()
    }()

    // wait until WaitGroup is done
    wg.Wait()

}