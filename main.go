package main

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "net/http"
    "github.com/gin-gonic/gin"
)
var db *sql.DB

func initDB() {
    var err error
    db, err = sql.Open("mysql", "root:10933800@tcp(127.0.0.1:3306)/userdb")
    if err != nil {
        panic(err)
    }
    if err = db.Ping(); err != nil {
        panic(err)
    }
}

type user struct {
    ID_NO           string `json:"id"`
    Password        string `json:"password"`
    Fname           string `json:"fname"`
    Lname           string `json:"lname"` // fixed missing quote
    Care_History    string `json:"takecarehistory"`
    Agreement       bool   `json:"agreement"` // changed boolean to bool
    Disability_type string `json:"disability_type"`
    Special_request string `json:"special_request"`
}

type userResponse struct {
    ID_NO           string `json:"id"`
    Fname           string `json:"fname"`
    Lname           string `json:"lname"`
    Care_History    string `json:"takecarehistory"`
    Agreement       bool   `json:"agreement"`
    Disability_type string `json:"disability_type"`
    Special_request string `json:"special_request"`
}


func main() {
    initDB()
    router := gin.Default()
    router.GET("/users", getUsers)
    router.GET("/users/:id", getUserByID)
    router.POST("/users", postUsers)
    router.POST("/login", login)

    router.Run(":8080")
}

func getUsers(c *gin.Context) {
    rows, err := db.Query("SELECT id, fname, lname, takecarehistory, agreement, disability_type, special_request FROM users")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var users []userResponse
    for rows.Next() {
        var u userResponse
        err := rows.Scan(&u.ID_NO, &u.Fname, &u.Lname, &u.Care_History, &u.Agreement, &u.Disability_type, &u.Special_request)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        users = append(users, u)
    }
    c.IndentedJSON(http.StatusOK, users)
}

func postUsers(c *gin.Context) {
    var newUser user
    if err := c.BindJSON(&newUser); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    _, err := db.Exec(
        "INSERT INTO users (id, password, fname, lname, takecarehistory, agreement, disability_type, special_request) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
        newUser.ID_NO, newUser.Password, newUser.Fname, newUser.Lname, newUser.Care_History, newUser.Agreement, newUser.Disability_type, newUser.Special_request,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.IndentedJSON(http.StatusCreated, newUser)
}

func getUserByID(c *gin.Context) {
    id := c.Param("id")
    var u userResponse
    err := db.QueryRow(
        "SELECT id, fname, lname, takecarehistory, agreement, disability_type, special_request FROM users WHERE id = ?",
        id,
    ).Scan(&u.ID_NO, &u.Fname, &u.Lname, &u.Care_History, &u.Agreement, &u.Disability_type, &u.Special_request)
    if err == sql.ErrNoRows {
        c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
        return
    } else if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.IndentedJSON(http.StatusOK, u)
}

func login(c *gin.Context) {
    var credentials struct {
        ID_NO    string `json:"id"`
        Password string `json:"password"`
    }
    if err := c.BindJSON(&credentials); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var storedPassword string
    err := db.QueryRow("SELECT password FROM users WHERE id = ?", credentials.ID_NO).Scan(&storedPassword)
    if err == sql.ErrNoRows {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid id or password"})
        return
    } else if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    if credentials.Password != storedPassword {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid id or password"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "login successful"})
}