package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os" // Add this line to import the os package
	"strconv"
	"time"
	"strings"


	"github.com/gin-gonic/gin"
	_ "github.com/marcboeker/go-duckdb" // Update the import path
)
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
	connStr := "data.db"
	
	_, err := os.Stat("data.db")
	if os.IsNotExist(err) {
		// Create the database file if it doesn't exist
		_, err := os.Create("data.db")
		if err != nil {
			return err
		}
	}
	d, err := sql.Open("duckdb", connStr)
	if err != nil {
		return err
	}
	db = d
	return nil
	if err != nil {
		return err
	}
	db = d
	return nil
}



func createEntity(c *gin.Context) {
	entityType := c.Param("entity")

	// Parse JSON data from request body
	var entityData map[string]interface{}
	if err := c.ShouldBindJSON(&entityData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start a transaction to manage the ID sequence and entity insertion
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			nextID = 1
		} else {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update the sequence table with the next ID
	_, err = tx.Exec("UPDATE id_sequence SET next_id = next_id + 1 WHERE entity_name = ?", entityType)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": nextID})
}


func getEntities(c *gin.Context) {
	entityType := c.Param("entity")
	var columns string

	// Get the list of columns to be retrieved from the query parameters
	colsParam, hasColsParam := c.GetQuery("select")
	if hasColsParam {
		columns = colsParam
	} else {
		columns = "*"
	}

	// Pagination
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Sorting
	var orderBy string
	orderParam, hasOrderParam := c.GetQuery("order")
	if hasOrderParam {
		orderBy = orderParam
	}

	// Construct the base query
	baseQuery := fmt.Sprintf("SELECT %s FROM %s", columns, entityType)

	// Parse where conditions from query parameters
	var whereConditions []string
	for key, values := range c.Request.URL.Query() {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for i, colName := range columnNames {
			entity[colName] = values[i]
		}
		entities = append(entities, entity)
	}

	// Construct the response object with additional information
	response := gin.H{
		"total_rows": len(entities),
		"limit":      limit,
		"offset":     offset,
		"data":       entities,
	}

	c.JSON(http.StatusOK, response)
}





func updateEntity(c *gin.Context) {
	entityType := c.Param("entity")
	id := c.Param("id")

	// Get JSON payload from the request body
	var updatedEntity map[string]interface{}
	if err := c.ShouldBindJSON(&updatedEntity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Entity updated successfully"})
}


func deleteEntity(c *gin.Context) {
	entityType := c.Param("entity")
	id := c.Param("id")

	// Delete the entity from the database
	_, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", entityType), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Entity deleted successfully"})
}


func main() {
	if err := initializeDB(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := gin.Default()

	r.POST("/:entity", createEntity)
	r.GET("/:entity", getEntities)
	r.PUT("/:entity/:id", updateEntity)
	r.DELETE("/:entity/:id", deleteEntity)

	r.Run(":8080")
}