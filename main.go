package main

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "net/http"
    "github.com/gin-gonic/gin"
    "fmt"
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

type task struct {
    ID            int64  `json:"id,omitempty"` // changed from TaskID
    TaskOwnerID   string `json:"task_owner_id"`
    TaskName      string `json:"task_name"`
    StartDateTime string `json:"start_date_time"`
    EndDateTime   string `json:"end_date_time"`
    Location      string `json:"location"`
    MoreDetail    string `json:"more_detail"`
    TaskWorkerID  string `json:"task_worker_id"`
}

type taskShowCaretaker struct {
    TaskName        string `json:"task_name"`
    StartDateTime   string `json:"start_date_time"`
    EndDateTime     string `json:"end_date_time"`
    DisabilityType  string `json:"disability_type"`
}

type taskUserHistory struct {
    ID            int64   `json:"id"`
    TaskName      string  `json:"task_name"`
    StartDateTime string  `json:"start_date_time"`
    EndDateTime   string  `json:"end_date_time"`
    CaretakerName *string `json:"caretaker_name"` // pointer allows null
}

func main() {
    initDB()
    router := gin.Default()
    router.GET("/users", getUsers)
    router.GET("/users/:id", getUserByID)
    router.POST("/users", postUsers)
    router.POST("/login", login)
    router.POST("/tasks", postTask)
    router.GET("/tasks/caretaker", getTasksForCaretaker)
    router.GET("/tasks/user/history", getTaskUserHistory)
    router.GET("/tasks/:task_id", getTask) // <-- Add this line

    router.Run("localhost:8080")
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

func postTask(c *gin.Context) {
    var newTask task
    if err := c.BindJSON(&newTask); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    result, err := db.Exec(
        "INSERT INTO tasks (task_owner_id, task_name, start_date_time, end_date_time, location, more_detail, task_worker_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
        newTask.TaskOwnerID, newTask.TaskName, newTask.StartDateTime, newTask.EndDateTime, newTask.Location, newTask.MoreDetail, newTask.TaskWorkerID,
    )
    if err != nil {
        fmt.Println("DB error:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    id, err := result.LastInsertId()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve task ID"})
        return
    }
    newTask.ID = id // Set the id

    c.IndentedJSON(http.StatusCreated, newTask)
}

func getTasksForCaretaker(c *gin.Context) {
    rows, err := db.Query(`
        SELECT t.task_name, t.start_date_time, t.end_date_time, u.disability_type
        FROM tasks t
        JOIN users u ON t.task_owner_id = u.id
    `)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var tasks []taskShowCaretaker
    for rows.Next() {
        var task taskShowCaretaker
        err := rows.Scan(&task.TaskName, &task.StartDateTime, &task.EndDateTime, &task.DisabilityType)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        tasks = append(tasks, task)
    }
    c.IndentedJSON(http.StatusOK, tasks)
}

func getTaskUserHistory(c *gin.Context) {
    userID := c.Query("user_id")
    rows, err := db.Query(`
        SELECT t.id, t.task_name, t.start_date_time, t.end_date_time, u.fname
        FROM tasks t
        LEFT JOIN users u ON t.task_worker_id = u.id
        WHERE t.task_owner_id = ?
    `, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var history []taskUserHistory
    for rows.Next() {
        var h taskUserHistory
        err := rows.Scan(&h.ID, &h.TaskName, &h.StartDateTime, &h.EndDateTime, &h.CaretakerName)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        history = append(history, h)
    }
    c.IndentedJSON(http.StatusOK, history)
}

func getTask(c *gin.Context) {
    taskID := c.Param("task_id")
    var t task
    err := db.QueryRow(`
        SELECT id, task_owner_id, task_name, start_date_time, end_date_time, location, more_detail, task_worker_id
        FROM tasks
        WHERE id = ?
    `, taskID).Scan(
        &t.ID, &t.TaskOwnerID, &t.TaskName, &t.StartDateTime, &t.EndDateTime, &t.Location, &t.MoreDetail, &t.TaskWorkerID,
    )
    if err == sql.ErrNoRows {
        c.JSON(http.StatusNotFound, gin.H{"message": "task not found"})
        return
    } else if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.IndentedJSON(http.StatusOK, t)
}