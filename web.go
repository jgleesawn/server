package main

import (
	"fmt"
	"net"
	//"net/http"
	"os"
	_ "github.com/lib/pq"
	"database/sql"

	"strconv"

)

func main() {

//	http.HandleFunc("/", hello)
//	http.HandleFunc("/test", test)
	fmt.Println("listening...")
	ln, err := net.Listen("tcp", ":"+os.Getenv("PORT"))
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}
func handleConnection(conn net.Conn) {
	db_url := os.Getenv("DATABASE_URL")
	db_name := "demoapp_db"
	sslmode := "sslmode=disable"
	db, err := sql.Open("postgres", db_url+"/"+db_name+"?"+sslmode)
	if err != nil {
		panic(err)
	}
	c := make(chan []byte)
	ec := make(chan error)
	go func(c chan []byte, ec chan error) {
		for {
			data := make([]byte, 512)
			l,err := conn.Read(data)
			if err != nil {
				ec<- err
				return
			}
			c<-data[0:l-1]
		}
	}(c,ec)
	for {
		select {
		case data := <- c:
			conn.Write(data)
			switch string(data) {
			case "get":
//				fmt.Println("get")
				var age int
				name := <-c
				rows, _ := db.Query(`SELECT name, age FROM users WHERE name = $1;`,name)
				rows.Next()
				rows.Scan(&name,&age)
	conn.Write([]byte(string(name)+" is "+strconv.Itoa(age)+" years old."))
				break;
			case "store":
//				fmt.Println("store")
				name := <-c
				age,_ := strconv.Atoi(string(<-c))
	db.QueryRow(`INSERT INTO users VALUES($1,$2);`,string(name),int(age))
//	fmt.Println(string(name))
//	fmt.Println(strconv.Itoa(age))
				break;
			}
			break;
		case err := <- ec:
			fmt.Println(err)
			break;
	//	case time.Tick(time.Second)
		}
	}
	var age int
	err = db.QueryRow(`INSERT INTO users VALUES('Peter',21) RETURNING age`).Scan(&age)
	fmt.Println(age)
	var name string
	err = db.QueryRow("SELECT name FROM users WHERE age = $1", age).Scan(&name)
	if err != nil {
		panic(err)
	}
	fmt.Println(name)
	name = "a"
	err = db.QueryRow("SELECT name FROM users WHERE age = $1", age).Scan(&name)
	fmt.Println(name)
	fmt.Fprintf(conn,"BODY")
}
/*
func test(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "test")
}
func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "hello, world")
}
*/
