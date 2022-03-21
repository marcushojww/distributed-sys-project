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
var Cart1 []Item
var Cart2 []Item
var Cart3 []Item

func createNodeServer(name string, port int) *http.Server {

	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)

	// homepage
	myRouter.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprint(res, "Hello: "+name)
	})

	// GET all items
	myRouter.HandleFunc("/items", func(res http.ResponseWriter, req *http.Request) {

		res.Header().Set("Access-Control-Allow-Origin", "*")
		res.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		json.NewEncoder(res).Encode(Items)
	})

	// GET cart depending on which node
	myRouter.HandleFunc("/cart", func(res http.ResponseWriter, req *http.Request) {

		res.Header().Set("Access-Control-Allow-Origin", "*")
		res.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if port == 9001 {
			json.NewEncoder(res).Encode(Cart1)
		} else if port == 9002 {
			json.NewEncoder(res).Encode(Cart2)
		} else if port == 9003 {
			json.NewEncoder(res).Encode(Cart3)
		}
	})

	// GET specific item
	myRouter.HandleFunc("/items/{id}", func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		key := vars["id"]
		fmt.Println(key)

		if port == 9001 {
			for _, item := range Cart1 {
				if item.ID == key {
					json.NewEncoder(w).Encode(item)
				}
			}
		} else if port == 9002 {
			for _, item := range Cart2 {
				if item.ID == key {
					json.NewEncoder(w).Encode(item)
				}
			}
		} else if port == 9003 {
			for _, item := range Cart3 {
				if item.ID == key {
					json.NewEncoder(w).Encode(item)
				}
			}
		}

	})

	// POST new item
	myRouter.HandleFunc("/addToCart", func(w http.ResponseWriter, r *http.Request) {
		// Enable CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		// get the body of our POST request
		// unmarshal this into a new Article struct
		// append this to our Articles array.
		reqBody, _ := ioutil.ReadAll(r.Body)
		var item Item
		json.Unmarshal(reqBody, &item)
		// update our global Articles array to include
		// our new Article
		if port == 9001 {
			Cart1 = append(Cart1, item)
		} else if port == 9002 {
			Cart2 = append(Cart2, item)
		} else if port == 9003 {
			Cart3 = append(Cart3, item)
		}

		json.NewEncoder(w).Encode(item)
	}).Methods("POST", "OPTIONS")

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
		fmt.Fprint(res, "Available commands: \nAdding items to certain user: /addToCart/id \nretrieving cart of certain user: /getCart/id \nRetrieve all items available: /items")

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
		msg := "Cart added successful!"
		json.NewEncoder(res).Encode(msg)
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
			URL = "http://localhost:9001/cart"
		} else if hashedNode == 2 {
			URL = "http://localhost:9002/cart"
		} else if hashedNode == 3 {
			URL = "http://localhost:9003/cart"
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

	Cart1 = []Item{}

	Cart2 = []Item{}

	Cart3 = []Item{}
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
	// wait until WaitGroup is done
	wg.Wait()

}

// {"id":"2","name":"User2","price":"$1.00","desc":"Make your hair look neat with this"}
