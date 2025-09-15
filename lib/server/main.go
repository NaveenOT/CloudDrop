package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var savePath string

func main() {
	fmt.Println("Hello World")
	savePath = os.Args[1]
	initDB()
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", nil)
	})
	r.GET("/files", allfiles)
	r.POST("/upload", uploadfile)
	r.GET("/download/:filename", downloadfile)

	r.Run(":8080")
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./coulddrop.db")
	if err != nil {
		log.Fatal("Unable to Open Database Connection", err)
	}
	creation_query := `CREATE TABLE IF NOT EXISTS files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			type TEXT,
			time TEXT
		);`
	_, err = db.Exec(creation_query)
	if err != nil {
		log.Fatal("Error in Creation of Database", err)
	}

}

func uploadfile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "Error Reading File '%v'", err.Error())
		return
	}
	savePath := filepath.Join(savePath, filepath.Base(file.Filename))
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.String(http.StatusInternalServerError, "Error Saving file: '%v'", err.Error())
		return
	}
	_, err = db.Exec(`INSERT INTO files(name, type, time) VALUES(?, ?, ?);`,
		file.Filename, file.Header.Get("Content-Type"), time.Now().Format(time.RFC3339))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error inserting into database '%v'", err.Error())
		return
	}
	c.String(http.StatusOK, "File '%s' Uploaded Successfully", file.Filename)
}

type meta struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Time string `json:"time"`
}

func downloadfile(c *gin.Context) {
	filename := c.Param("filename")
	path := filepath.Join(savePath, filepath.Base(filename))
	c.FileAttachment(path, filename)
}

func allfiles(c *gin.Context) {
	rows, err := db.Query(`SELECT name, type, time FROM files`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	defer rows.Close()
	var files []meta
	for rows.Next() {
		var f meta
		rows.Scan(&f.Name, &f.Type, &f.Time)
		files = append(files, f)
	}
	c.JSON(http.StatusOK, files)
}
