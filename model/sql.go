package model

import "encoding/json"

// Query model
type Query struct {
	Statement     string
	QueryType     string
	DirectoryPath string
	Args          []string
	RawOutput     bool
	Db            DBConn
}

//DBConn model
type DBConn struct {
	Host     string
	User     string
	Port     string
	Password string
	Dbname   string
}

// MarshalBinary encodes the struct into a binary blob
// Here I cheat and use regular json :)
func (q *Query) MarshalBinary() ([]byte, error) {
	return json.Marshal(q)
}

// UnmarshalBinary decodes the struct into a User
func (q *Query) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &q); err != nil {
		return err
	}
	return nil
}
