package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var queue []string
var masterConn net.Conn

func initLocalDB() {
  var err error
  db, err = sql.Open("mysql", "root:rootroot@tcp(127.0.0.1:3306)/slavedb")
  if err != nil || db.Ping() != nil {
    panic("❌ Failed to connect to slave DB")
  }
  fmt.Println("✅ Slave connected to local DB.")
}

func connectToMaster() net.Conn {
  conn, err := net.Dial("tcp", "192.168.171.242:9090")
  if err != nil {
    panic("❌ Cannot connect to master")
  }
  fmt.Println("🔌 Connected to master.")
  return conn
}

func listenToMaster(conn net.Conn) {
  scanner := bufio.NewScanner(conn)
  for scanner.Scan() {
    query := scanner.Text()
    fmt.Println("📥 Received from master:", query)
    queue = append(queue, query)
  }
  fmt.Println("❌ Lost connection to master.")
  os.Exit(1)
}

func applyQueries() {
  for {
    time.Sleep(5 * time.Second)
    if len(queue) == 0 {
      continue
    }
    fmt.Printf("🔁 Applying %d queries...\n", len(queue))
    for _, q := range queue {
      if strings.HasPrefix(strings.ToUpper(q), "SELECT") {
        continue
      }

      _, err := db.Exec(q)
      if err != nil {

      } else {
        fmt.Println("✅ Applied:", q)
      }
    }
    queue = nil
  }
}

func startCLI() {
  reader := bufio.NewReader(os.Stdin)
  for {
    fmt.Print("📝 Enter SQL (safe only): ")
    query, _ := reader.ReadString('\n')
    query = strings.TrimSpace(query)

    if query == "" {
      continue
    }
    qUpper := strings.ToUpper(query)

    if strings.HasPrefix(qUpper, "CREATE") || strings.HasPrefix(qUpper, "DROP") || strings.HasPrefix(qUpper, "ALTER") {
      fmt.Println("⛔️ Not allowed: CREATE, DROP, ALTER only allowed from master.")
      continue
    }

    _, err := db.Exec(query)
    if err != nil {
      fmt.Println("❌ Error executing on slave:", err)
      continue
    }

    fmt.Println("✅ Executed on slave.")

    _, err = fmt.Fprintln(masterConn, query)
    if err != nil {
      fmt.Println("❌ Failed to send query to master:", err)
    } else {
      fmt.Println("📤 Sent to master for replication.")
    }
  }
}

func main() {
  initLocalDB()
  masterConn = connectToMaster()
  defer masterConn.Close()

  go listenToMaster(masterConn)
  go applyQueries()
  go startCLI()

  select {}
}
