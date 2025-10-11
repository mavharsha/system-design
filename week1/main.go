package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DBConnectionPool represents a custom connection pool with a blocking queue
type DBConnectionPool struct {
	connections chan *sql.DB // Buffered channel acts as blocking queue
	dsn         string
	poolSize    int
}

// NewDBConnectionPool creates a new connection pool with specified size
func NewDBConnectionPool(dsn string, poolSize int) (*DBConnectionPool, error) {
	pool := &DBConnectionPool{
		connections: make(chan *sql.DB, poolSize), // Buffered channel = blocking queue
		dsn:         dsn,
		poolSize:    poolSize,
	}

	// Initialize the pool with connections
	for i := 0; i < poolSize; i++ {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to create connection %d: %v", i, err)
		}

		// Test the connection
		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping connection %d: %v", i, err)
		}

		// Put connection in the pool
		pool.connections <- db
		log.Printf("Connection %d initialized and added to pool", i+1)
	}

	return pool, nil
}

// GetConnection retrieves a connection from the pool (blocks if none available)
func (p *DBConnectionPool) GetConnection() *sql.DB {
	// This will block if the channel is empty (all connections in use)
	// Once a connection is available, it will be returned
	log.Println("Requesting connection from pool...")
	conn := <-p.connections
	log.Println("Connection acquired from pool")
	return conn
}

// PutConnection returns a connection back to the pool
func (p *DBConnectionPool) PutConnection(conn *sql.DB) {
	// This will block if the channel is full (should never happen in correct usage)
	log.Println("Returning connection to pool")
	p.connections <- conn
}

// Close closes all connections in the pool
func (p *DBConnectionPool) Close() {
	close(p.connections)
	for conn := range p.connections {
		conn.Close()
	}
	log.Println("All connections closed")
}

func main() {
	// Example DSN (Data Source Name) for MySQL
	// Format: username:password@tcp(host:port)/database
	dsn := "user:password@tcp(localhost:3306)/online_status_db"

	// Create a connection pool with 10 connections
	pool, err := NewDBConnectionPool(dsn, 10)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}
	defer pool.Close()

	// Example usage: Simulate multiple concurrent requests
	for i := 0; i < 15; i++ {
		go func(requestID int) {
			// Get a connection from the pool (blocks if all 10 are in use)
			conn := pool.GetConnection()
			
			// Use the connection to perform DB operations
			log.Printf("Request %d: Using connection for heartbeat update", requestID)
			
			// Simulate DB operation
			_, err := conn.Exec("UPDATE user_status SET last_seen = ? WHERE user_id = ?", 
				time.Now().Unix(), fmt.Sprintf("user_%d", requestID))
			if err != nil {
				log.Printf("Request %d: Error: %v", requestID, err)
			}
			
			// Simulate some work
			time.Sleep(100 * time.Millisecond)
			
			// Return the connection back to the pool
			pool.PutConnection(conn)
			log.Printf("Request %d: Completed", requestID)
		}(i)
	}

	// Wait for all goroutines to complete
	time.Sleep(3 * time.Second)
	log.Println("All requests completed")
}

