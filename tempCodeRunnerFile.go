package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql" // Import MySQL driver
	"github.com/joho/godotenv"
)

type User struct {
	User_ID  int    `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func main() {
	// เรียกใช้งาน environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file!")
	}

	// เริ่มเซิร์ฟเวอร์
	http.HandleFunc("/insert", createUser)
	fmt.Println("Server running at port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// dbConnection สร้างการเชื่อมต่อฐานข้อมูล
func dbConnection() (*sql.DB, error) {
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUsername, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("Error connecting database: %v", err)
	}

	// ตรวจสอบการเชื่อมต่อ
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Database not responding: %v", err)
	}

	return db, nil
}

// createUser รับ request สร้างผู้ใช้งานใหม่
func createUser(w http.ResponseWriter, r *http.Request) {
	// ตรวจสอบ method ว่าเป็น POST หรือไม่
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// เชื่อมต่อกับฐานข้อมูล
	db, err := dbConnection()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer db.Close()

	// อ่านข้อมูลจาก body
	var user User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// ตรวจสอบข้อมูลที่จำเป็น
	if user.Username == "" || user.Password == "" || user.Email == "" {
		http.Error(w, "Missing required fields", http.StatusInternalServerError)
		return
	}

	// เพิ่มข้อมูลผู้ใช้ใหม่ในฐานข้อมูล
	result, err := db.Exec("INSERT INTO users (username, password, email) VALUES (?, ?, ?)", user.Username, user.Password, user.Email)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		log.Println("Error inserting user:", err)
		return
	}

	// ดึง ID ของผู้ใช้ที่ถูกสร้าง
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Error retrieving last insert ID", http.StatusInternalServerError)
		log.Println("Error getting last insert ID:", err)
		return
	}

	// เซ็ตค่า user_id ให้กับ user
	user.User_ID = int(lastInsertID)
	fmt.Println(user.User_ID)

	// ตอบกลับ JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
