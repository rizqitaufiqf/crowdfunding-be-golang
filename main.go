package main

import (
	"crowdfunding/auth"
	"crowdfunding/campaign"
	"crowdfunding/handler"
	"crowdfunding/helper"
	"crowdfunding/user"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	// load env file
	envErr := godotenv.Load()
	if envErr != nil {
		log.Fatal("Error loading .env file")
	}
	// setup env values to variable
	host := os.Getenv("DATABASE_HOST")
	username := os.Getenv("DATABASE_USERNAME")
	password := os.Getenv("DATABASE_PASSWORD")
	databaseName := os.Getenv("DATABASE_NAME")
	port := os.Getenv("DATABASE_PORT")
	sslMode := os.Getenv("DATABASE_SSL_ENABLED")
	timezone := os.Getenv("TIMEZONE")

	// generate connection string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s", host, username, password, databaseName, port, sslMode, timezone)
	// connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("Connection to database successfully")

	// init Repository
	userRepository := user.NewRepository(db)
	campaignRepository := campaign.NewRepository(db)
	// init Service
	userService := user.NewService(userRepository)
	authService := auth.NewService()
	// init Handler(Controller)
	userHandler := handler.NewUserHandler(userService, authService)

	//* Test Campaign Repository
	fmt.Println("TEST")
	fmt.Println("TEST")
	campaigns, err := campaignRepository.FindAll()
	if err != nil {
		fmt.Println("ERROR", err)
	}
	for _, camp := range campaigns {
		fmt.Println(camp.Name)
		if len(camp.CampaignImages) > 0 {
			fmt.Println("\t", camp.CampaignImages[0].FileName)
		}
	}
	fmt.Println("TEST")
	fmt.Println(campaignRepository.FindAllByID("36c96295-4b5f-4a81-8cc5-4ee61e3bb3e2"))
	//* End of Test Campaign Repository

	// init router
	router := gin.Default()
	// group router endpoint
	api := router.Group("/api/v1")
	// generate endpoint
	api.POST("/users", userHandler.RegisterUser)
	api.POST("/login", userHandler.Login)
	api.POST("/check-email", userHandler.CheckEmail)
	api.POST("/avatars", authMiddleware(authService, userService), userHandler.UploadAvatar)
	// run web service
	err = router.Run("localhost:8080")
	if err != nil {
		log.Fatal(err.Error())
	}

}

func authMiddleware(authService auth.Service, userService user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.Contains(authHeader, "Bearer") {
			response := helper.APIResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			// we use AbortWithStatusJSON because this function is in middleware, so after reject will not go to
			// the next process
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		tokenString := strings.Split(authHeader, " ")[1]
		token, err := authService.ValidateToken(tokenString)
		if err != nil {
			response := helper.APIResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		claim, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			response := helper.APIResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		userID := claim["user_id"].(string)
		usr, err := userService.GetUserByID(userID)
		if err != nil {
			response := helper.APIResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		c.Set("currentUser", usr)
	}
}
