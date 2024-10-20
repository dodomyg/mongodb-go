package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dodomyg/go-crud/structs"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var collection *mongo.Collection

func main() {
	fmt.Println("Hello World!")

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	uri := os.Getenv("MONGO_URI")
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOpts)

	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB...")

	collection = client.Database("mydb").Collection("todos")

	app := fiber.New()
	port := os.Getenv("PORT")

	app.Get("/", getTodos)
	app.Post("/create", createTodo)
	app.Patch("/update/:id", updateTodo)
	app.Delete("/delete/:id", deleteTodo)

	if port == "" {
		port = "8000"
	}
	log.Fatal(app.Listen("0.0.0.0:" + port))

}

func getTodos(c *fiber.Ctx) error {
	var todos []structs.Todo

	cursor, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo structs.Todo
		err := cursor.Decode(&todo)

		if err != nil {
			return err
		}
		todos = append(todos, todo)
	}
	return c.JSON(todos)
}

func createTodo(c *fiber.Ctx) error {
	todo := new(structs.Todo)

	err := c.BodyParser(todo)

	if err != nil {
		return err
	}
	if todo.Title == "" {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Title is required", "data": nil})
	}

	res, _ := collection.InsertOne(context.Background(), todo)

	todo.Id = res.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(fiber.Map{"status": "success", "message": "Todo created successfully", "data": todo})
}

func updateTodo(c *fiber.Ctx) error {
	id := c.Params("id")

	objId, _ := primitive.ObjectIDFromHex(id)

	var myTodo structs.Todo
	filter := bson.M{"_id": objId}
	error := collection.FindOne(context.Background(), filter).Decode(&myTodo)
	if error != nil {
		return error
	}

	secFilter := bson.M{"_id": objId}

	update := bson.M{"$set": bson.M{"completed": !myTodo.Completed}}

	res, err := collection.UpdateOne(context.Background(), secFilter, update)

	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "No todo found with ID", "data": nil})
	}

	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "Todo updated successfully"})
}

func deleteTodo(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))

	if err != nil {
		return err
	}

	filter := bson.M{"_id": id}

	res, err := collection.DeleteOne(context.Background(), filter)

	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "No todo found with ID", "data": nil})
	}

	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "Todo deleted successfully"})
}
