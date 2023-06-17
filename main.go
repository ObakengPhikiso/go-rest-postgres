package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ObakengPhikiso/go-rest-postgres/models"
	"github.com/ObakengPhikiso/go-rest-postgres/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
}

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) CreateBook(context *fiber.Ctx) error {
	book := Book{}

	err := context.BodyParser(&book)

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{"message": "Request failed"})
		return err
	}

	err = r.DB.Create(&book).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "Error creating a book"})
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "Book created successfully"})
	return nil
}

func (r *Repository) UpdateBook(context *fiber.Ctx) error {
	book := models.Books{}
	err := context.BodyParser(&book)

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{"message": "Request failed"})
		return err
	}

	err = r.DB.Save(&book).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "Error updating a book"})
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "Book updated successfully"})
	return nil

}

func (r *Repository) DeleteBook(context *fiber.Ctx) error {
	bookModel := models.Books{}
	id := context.Params("id")

	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": "The book id is required"})
		return nil
	}

	err := r.DB.Delete(&bookModel, id).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "Could not delete book"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "Book deleted successfully"})
	return nil

}

func (r *Repository) GetBookByID(context *fiber.Ctx) error {
	bookModel := models.Books{}

	id := context.Params("id")

	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": "The book id is required"})
		return nil
	}

	err := r.DB.Where("id = ?", id).First(&bookModel).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "Could not find book"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "Book fetched successfully", "data": bookModel})
	return nil
}

func (r *Repository) GetBooks(context *fiber.Ctx) error {
	bookModels := &[]models.Books{}

	err := r.DB.Find(&bookModels).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "Could not get books"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "Books fetched successfully", "data": bookModels})
	return nil
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1")
	api.Post("/create_books", r.CreateBook)
	api.Put("/update_book", r.UpdateBook)
	api.Delete("/delete_book/:id", r.DeleteBook)
	api.Get("/get_book/:id", r.GetBookByID)
	api.Get("/books", r.GetBooks)
}

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASSWORD"),
		User:     os.Getenv("DB_USER"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSL_MODE"),
	}

	db, err := storage.NewConnection(config)

	if err != nil {
		log.Fatal("Couldn't connect to database")
	}

	err = models.MigrateBooks(db)

	if err != nil {
		log.Fatal("Could migrate the database")
	}
	r := Repository{
		DB: db,
	}
	app := fiber.New()

	r.SetupRoutes(app)
	app.Listen(":8080")

}
