package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"toster/models"

	socketio "github.com/doquangtan/socket.io/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type LinkPreview struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	SiteName    string `json:"siteName"`
	MessageText string `json:"messageText"`
}

type FilePreview struct {
	URL         string `json:"url"`
	FileName    string `json:"fileName"`
	FileType    string `json:"fileType"`
	ViewerURL   string `json:"viewerUrl"`
	MessageText string `json:"messageText"`
}

type UploadedFile struct {
	FileName string `json:"fileName"`
	FileURL  string `json:"fileUrl"`
	FileType string `json:"fileType"`
	FileSize int64  `json:"fileSize"`
}

type Room struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"createdAt"`
	UserCount int    `json:"userCount"`
}

type ChatMessage struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	RoomID    string `json:"roomId"`
	Timestamp int64  `json:"timestamp"`
	UserID    string `json:"userId"`
}

type AuthPayload struct {
	Token string `json:"token"`
}

type JWTClaims struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var roomsLock = make(chan bool, 1)

var messagesLock = make(chan bool, 1)

const MAX_MESSAGES_PER_ROOM = 50

func init() {
	roomsLock <- true
	<-roomsLock
	messagesLock <- true
	<-messagesLock
}

func isDocumentFile(url string) bool {
	var (
		ext          = strings.ToLower(filepath.Ext(url))
		documentExts = []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx"}
	)

	return slices.Contains(documentExts, ext)
}

func getFileType(filename string) string {
	var ext = strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		return "pdf"
	case ".doc", ".docx":
		return "document"
	case ".xls", ".xlsx":
		return "spreadsheet"
	case ".ppt", ".pptx":
		return "presentation"
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return "image"
	case ".mp4", ".webm", ".ogg", ".avi", ".mov", ".wmv", ".flv", ".mkv":
		return "video"
	default:
		return "file"
	}
}

func createFilePreview(url, messageText string) *FilePreview {
	var (
		fileName  = filepath.Base(url)
		fileType  = getFileType(fileName)
		viewerURL string
	)

	switch fileType {
	case "pdf":
		viewerURL = url
	case "document", "spreadsheet", "presentation":
		viewerURL = fmt.Sprintf("https://view.officeapps.live.com/op/embed.aspx?src=%s", url)
	case "video":
		viewerURL = url
	default:
		viewerURL = url
	}

	return &FilePreview{
		URL:         url,
		FileName:    fileName,
		FileType:    fileType,
		ViewerURL:   viewerURL,
		MessageText: messageText,
	}
}

func createRoom(name, description string, db *gorm.DB) int64 {
	roomsLock <- true
	defer func() { <-roomsLock }()

	room := &models.Room{
		Name:        name,
		Description: description,
		Type:        "public",
	}

	if err := db.Model(&models.Room{}).Create(&room).Error; err != nil {
		fmt.Println("❌ Error creating room in database:", err)
		return 0
	}

	return room.Id
}

func getRoomsList(db *gorm.DB) []*Room {
	roomsLock <- true
	defer func() { <-roomsLock }()

	var response = make([]*Room, 0)

	if err := db.Model(&models.Room{}).Scopes(func(tx *gorm.DB) *gorm.DB { return tx.Order("created_at ASC") }).Scan(&response).Error; err != nil {
		fmt.Println("❌ Error fetching rooms from database:", err)
		return nil
	}

	return response
}

func joinRoom(roomID int64, db *gorm.DB) *Room {
	roomsLock <- true
	defer func() { <-roomsLock }()

	var room models.Room

	if err := db.Model(&models.Room{}).Where("id = ?", roomID).First(&room).Error; err != nil {
		fmt.Println("❌ Error fetching room from database:", err)
		return nil
	}

	room.UserCount++

	if err := db.Save(&room).Error; err != nil {
		fmt.Println("❌ Error updating room user count in database:", err)
		return nil
	}

	return &Room{
		ID:        roomID,
		Name:      room.Name,
		CreatedAt: room.CreatedAt,
		UserCount: room.UserCount,
	}
}

func leaveRoom(roomID int64, db *gorm.DB) {
	roomsLock <- true
	defer func() { <-roomsLock }()

	var room models.Room

	if err := db.Model(&models.Room{}).Where("id = ?", roomID).First(&room).Error; err != nil {
		fmt.Println("❌ Error fetching room from database:", err)
		return
	}

	if room.UserCount > 0 {
		room.UserCount--
	}

	if err := db.Save(&room).Error; err != nil {
		fmt.Println("❌ Error updating room user count in database:", err)
		return
	}
}

func saveMessage(roomID, userID int64, messageType, content string, db *gorm.DB) int64 {
	messagesLock <- true
	defer func() { <-messagesLock }()

	message := models.Message{
		Type:    messageType,
		Content: content,
		RoomId:  roomID,
		UserId:  userID,
	}

	if err := db.Model(&models.Message{}).Create(&message).Error; err != nil {
		fmt.Println("❌ Error saving message to database:", err)
		return 0
	}

	return message.Id
}

func getRoomHistory(roomID int64, db *gorm.DB) []*ChatMessage {
	messagesLock <- true
	defer func() { <-messagesLock }()

	var response = make([]*ChatMessage, 0)

	if err := db.Model(&models.Message{}).Where("room_id = ?", roomID).Order("created_at DESC").Scan(&response).Error; err != nil {
		return nil
	}

	return response
}

func socketIoHandle(io *socketio.Io, db *gorm.DB) {

	io.OnConnection(func(socket *socketio.Socket) {
		var currentRoom int64

		socket.On("connected", func(event *socketio.EventPayload) {
			var roomsList = getRoomsList(db)
			fmt.Println("New client connected:", socket.Id)
			fmt.Println("Rooms list sent to client:", roomsList)
			event.Socket.Emit("rooms list", roomsList)
			event.Socket.Emit("chat message", "🟢 Подключено к серверу")
		})

		socket.On("chat message", func(event *socketio.EventPayload) {
			if message, ok := event.Data[0].(string); ok {
				fmt.Println("Received message:", message)
				if currentRoom != 0 {
					saveMessage(currentRoom, 0, "text", message, db)
					fmt.Println("Saved message to DB:", message)
					socket.To(fmt.Sprintf("%d", currentRoom)).Emit("chat message", message)
				}
			}
		})

		socket.On("disconnecting", func(event *socketio.EventPayload) {
			println("disconnecting", socket.Nps, socket.Id)
		})

		socket.On("disconnect", func(event *socketio.EventPayload) {
			println("disconnect", socket.Nps, socket.Id)
		})

		socket.On("reaction", func(event *socketio.EventPayload) {
			if reactionData, ok := event.Data[0].(map[string]any); ok {
				if currentRoom != 0 {
					socket.To(fmt.Sprintf("%d", currentRoom)).Emit("reaction", reactionData)
				}
			}
		})

		socket.On("join room", func(event *socketio.EventPayload) {
			if roomData, ok := event.Data[0].(map[string]any); ok {
				var roomID = cast.ToString(roomData["roomId"])
				var roomIdInt = cast.ToInt64(roomData["roomId"])

				if currentRoom != 0 {
					socket.Leave(fmt.Sprintf("%d", currentRoom))
					leaveRoom(currentRoom, db)
				}

				var room = joinRoom(roomIdInt, db)
				if room != nil {
					currentRoom = roomIdInt
					socket.Join(roomID)
					socket.Emit("room joined", room)

					history := getRoomHistory(roomIdInt, db)
					if len(history) > 0 {
						socket.Emit("room history", history)
					}

					saveMessage(roomIdInt, 0, "text", fmt.Sprintf("👤 Пользователь присоединился к комнате %s", room.Name), db)
					socket.To(roomID).Emit("chat message", fmt.Sprintf("👤 Пользователь присоединился к комнате %s", room.Name))

					roomsList := getRoomsList(db)
					io.Emit("rooms list", roomsList)
				}
			}
		})

		socket.On("create room", func(event *socketio.EventPayload) {
			if roomData, ok := event.Data[0].(map[string]any); ok {
				var (
					name        = roomData["name"].(string)
					description = roomData["description"].(string)
				)

				var (
					newRoom   = createRoom(name, description, db)
					roomsList = getRoomsList(db)
				)

				io.Emit("rooms list", roomsList)
				socket.Emit("room created", newRoom)
			}
		})

		socket.On("get rooms", func(event *socketio.EventPayload) {
			var roomsList = getRoomsList(db)
			socket.Emit("rooms list", roomsList)
		})

		socket.On("file uploaded", func(event *socketio.EventPayload) {
			if fileData, ok := event.Data[0].(map[string]any); ok {
				fileRequest, err := unMarshal[UploadedFile](fileData)
				if err != nil {
					fmt.Println("❌ Error unmarshaling file data:", err)
					return
				}

				if currentRoom != 0 {
					if fileRequest.FileType == "image" {
						var imageData = map[string]any{
							"url":      fileRequest.FileURL,
							"fileName": fileRequest.FileName,
						}

						saveMessage(currentRoom, 0, "image", fileRequest.FileURL, db)
						socket.To(fmt.Sprintf("%d", currentRoom)).Emit("image message", imageData)

					} else if fileRequest.FileType == "video" {
						var videoData = map[string]any{
							"url":      fileRequest.FileURL,
							"fileName": fileRequest.FileName,
						}

						saveMessage(currentRoom, 0, "video", fileRequest.FileURL, db)
						socket.To(fmt.Sprintf("%d", currentRoom)).Emit("video message", videoData)

					} else if isDocumentFile(fileRequest.FileName) {
						var filePreview = createFilePreview(fileRequest.FileURL, fileRequest.FileName)

						saveMessage(currentRoom, 0, "file", filePreview.URL, db)
						socket.To(fmt.Sprintf("%d", currentRoom)).Emit("file preview", filePreview)
					} else {
						var filePreview = createFilePreview(fileRequest.FileURL, fileRequest.FileName)
						saveMessage(currentRoom, 0, "file", filePreview.URL, db)
						socket.To(fmt.Sprintf("%d", currentRoom)).Emit("file preview", filePreview)
					}
				}
			} else {
				fmt.Println("❌ Invalid file data received")
			}
		})
	})
}

func unMarshal[T any](data any) (T, error) {
	var result T

	dataByte, err := json.Marshal(data)
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(dataByte, &result); err != nil {
		return result, err
	}

	return result, nil
}

func usingWithGoFiber(db *gorm.DB) {
	if err := os.MkdirAll("uploads", 0755); err != nil {
		fmt.Printf("Ошибка создания папки uploads: %v\n", err)
	}

	app := fiber.New(fiber.Config{
		BodyLimit: 2 * 1024 * 1024 * 1024, // 2GB
	})

	app.Use(cors.New())

	var io = socketio.New()
	socketIoHandle(io, db)

	app.Static("/", "./")
	app.Static("/uploads", "./uploads")

	app.Post("/upload", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Не удалось получить файл",
			})
		}

		// Генерируем уникальное имя файла
		fileExt := filepath.Ext(file.Filename)
		fileName := uuid.New().String() + fileExt
		savePath := filepath.Join("uploads", fileName)

		// Сохраняем файл
		if err := c.SaveFile(file, savePath); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "❌ Не удалось сохранить файл",
			})
		}

		fileURL := fmt.Sprintf("http://localhost:3300/uploads/%s", fileName)

		uploadedFile := UploadedFile{
			FileName: file.Filename,
			FileURL:  fileURL,
			FileType: getFileType(file.Filename),
			FileSize: file.Size,
		}

		return c.JSON(uploadedFile)
	})

	app.Use(cors.New())

	app.Route("/socket.io", io.FiberRoute)

	fmt.Println("🚀 Fiber server listening on port 3300...")
	fmt.Println("📡 Socket.IO server ready for connections...")
	app.Listen(":3300")
}

func usingWithHttp(db *gorm.DB) {
	if err := os.MkdirAll("uploads", 0755); err != nil {
		fmt.Printf("Ошибка создания папки uploads: %v\n", err)
	}

	io := socketio.New()
	socketIoHandle(io, db)

	http.Handle("/socket.io/", io.HttpHandler())
	// http.Handle("/", http.FileServer(http.Dir("./")))

	fmt.Println(http.ListenAndServe(":3300", nil))
}

func usingWithEcho(db *gorm.DB) {
	if err := os.MkdirAll("uploads", 0755); err != nil {
		fmt.Printf("Ошибка создания папки uploads: %v\n", err)
	}

	io := socketio.New()
	socketIoHandle(io, db)

	fmt.Println("✅ Database connected and migrated")

	e := echo.New()
	// e.Use(middleware.CORS())

	e.Static("/", "./")
	e.Static("/uploads", "./uploads")

	e.Any("/socket.io/*", echo.WrapHandler(io.HttpHandler()))

	e.Logger.Fatal(e.Start(":3300"))
}

func main() {
	var databaseUrl = "postgres://postgres:postgres@localhost:5432/chat_service"

	fmt.Println("Connecting to database at", databaseUrl)

	db, err := gorm.Open(postgres.Open(databaseUrl), &gorm.Config{})
	if err != nil {
		fmt.Println("❌ Failed to connect to database:", err)
		return
	}

	if err := db.AutoMigrate(&models.Room{}, &models.Message{}); err != nil {
		fmt.Println("❌ Failed to migrate database:", err)
		return
	}

	fmt.Println("✅ Database connected and migrated")

	usingWithEcho(db)
}
