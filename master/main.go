package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var clients []net.Conn
var mu sync.Mutex

func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:rootroot@tcp(127.0.0.1:3306)/masterdb")
	if err != nil || db.Ping() != nil {
		panic("❌ Master DB connection failed")
	}
	fmt.Println("✅ Master connected to MySQL.")
}

func handleSlave(conn net.Conn) {
	defer conn.Close()
	mu.Lock()
	clients = append(clients, conn)
	mu.Unlock()

	fmt.Println("🔌 New slave connected:", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		query := scanner.Text()
		fmt.Println("📥 Received query:", query)

		qUpper := strings.ToUpper(query)

		// Allow CREATE, DROP, ALTER, TRUNCATE commands on the master
		if strings.HasPrefix(qUpper, "CREATE") || strings.HasPrefix(qUpper, "DROP") ||
			strings.HasPrefix(qUpper, "ALTER") || strings.HasPrefix(qUpper, "TRUNCATE") {
			_, err := db.Exec(query)
			if err != nil {
				fmt.Println("❌ Execution error:", err)
				fmt.Fprintln(conn, "ERROR: "+err.Error())
				continue
			}
			// Broadcast the query to slaves for replication
			fmt.Println("✅ Executed successfully and broadcasted:", query)
			fmt.Fprintln(conn, "OK")
			broadcastToSlaves(query, conn)
			continue
		}

		// Process SELECT queries
		if strings.HasPrefix(qUpper, "SELECT") {
			rows, err := db.Query(query)
			if err != nil {
				fmt.Println("❌ SELECT error:", err)
				fmt.Fprintln(conn, "ERROR: "+err.Error())
				continue
			}
			defer rows.Close()

			cols, err := rows.Columns()
			if err != nil {
				fmt.Println("❌ Columns error:", err)
				fmt.Fprintln(conn, "ERROR: "+err.Error())
				continue
			}

			// Sending column names
			fmt.Fprintf(conn, "COLUMNS:%d\n", len(cols))
			for _, col := range cols {
				fmt.Fprintln(conn, col)
			}

			// Processing data rows
			values := make([]interface{}, len(cols))
			valuePtrs := make([]interface{}, len(cols))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			for rows.Next() {
				rows.Scan(valuePtrs...)
				for _, col := range values {
					var v string
					if col == nil {
						v = "NULL"
					} else {
						switch val := col.(type) {
						case []byte:
							v = string(val)
						default:
							v = fmt.Sprintf("%v", val)
						}
					}
					fmt.Fprintln(conn, v)
				}
			}
			fmt.Fprintln(conn, "END_RESULTS")
			continue
		}

		// Handle SHOW DATABASES command
		if strings.HasPrefix(qUpper, "SHOW DATABASES") {
			rows, err := db.Query(query)
			if err != nil {
				fmt.Println("❌ SHOW DATABASES error:", err)
				fmt.Fprintln(conn, "ERROR: "+err.Error())
				continue
			}
			defer rows.Close()

			// Sending database names
			fmt.Fprintln(conn, "Databases:")
			for rows.Next() {
				var dbName string
				if err := rows.Scan(&dbName); err != nil {
					fmt.Println("❌ Scan error:", err)
					fmt.Fprintln(conn, "ERROR: "+err.Error())
					continue
				}
				fmt.Fprintln(conn, dbName)
			}
			fmt.Fprintln(conn, "END_RESULTS")
			continue
		}

		// Handle SHOW TABLES command
		if strings.HasPrefix(qUpper, "SHOW TABLES") {
			rows, err := db.Query(query)
			if err != nil {
				fmt.Println("❌ SHOW TABLES error:", err)
				fmt.Fprintln(conn, "ERROR: "+err.Error())
				continue
			}
			defer rows.Close()

			// Sending table names
			fmt.Fprintln(conn, "Tables:")
			for rows.Next() {
				var tableName string
				if err := rows.Scan(&tableName); err != nil {
					fmt.Println("❌ Scan error:", err)
					fmt.Fprintln(conn, "ERROR: "+err.Error())
					continue
				}
				fmt.Fprintln(conn, tableName)
			}
			fmt.Fprintln(conn, "END_RESULTS")
			continue
		}

		// Handle other SQL commands like INSERT, UPDATE, DELETE
		_, err := db.Exec(query)
		if err != nil {
			fmt.Println("❌ Execution error:", err)
			fmt.Fprintln(conn, "ERROR: "+err.Error())
			continue
		}
		fmt.Println("✅ Executed successfully")
		fmt.Fprintln(conn, "OK")
		broadcastToSlaves(query, conn)
	}
	fmt.Println("❌ Slave disconnected:", conn.RemoteAddr())
}

func broadcastToSlaves(query string, exclude net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	for _, conn := range clients {
		if conn != exclude {
			_, err := fmt.Fprintln(conn, query)
			if err != nil {
				fmt.Println("❌ Broadcast failed:", err)
			}
		}
	}
}

func main() {
	initDB()
	listener, err := net.Listen("tcp", "0.0.0.0:9090") // Accepts connections from other PCs
	if err != nil {
		panic("❌ Failed to start server: " + err.Error())
	}
	defer listener.Close()
	fmt.Println("🖧 Master listening on port 9090...")

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("❌ Accept error:", err)
				continue
			}
			go handleSlave(conn)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("📝 Master SQL: ")
		query, _ := reader.ReadString('\n')
		query = strings.TrimSpace(query)
		if query == "exit" {
			break
		}

		if strings.ToUpper(query) == "SHOW SLAVES" {
			mu.Lock()
			fmt.Printf("Connected slaves (%d):\n", len(clients))
			for i, conn := range clients {
				fmt.Printf("%d: %v\n", i+1, conn.RemoteAddr())
			}
			mu.Unlock()
			continue
		}

		// Handle CREATE, DROP, ALTER, TRUNCATE commands
		_, err := db.Exec(query)
		if err != nil {
			fmt.Println("❌ Error:", err)
			continue
		}

		// Broadcast the query to slaves for replication
		fmt.Println("✅ Executed")
		broadcastToSlaves(query, nil)
	}
}
