
# Distributed Master-Slave Database System

This project implements a dynamic distributed database system using Go with a master-slave architecture. It supports full SQL operations, inter-node communication via TCP, and real-time data replication.

## Features

- 🧠 Master-Slave Architecture  
- 🗃️ Dynamic Database & Table Creation  
- 🔄 Full SQL Support (CREATE, INSERT, SELECT, UPDATE, DELETE, DROP)  
- 📡 Inter-Node Communication using TCP  
- 💽 Data Replication from Master to Slaves  
- 🖥️ Web Interface for Node Management (e.g., shutdown, wallpaper update)

---

## Components

### 🧩 Master Node

- Accepts and processes all SQL operations.
- Can create and drop databases/tables.
- Sends changes (queries, commands, files) to slave nodes.
- Hosts a web UI to monitor and control slave nodes.

### 📦 Slave Nodes (Snaps)

- Receives queries and commands from the master.
- Executes valid SQL operations and updates its local database accordingly.
- Responds to shutdown and wallpaper change commands.
- Communicates with the master over TCP.

---

## Web Interface

- View connected slave nodes (Snaps)
- Send control commands:  
  - 🔌 Shutdown  
  - 🖼️ Change Wallpaper  
- Real-time updates from all nodes

---

## How to Run

1. **Clone the repository**  
   ```bash
   git clone https://github.com/Alaamohamed2002/distributed-db-system.git
   cd distributed-db-system
   ```

2. **Start the Master Node**  
   ```bash
   go run master.go
   ```

3. **Start One or More Slave Nodes (Snaps)**  
   ```bash
   go run slave.go
   ```

4. **Open Web UI**  
   Navigate to `http://localhost:8080` in your browser.

---

## Dependencies

- Go (1.18+ recommended)
- SQLite (used internally for database operations)
- Standard Go packages only (net, os, bufio, etc.)

---

## Communication Protocol

- TCP sockets are used for messaging between nodes.
- Custom protocol for sending queries and command types (e.g., WALLPAPER, SHUTDOWN).
- Data format: serialized strings, optionally JSON for web UI.

---

## Example SQL Operations

```sql
CREATE DATABASE SnapDB;
USE SnapDB;
CREATE TABLE Users (id INTEGER, name TEXT);
INSERT INTO Users VALUES (1, 'Alice');
SELECT * FROM Users;
UPDATE Users SET name = 'Bob' WHERE id = 1;
DELETE FROM Users WHERE id = 1;
DROP DATABASE SnapDB;
```

---

## Future Improvements

- HTTPS support for web interface
- Authentication and access control
- Graphical data visualization
- Real-time logs via WebSockets

---

## License

MIT License

---

## Author

Alaa Mohamed  
GitHub: [Alaamohamed2002](https://github.com/Alaamohamed2002)
