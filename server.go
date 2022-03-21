package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Item struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price string `json:"price"`
	Desc  string `json:"desc"`
}

type User struct {
	ID            string `json:"id"`
	Shopping_Cart string `json:"shopping_cart"`
	Purchase_List string `json:"purchase_list"`
}

type Node struct {
	id         int
	httpServer *http.Server
	port       int
}

type RingServer struct {
	nodeArray  []Node
	httpServer *http.Server
}

const NUM_NODES = 3

var Items []Item
var Items1 []Item
var Items2 []Item
var Items3 []Item

func createNodeServer(name string, port int) *http.Server {

	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)

	// homepage
	myRouter.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprint(res, "Hello: "+name)
	})

	// getting all items
	myRouter.HandleFunc("/items", func(res http.ResponseWriter, req *http.Request) {
		if port == 9001 {
			json.NewEncoder(res).Encode(Items1)
		} else if port == 9002 {
			json.NewEncoder(res).Encode(Items2)
		} else if port == 9003 {
			json.NewEncoder(res).Encode(Items3)
		}

	})

	// getting specific item
	myRouter.HandleFunc("/items/{id}", func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		key := vars["id"]
		fmt.Println(key)

		if port == 9001 {
			for _, item := range Items1 {
				if item.ID == key {
					json.NewEncoder(w).Encode(item)
				}
			}
		} else if port == 9002 {
			for _, item := range Items2 {
				if item.ID == key {
					json.NewEncoder(w).Encode(item)
				}
			}
		} else if port == 9003 {
			for _, item := range Items3 {
				if item.ID == key {
					json.NewEncoder(w).Encode(item)
				}
			}
		}

	})

	// Create new item
	myRouter.HandleFunc("/addToCart", func(w http.ResponseWriter, r *http.Request) {
		// get the body of our POST request
		// unmarshal this into a new Article struct
		// append this to our Articles array.
		reqBody, _ := ioutil.ReadAll(r.Body)
		var item Item
		json.Unmarshal(reqBody, &item)
		// update our global Articles array to include
		// our new Article
		if port == 9001 {
			Items1 = append(Items1, item)
		} else if port == 9002 {
			Items2 = append(Items2, item)
		} else if port == 9003 {
			Items3 = append(Items3, item)
		}

		json.NewEncoder(w).Encode(item)
	}).Methods("POST")

	// create new server
	server := http.Server{
		Addr:    fmt.Sprintf(":%v", port), // :{port}
		Handler: myRouter,
	}

	// return new server (pointer)
	return &server
}

func createRingServer(name string, port int) *http.Server {

	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)

	// Homepage of Ring Server
	myRouter.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprint(res, "Hello: "+name)
	})
	myRouter.HandleFunc("/items", func(res http.ResponseWriter, req *http.Request) {
		json.NewEncoder(res).Encode(Items)
	})

	myRouter.HandleFunc("/addToCart/{id}", func(res http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		key := vars["id"]
		intKey, _ := strconv.Atoi(key)
		var hashedNode int
		hashedNode = (intKey % 3) + 1
		var URL string
		if hashedNode == 1 {
			URL = "http://localhost:9001/addToCart"
		} else if hashedNode == 2 {
			URL = "http://localhost:9002/addToCart"
		} else if hashedNode == 3 {
			URL = "http://localhost:9003/addToCart"
		}

		reqBody, _ := ioutil.ReadAll(req.Body)
		responseBody := bytes.NewBuffer(reqBody)
		//Leverage Go's HTTP Post function to make request
		resp, err := http.Post(URL, "application/json", responseBody)
		//Handle Error
		if err != nil {
			log.Fatalf("An Error Occured %v", err)
		}

		defer resp.Body.Close()
		//Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		sb := string(body)
		log.Printf(sb)
	})

	myRouter.HandleFunc("/getCart/{id}", func(res http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		key := vars["id"]
		intKey, _ := strconv.Atoi(key)
		var hashedNode int
		hashedNode = (intKey % 3) + 1
		var URL string
		if hashedNode == 1 {
			URL = "http://localhost:9001/items"
		} else if hashedNode == 2 {
			URL = "http://localhost:9002/items"
		} else if hashedNode == 3 {
			URL = "http://localhost:9003/items"
		}
		resp, err := http.Get(URL)
		if err != nil {
			log.Fatalln(err)
		}
		//We Read the response body on the line below.
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		var item []Item
		json.Unmarshal(body, &item)
		json.NewEncoder(res).Encode(item)
	})

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", port), // :{port}
		Handler: myRouter,
	}

	// return new server (pointer)
	return &server
}

func main() {
	// Initialize ring server
	nodeArray := []Node{}
	rServer := createRingServer("Node "+strconv.Itoa(0), 9000)
	ringServer := RingServer{nodeArray, rServer} // ring server is port 9000

	Items = []Item{
		{ID: "1", Name: "Comb", Desc: "Make your hair look neat with this", Price: "$1.00"},
		{ID: "2", Name: "Pokka Green Tea", Desc: "Jasmine green tea", Price: "$2.00"},
		{ID: "3", Name: "Teddy Bear", Desc: "Plushy toy", Price: "$10.00"},
	}

	Items1 = []Item{
		{ID: "1", Name: "Comb", Desc: "Make your hair look neat with this", Price: "$1.00"},
		{ID: "2", Name: "Pokka Green Tea", Desc: "Jasmine green tea", Price: "$2.00"},
		{ID: "3", Name: "Teddy Bear", Desc: "Plushy toy", Price: "$10.00"},
	}

	Items2 = []Item{
		{ID: "1", Name: "Comb", Desc: "Make your hair look neat with this", Price: "$1.00"},
		{ID: "2", Name: "Pokka Green Tea", Desc: "Jasmine green tea", Price: "$2.00"},
		{ID: "3", Name: "Teddy Bear", Desc: "Plushy toy", Price: "$10.00"},
	}

	Items3 = []Item{
		{ID: "1", Name: "Comb", Desc: "Make your hair look neat with this", Price: "$1.00"},
		{ID: "2", Name: "Pokka Green Tea", Desc: "Jasmine green tea", Price: "$2.00"},
		{ID: "3", Name: "Teddy Bear", Desc: "Plushy toy", Price: "$10.00"},
	}
	// create a WaitGroup
	wg := new(sync.WaitGroup)

	// add two goroutines to `wg` WaitGroup
	wg.Add(NUM_NODES)

	for i := 1; i <= NUM_NODES; i++ {
		port := 9000 + i
		server := createNodeServer("Node "+strconv.Itoa(i), port)
		n := Node{i, server, port}
		// Append node to ring server array
		ringServer.nodeArray = append(ringServer.nodeArray, n)
	}
	fmt.Println("Completed ring server array:", ringServer.nodeArray)
	// Activate gorouter to launch servers
	for _, n := range ringServer.nodeArray {
		fmt.Println("Server", n.id, "started on port", n.port, ". HTTP Server:", n.httpServer)
		go n.httpServer.ListenAndServe()
	}
	go ringServer.httpServer.ListenAndServe()
	// // goroutine to launch a server on port 9000
	// go func() {
	//     server := createServer( "Node 1", 9001 )
	// 	n := Node{ 1, server, 9001}
	// 	fmt.Println("Server", n.id, " started on port ", n.port,". HTTP Server: ", n.httpServer)
	//     fmt.Println( server.ListenAndServe() )
	//     wg.Done()
	// }()

	// // // goroutine to launch a server on port 9001
	// go func() {
	//     server := createServer( "Node 2", 9002 )
	// 	n := Node{ 2, server, 9002}
	// 	fmt.Println("Server", n.id, " started on port ", n.port,". HTTP Server: ", n.httpServer)
	//     fmt.Println( server.ListenAndServe() )
	//     wg.Done()
	// }()

	//  // // goroutine to launch a server on port 9002
	//  go func() {
	//     server := createServer( "Node 3", 9003 )
	// 	n := Node{ 3, server, 9003}
	// 	fmt.Println("Server", n.id, " started on port ", n.port,". HTTP Server: ", n.httpServer)
	//     fmt.Println( server.ListenAndServe() )
	//     wg.Done()
	// }()

	// wait until WaitGroup is done
	wg.Wait()

}
