package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	_ "github.com/lib/pq"
	"database/sql"

	"strconv"

	//"bufio"
)

func main() {

	fmt.Println("listening...")
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
 	}
	http.HandleFunc("/", hijack_wrap)
	http.HandleFunc("/web",web_test)
	err := http.ListenAndServe(":"+port, nil)

	if err != nil {
		panic(err)
	}
}
func web_test(res http.ResponseWriter, req *http.Request) {
	fmt.Println(res,"web_test")
}
func hijack_wrap(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("test"))
	if f, ok := res.(http.Flusher); ok { 
		f.Flush() 
	} else { 
		fmt.Println("Damn, no flush"); 
	}
	fmt.Fprintln(res, "test1")
	hj, ok := res.(http.Hijacker)
	if !ok {
		http.Error(res, "server doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(res, "test2")
	conn, _, _ := hj.Hijack()
	handleConnection(conn)


}
func handleConnection(conn net.Conn) { // *bufio.ReadWriter) {
	conn.Write([]byte("Connected."))
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
}
