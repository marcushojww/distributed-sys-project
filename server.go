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
	"time"

	"github.com/gorilla/mux"
)

type Item struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price string `json:"price"`
	Desc  string `json:"desc"`
	Img string `json:"img"`
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
	nodeArray  []Node
	successors []Node
}

type RingServer struct {
	nodeArray  []Node
	httpServer *http.Server
}

const NUM_NODES = 5
const REPLICATION_FACTOR = 3

type BackupRingServer struct {
	nodeArray []Node
	httpServer *http.Server
}

var Items []Item
var Cart1 []Item
var Cart2 []Item
var Cart3 []Item
var Cart4 []Item
var Cart5 []Item

var isRingServerDown bool

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
		} else if port == 9004 {
			json.NewEncoder(res).Encode(Cart4)
		} else if port == 9005 {
			json.NewEncoder(res).Encode(Cart5)
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
		} else if port == 9004 {
			for _, item := range Cart4 {
				if item.ID == key {
					json.NewEncoder(w).Encode(item)
				}
			}
		} else if port == 9005 {
			for _, item := range Cart5 {
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
		} else if port == 9004 {
			Cart4 = append(Cart4, item)
		} else if port == 9005 {
			Cart5 = append(Cart5, item)
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
		// Enable CORS
		res.Header().Set("Access-Control-Allow-Origin", "*")
		res.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		json.NewEncoder(res).Encode(Items)
	})

	myRouter.HandleFunc("/addToCart/{uid}", func(res http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		key := vars["uid"]
		intKey, _ := strconv.Atoi(key)
		var hashedNode int
		hashedNode = (intKey % 5) + 1
		intHash := hashedNode
		
		// var responseBody *bytes.Buffer
		reqBody, _ := ioutil.ReadAll(req.Body)
		// var responseBodyCopy *bytes.Buffer
		// req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
		fmt.Println(req.Body)
		var URL string

		for i := 0; i < REPLICATION_FACTOR; i++ {
			responseBody := bytes.NewBuffer(reqBody)
			fmt.Println(responseBody)
			intHash = hashedNode + i
			stringedHash := strconv.Itoa(intHash)
			URL = "http://localhost:900" + stringedHash + "/addToCart"
			fmt.Println(URL)

			//Leverage Go's HTTP Post function to make request
			resp, err := http.Post(URL, "application/json", responseBody)
			//Handle Error
			if err != nil {
				log.Fatalf("An Error Occured %v", err)
			}
			defer resp.Body.Close()
			//Read the response body
			reqBody, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}
		}
		// 
		item := Item{}
		error := json.Unmarshal(reqBody, &item)
		json.NewEncoder(res).Encode(item)
		if error != nil {
			log.Fatalln(error)
		}
		
	})

	myRouter.HandleFunc("/getCart/{id}", func(res http.ResponseWriter, req *http.Request) {
		fmt.Println(req)
		vars := mux.Vars(req)
		key := vars["id"]
		intKey, _ := strconv.Atoi(key)
		var hashedNode int
		hashedNode = (intKey % 5) + 1
		stringedHash := strconv.Itoa(hashedNode)
		var URL string
		URL = "http://localhost:900" + stringedHash + "/cart"
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

func pingRingServer(backupRingServer BackupRingServer) {
	for range time.Tick(time.Second * 2) {

		fmt.Println("Is Ring Server Down:", isRingServerDown)

		resp, err := http.Get("http://localhost:9000")

		// If cannot reach primary ring server
		if err != nil {
			// If global var is not yet updated
			if !isRingServerDown {
				fmt.Println("Primary Ring Server down. Starting Backup Ring Server...")
				backupHttpServer := createRingServer("Node "+strconv.Itoa(0), 9000)
				backupRingServer.httpServer = backupHttpServer
				fmt.Println("Backup Ring Server listening at port 9000")
				go backupRingServer.httpServer.ListenAndServe()

				// Update global var
				isRingServerDown = true
			}

		}
		fmt.Println("Server response at port 9000: ", resp)

	}
}

func main() {

	// Initialize ring server
	nodeArray := []Node{}
	successors := []Node{}
	rServer := createRingServer("Node "+strconv.Itoa(0), 9000)
	ringServer := RingServer{nodeArray, rServer} // ring server is port 9000

	// Initialize back up ring server
	// backupRServer := createRingServer("Node "+strconv.Itoa(0), 8999)
	backupRingServer := BackupRingServer{}

	Items = []Item{
		{ID: "1", Name: "Comb", Desc: "Make your hair look neat with this", Price: "$1.00", Img: "https://m.media-amazon.com/images/I/71WmBY-nquL.jpg"},
		{ID: "2", Name: "Pokka Green Tea", Desc: "Jasmine green tea", Price: "$2.00", Img: "https://coldstorage-s3.dexecure.net/product/056471_1528887337809.jpg"},
		{ID: "3", Name: "Teddy Bear", Desc: "Plushy toy", Price: "$10.00", Img: "https://nationaltoday.com/wp-content/uploads/2021/08/Teddy-Bear-Day.jpg"},
		{ID: "4", Name: "Soccer ball", Desc: "Play soccer like your favourite players!", Price: "$20.00", Img: "https://images.unsplash.com/photo-1614632537190-23e4146777db?ixlib=rb-1.2.1&ixid=MnwxMjA3fDB8MHxleHBsb3JlLWZlZWR8M3x8fGVufDB8fHx8&w=1000&q=80"},
	}

	Cart1 = []Item{}

	Cart2 = []Item{}

	Cart3 = []Item{}

	Cart4 = []Item{}

	Cart5 = []Item{}
	// create a WaitGroup
	wg := new(sync.WaitGroup)

	// var pointer1, pointer2, firstNode, secondNode Node
	// add two goroutines to `wg` WaitGroup
	wg.Add(NUM_NODES)

	for i := 1; i <= NUM_NODES; i++ {
		port := 9000 + i
		server := createNodeServer("Node "+strconv.Itoa(i), port)

		n := Node{i, server, port, nodeArray, successors}
		ringServer.nodeArray = append(ringServer.nodeArray, n)
		
	}

	fmt.Println("Completed ring server array:", ringServer.nodeArray)
	
	// Assign nodeArray to back up ring server
	backupRingServer.nodeArray = ringServer.nodeArray
	fmt.Println("Back up ring server array:", backupRingServer.nodeArray)

	// Activate gorouter to launch servers
	for _, n := range ringServer.nodeArray {
		fmt.Println("Server", n.id, "started on port", n.port, ". HTTP Server:", n.httpServer)
		go n.httpServer.ListenAndServe()
	}

	// go func() {
	// 	for {
	// 		select {
	// 		case <- kill:
	// 			fmt.Println("ENDING GO ROUTINE")
	// 			return
	// 		default:
	// 			fmt.Println("PRIMARY RING SERVER IS NOW LISTENING...")
	// 			ringServer.httpServer.ListenAndServe()
	// 		}
	// 	}
	// }()
	
	go ringServer.httpServer.ListenAndServe()
	// go pingRingServer(backupRingServer)


	// time.Sleep(5 * time.Second)
	// fmt.Println("Shutting down Primary Ring Server")
	// if err := ringServer.httpServer.Shutdown(context.TODO()); err != nil {
	// 	panic(err)
	// }


	// wait until WaitGroup is done
	wg.Wait()

}

// {"id":"2","name":"User2","price":"$1.00","desc":"Make your hair look neat with this"}
