/*
The following application is to demonstratte the use goroutines and channels in a web server
The theme is tour management where themultiple users can read or write into 'tours-template.json' filesimultaneously
The idea is to read and/or write the data into 'tours-template.json' file simultaneosly
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

var wg = sync.WaitGroup{}

func main() {
	//Here there are 3 endpoints
	//the below handle is for the root path of the server which shows the welcome message to the application
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi Welcome to tours Application")

	})
	//the following handle is to read the existent tours-template.json file and write the file into the client side
	http.HandleFunc("/viewTours", func(w http.ResponseWriter, r *http.Request) {
		file, _ := ioutil.ReadFile("tours-template.json")
		resBodyString := string(file)
		fmt.Fprintf(w, resBodyString)
	})
	//the following handle is to add a tour by the user
	http.HandleFunc("/addTour", HandleAddTour)

	//starting a server
	err := http.ListenAndServe("localhost:3000", nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err.Error())
	}
}

func HandleAddTour(w http.ResponseWriter, r *http.Request) {
	// With Reference to --> https://medium.com/eaciit-engineering/better-way-to-read-and-write-json-file-in-golang-9d575b7254f2

	//-------------------Reading JSON File-----------------------------------
	ch := make(chan []map[string]interface{})
	defer close(ch) //it closes the channel after the excecution of the function
	go readJSON(ch) //reading the json file, and simultaneosly writing the json on line no.79

	//-------------------Reading Request Body-------------------------------
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	//from here on the following operations are needed to be performed in order to write into a json file
	tourData := map[string]interface{}{}
	json.Unmarshal(data, &tourData)

	// ------------------Creating  New Interface for new Data-------------------
	newToursData := []map[string]interface{}{}

	//------------Getting data from JSON File through channel ch---------------------
	allToursData := <-ch

	//---------Adding file data and req data into prev created interface----------
	for _, d := range allToursData {
		newToursData = append(newToursData, d)
	}

	newToursData = append(newToursData, tourData)

	jsonString, _ := json.Marshal(newToursData)

	//---------------------------Writing the file-------------------------------
	wg.Add(1)
	go writeNewData(jsonString)
	// We can write any number of statements here which we can to execute concurrently with write function
	// before wait()
	go testConcurrency()

	wg.Wait() //---> with this the application gets blocked untill file is written and Done() method is called

	fmt.Fprintf(w, "Your Tour Has been added Successfully", string(data))
}

func readJSON(ch chan<- []map[string]interface{}) {
	allToursData := []map[string]interface{}{}
	file, _ := ioutil.ReadFile("tours-template.json")
	json.Unmarshal(file, &allToursData)
	fmt.Println("Reading file is done here")
	ch <- allToursData
}

func writeNewData(jsonString []byte) {
	ioutil.WriteFile("tours-template.json", jsonString, os.ModePerm)
	fmt.Println("Writing into a file is done here")
	defer wg.Done()
}

func testConcurrency() {
	for i := 1; i < 5; i++ {
		fmt.Println("Testing for concurrency", i)
		time.Sleep(time.Second)
	}
}
