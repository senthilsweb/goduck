package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"database/sql"
	"embed"
	"path/filepath"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/gorilla/mux"
	"github.com/natefinch/lumberjack"
	log "github.com/sirupsen/logrus"
)

// Command-line flags
var (
	flagPort          int
	flagSource        string
	flagEnv           string
	flagStaticDirPath string
	flagIndexFileName string
)

func init() {
	log.Info("Application init function start")
	flag.IntVar(&flagPort, "p", 8080, "port number to use for web ui server")
	flag.StringVar(&flagStaticDirPath, "d", "./dist", "Website directory path")
	flag.StringVar(&flagIndexFileName, "f", "index.html", "Index document")
	flag.StringVar(&flagEnv, "e", "dev", "Development or Production stage")
	flag.StringVar(&flagSource, "s", "embed", "Host site from embedded source or local filesystem")

	initLogger()

	log.Info("Application init function end.")
}

var baseEntityCols = []string{"id", "created_at", "modified_at", "update_count"}

var (
	db *sql.DB
)

type Entity struct {
	ID          int       `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
	UpdateCount int       `json:"update_count"`
}

func initializeDB() error {
	/*connStr := "stage.db"

	_, err := os.Stat("data.db")
	if os.IsNotExist(err) {
		// Create the database file if it doesn't exist
		_, err := os.Create("data.db")
		if err != nil {
			return err
		}
	}*/
	d, err := sql.Open("duckdb", "") //in memory
	if err != nil {
		return err
	}
	db = d
	//return nil
	if err != nil {
		return err
	}
	db = d
	return nil
}

func createEntity(w http.ResponseWriter, r *http.Request) {
	entityType := mux.Vars(r)["entity"]

	// Parse JSON data from request body
	var entityData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&entityData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Start a transaction to manage the ID sequence and entity insertion
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve the current ID from the sequence table and update it
	var nextID int
	err = tx.QueryRow("SELECT next_id FROM id_sequence WHERE entity_name = ?", entityType).Scan(&nextID)
	if err != nil {
		if err == sql.ErrNoRows {
			// If the entity doesn't exist in the sequence table, create it
			_, err = tx.Exec("INSERT INTO id_sequence (entity_name, next_id) VALUES (?, 1)", entityType)
			if err != nil {
				tx.Rollback()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			nextID = 1
		} else {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Construct the column list and placeholders for the dynamic query
	columns := []string{"id"}
	placeholders := []string{"?"}
	values := []interface{}{nextID}
	for columnName, value := range entityData {
		columns = append(columns, columnName)
		placeholders = append(placeholders, "?")
		values = append(values, value)
	}

	columns = append(columns, "created_at", "modified_at", "update_count")
	placeholders = append(placeholders, "?, ?, ?")
	values = append(values, time.Now(), time.Now(), 0)

	// Construct the dynamic SQL query for entity insertion
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", entityType, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	// Insert the entity using the retrieved ID and JSON data
	_, err = tx.Exec(query, values...)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the sequence table with the next ID
	_, err = tx.Exec("UPDATE id_sequence SET next_id = next_id + 1 WHERE entity_name = ?", entityType)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id": nextID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getEntities(w http.ResponseWriter, r *http.Request) {
	entityType := mux.Vars(r)["entity"]
	var columns string

	// Get the list of columns to be retrieved from the query parameters
	colsParam, hasColsParam := r.URL.Query()["select"]
	if hasColsParam {
		columns = colsParam[0]
	} else {
		columns = "*"
	}

	// Sorting
	var orderBy string
	orderParam, hasOrderParam := r.URL.Query()["order"]
	if hasOrderParam {
		orderBy = orderParam[0]
	}

	// Construct the base query for total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", entityType)

	// Parse where conditions from query parameters
	var whereConditions []string
	for key, values := range r.URL.Query() {
		if key != "select" && key != "limit" && key != "offset" && key != "order" {
			operator := "="
			if strings.HasSuffix(key, ".eq") {
				operator = "="
				key = strings.TrimSuffix(key, ".eq")
			} else if strings.HasSuffix(key, ".gt") {
				operator = ">"
				key = strings.TrimSuffix(key, ".gt")
			} else if strings.HasSuffix(key, ".gte") {
				operator = ">="
				key = strings.TrimSuffix(key, ".gte")
			} else if strings.HasSuffix(key, ".lt") {
				operator = "<"
				key = strings.TrimSuffix(key, ".lt")
			} else if strings.HasSuffix(key, ".lte") {
				operator = "<="
				key = strings.TrimSuffix(key, ".lte")
			} else if strings.HasSuffix(key, ".neq") {
				operator = "<>"
				key = strings.TrimSuffix(key, ".neq")
			} else if strings.HasSuffix(key, ".like") {
				operator = "ILIKE"
				key = strings.TrimSuffix(key, ".like")
			} // Add more cases for other operators

			// Construct filter conditions for each value
			for _, value := range values {
				// Handle spaces in the value by quoting it
				quotedValue := strings.ReplaceAll(value, "'", "''") // Escape single quotes
				whereConditions = append(whereConditions, fmt.Sprintf("LOWER(%s) %s LOWER('%s')", key, operator, quotedValue))
			}
		}
	}

	// Add WHERE clause if conditions are present
	if len(whereConditions) > 0 {
		countQuery = fmt.Sprintf("%s WHERE %s", countQuery, strings.Join(whereConditions, " OR "))
	}

	// Execute the count query
	var totalCount int
	err := db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Pagination
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	// Construct the base query for data retrieval
	baseQuery := fmt.Sprintf("SELECT %s FROM %s", columns, entityType)

	// Add WHERE clause if conditions are present
	if len(whereConditions) > 0 {
		baseQuery = fmt.Sprintf("%s WHERE %s", baseQuery, strings.Join(whereConditions, " OR "))
	}

	// Add ORDER BY clause if sorting is requested
	if orderBy != "" {
		baseQuery = fmt.Sprintf("%s ORDER BY %s", baseQuery, orderBy)
	}

	// Add LIMIT and OFFSET clauses
	query := fmt.Sprintf("%s LIMIT %d OFFSET %d", baseQuery, limit, offset)

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var entities []map[string]interface{}
	for rows.Next() {
		entity := make(map[string]interface{})
		values := make([]interface{}, len(columnNames))
		valuePointers := make([]interface{}, len(columnNames))
		for i := range columnNames {
			valuePointers[i] = &values[i]
		}

		err := rows.Scan(valuePointers...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for i, colName := range columnNames {
			entity[colName] = values[i]
		}
		entities = append(entities, entity)
	}

	// Calculate the current page based on offset and limit
	currentPage := (offset / limit) + 1

	// Calculate the next offset
	nextOffset := (currentPage * limit) + 1

	start := offset + 1
	end := offset + limit
	if end > totalCount {
		end = totalCount
	}

	// Construct the response object with additional information
	response := map[string]interface{}{
		"total_rows":   totalCount, // Updated to use the total count from the count query
		"limit":        limit,
		"offset":       offset,
		"current_page": currentPage,
		"next_offset":  nextOffset,
		"start":        start,
		"end":          end,
		"data":         entities,
	}

	w.Header().Set("Content-Type", "application/json")
	print(response)
	json.NewEncoder(w).Encode(response)
}

func updateEntity(w http.ResponseWriter, r *http.Request) {
	entityType := mux.Vars(r)["entity"]
	id := mux.Vars(r)["id"]

	// Get JSON payload from the request body
	var updatedEntity map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updatedEntity); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Build the update query and arguments
	var updateQuery string
	var args []interface{}
	for key, value := range updatedEntity {
		if key != "id" {
			updateQuery += key + " = ?, "
			args = append(args, value)
		}
	}
	updateQuery += "modified_at = ?, update_count = update_count + 1"
	args = append(args, time.Now())

	// Execute the update query
	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", entityType, updateQuery)
	args = append(args, id)
	_, err := db.Exec(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Entity updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func deleteEntity(w http.ResponseWriter, r *http.Request) {
	entityType := mux.Vars(r)["entity"]
	id := mux.Vars(r)["id"]

	// Delete the entity from the database
	_, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", entityType), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Entity deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func executeSelectQuery(query string) ([]map[string]interface{}, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for rows.Next() {
		entity := make(map[string]interface{})
		values := make([]interface{}, len(columnNames))
		valuePointers := make([]interface{}, len(columnNames))
		for i := range columnNames {
			valuePointers[i] = &values[i]
		}

		err := rows.Scan(valuePointers...)
		if err != nil {
			return nil, err
		}

		for i, colName := range columnNames {
			entity[colName] = values[i]
		}
		result = append(result, entity)
	}

	return result, nil
}

func executeDDLQuery(query string) error {
	fmt.Println(query)
	_, err := db.Exec(query)
	return err
}

func executeCustomQuery(w http.ResponseWriter, r *http.Request) {
	var queryData struct {
		Query string `json:"query" binding:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&queryData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.HasPrefix(strings.ToUpper(queryData.Query), "SELECT") {
		// Handle SELECT query
		results, err := executeSelectQuery(queryData.Query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"total_rows": len(results),
			"data":       results,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		// Handle DDL query
		fmt.Println("Executing DDL query " + queryData.Query)
		err := executeDDLQuery(queryData.Query)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"message": "Query executed successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func serveEmbeddedSite(w http.ResponseWriter, r *http.Request) {
	if flagEnv == "dev" {
		log.Info("Development environment: Serving site from local filesystem.")
		serveLocalSite(w, r)
		return
	}

	if flagEnv == "prod" {
		log.Info("Production environment: Serving site from embedded resources.")
		http.FileServer(http.Dir(flagStaticDirPath)).ServeHTTP(w, r)
		return
	}
}

func serveLocalSite(w http.ResponseWriter, r *http.Request) {
	indexPath := filepath.Join(flagStaticDirPath, flagIndexFileName)
	http.ServeFile(w, r, indexPath)
}

//go:embed dist
var embeddedFiles embed.FS

func serveEmbeddedSiteWithFS(w http.ResponseWriter, r *http.Request) {
	if flagEnv == "prod" {
		log.Info("Production environment: Serving site from embedded resources.")
		http.FileServer(http.FS(embeddedFiles)).ServeHTTP(w, r)
		return
	}

	log.Info("Development environment: Serving site from local filesystem.")
	serveLocalSite(w, r)
}

func main() {
	// Parse command-line flags
	flag.Parse()

	// Initialize the database
	err := initializeDB()
	if err != nil {
		log.Fatal("Error initializing the database: ", err)
	}

	// Create the HTTP router
	r := mux.NewRouter()

	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		// an example API handler
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	r.HandleFunc("/{entity}", createEntity).Methods("POST")
	r.HandleFunc("/{entity}", getEntities).Methods("GET")
	r.HandleFunc("/{entity}/{id}", updateEntity).Methods("PUT")
	r.HandleFunc("/{entity}/{id}", deleteEntity).Methods("DELETE")
	r.HandleFunc("/query", executeCustomQuery).Methods("POST")

	r.PathPrefix("/").HandlerFunc(serveEmbeddedSiteWithFS)

	http.Handle("/", r)

	// Start the HTTP server
	addr := fmt.Sprintf(":%d", flagPort)
	log.Infof("Starting server on %s...", addr)

	// Set up graceful shutdown
	server := &http.Server{Addr: addr, Handler: r}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting server: ", err)
		}
	}()

	// Wait for signals to gracefully shut down the server
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Info("Shutting down server...")

	// Set a timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Info("Server exiting")
}

// initLogger initializes the logger with the desired configuration.
func initLogger() {
	// Log as JSON instead of the default ASCII formatter
	log.SetFormatter(&log.JSONFormatter{})

	// Log to a file, max size 100MB, 3 backups, and delete logs older than 28 days
	log.SetOutput(&lumberjack.Logger{
		Filename:   "app.log",
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}
