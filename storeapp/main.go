package main

import (
	"log"
	"os"
	"strings"

	"github.com/eonebyte/myapp/app/config"
	"github.com/eonebyte/myapp/app/handler"
	"github.com/eonebyte/myapp/app/controllers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

    dbHost :=   os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbUser := os.Getenv("DB_USERNAME")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")
    
    // Buat objek dbConfig dengan nilai-nilai dari environment variables
    dbConfig := config.DatabaseConfig{
        Host:     dbHost,
        Port:     dbPort,
        Username: dbUser,
        Password: dbPassword,
        DBName:   dbName,
    }

    dsn := dbConfig.GetDSN()
	config.InitDatabase(dsn)

	// Get the value of the BASE_URL environment variable
	baseUrl := os.Getenv("BASE_URL")

	// Initialize standard Go html template engine
	engine := html.New("./app/views", ".html")
	engine.AddFuncMap(fiber.Map{
		"addIndex": func(a, b int) int { return a + b },
		"seq": func(max int) []int {
			seq := make([]int, max)
			for i := 1; i <= max; i++ {
				seq[i-1] = i
			}
			return seq
		},
	})
	// If you want other engine, just replace with following
	// Create a new engine with django
	// engine := django.New("./views", ".django")

	app := fiber.New(fiber.Config{

		Views: engine,
	})

	//Middleware helmet
	// app.Use(helmet.New())

	//Middleware logger
	app.Use(logger.New())

	//Session Declaration
	store := session.New()
	

	// Initialize default config
	// Or extend your config for customization
	app.Use(cors.New(cors.Config{
		Next:             nil,
		AllowOriginsFunc: nil,
		AllowOrigins:     "*",
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodHead,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
		}, ","),
		AllowHeaders:     "",
		AllowCredentials: false,
		ExposeHeaders:    "",
		MaxAge:           0,
	}))

	app.Static("/", "./static")

	app.Use(handler.CheckDatabaseInstallation)
	app.Get("/install", controllers.PageInstall)
	app.Post("/handle-install", controllers.HandleInstall)

	//wiring of controllers
	homeController := controllers.NewHomeController(store, baseUrl)
	authController := controllers.NewAuthController(store, baseUrl)
	paymentController := controllers.NewPaymentController(store, baseUrl)
	categoryController := controllers.NewCategoryController(store, baseUrl)
	itemController := controllers.NewItemController(store, baseUrl)
	bannerController := controllers.NewBannerController(store, baseUrl)
	userController := controllers.NewUserController(store, baseUrl)
	settingController := controllers.NewSettingController(store, baseUrl)
	orderController := controllers.NewOrderController(store, baseUrl)
	shippingController := controllers.NewShippingController(store, baseUrl)

	

	app.Get("/", homeController.Home)


	//Group: Authentication
	auth := app.Group("/auth")
	auth.Get("", authController.AuthPage)
	auth.Post("/login_post", authController.HandleLogin)
	auth.Post("/register_post", authController.HandleRegister)
	auth.Get("/logout", authController.Logout)

	//Group : Admin - Auth
	admin := app.Group("/dashboard")
	admin.Get("", authController.Dashboard)
	admin.Get("/profil-user", userController.ProfilUser)
	admin.Post("/profil-user-update", userController.UpdateProfilUser)
	admin.Get("/admin", authController.DashboardAdmin)
	admin.Get("/orders", authController.Orders)
	admin.Get("/keranjang", authController.Keranjang)
	admin.Get("/checkout", authController.CheckoutPage)

	

	//Group : Admin - Manage Category
	admin.Get("/categories", categoryController.PageCategory)
	admin.Post("/create_category", categoryController.CreateCategory)
	admin.Post("/update_category", categoryController.UpdateCategory)
	admin.Post("/delete_category", categoryController.DeleteCategory)

	//Group : Admin - Manage Items/Products
	admin.Get("/products", itemController.ItemPage)
	admin.Post("/create_item", itemController.CreateItem)
	admin.Post("/update_item", itemController.UpdateItem)
	admin.Post("/delete_item", itemController.DeleteItem)

	//Group : Admin - Manage Banners
	admin.Get("/banners", bannerController.PageBanner)
	admin.Post("/create_banner", bannerController.CreateBanner)
	admin.Post("/update_banner", bannerController.UpdateBanner)
	admin.Post("/delete_banner", bannerController.DeleteBanner)

	//Group : Admin - Manage Users
	admin.Get("/users", userController.PageUser)
	admin.Post("/update_user", userController.UpdateUser)
	admin.Post("/delete_user", userController.DeleteUser)

	//Group : Admin - Manage Admin Orders
	admin.Get("/admin-orders", orderController.AdminOrders)
	admin.Post("/atur_shipping", orderController.AturPengiriman)
	admin.Post("/delete_order", orderController.DeleteOrder)

	//Group : Admin - Manage Shippings
	admin.Get("/shippings", shippingController.PageShipping)
	admin.Post("/update_shipping", shippingController.UpdateShipping)

	//Group : Admin - Manage Settings
	admin.Get("/settings", settingController.PageSetting)
	admin.Post("/update_setting", settingController.UpdateSetting)

	//Group : Admin - Payment
	admin.Post("/payment", paymentController.Payment)
	admin.Get("/notifikasi", paymentController.Notifikasi)
	admin.Get("/api-notifikasi", paymentController.ApiNotifikasi)

	//Group: Api
	api := app.Group("/api")
	api.Get("/getCost/:destination", controllers.GetCost)
	api.Get("/getProvider/:destination/:courier/", controllers.GetProvider)

	app.Get("/item/:id", homeController.ProdukDetail)

	app.Get("/shopping", func(c *fiber.Ctx) error {
		return c.Render("shopping", fiber.Map{})
	})

	// Read the port from the environment variable or use a default value
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	// Start server
	err = app.Listen("localhost:"+port)
	if err != nil {
		return
	}

}








