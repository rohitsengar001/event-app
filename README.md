# Complete REST API in Go – Build an Event App (Gin, JWT, SQL, Swagger) |

In this tutorial, we will build a REST API in Go using the Gin framework. We will create a simple event app where users can sign up, log in, create events, delete events, edit events and attend events. We will use JWT authentication, authorization to protect routes, middleware, SQL, migrations, and Swagger documentation.

## Table of contents

- Open Table of contents
    - [Setting up the project](https://codingwithpatrik.dev/posts/rest-api-in-gin/#setting-up-the-project)
    - [Database Tables Overview](https://codingwithpatrik.dev/posts/rest-api-in-gin/#database-tables-overview)
    - [Migrations](https://codingwithpatrik.dev/posts/rest-api-in-gin/#migrations)
    - [Models](https://codingwithpatrik.dev/posts/rest-api-in-gin/#models)
    - [CRUD for Events](https://codingwithpatrik.dev/posts/rest-api-in-gin/#crud-for-events)
    - [Creating a User](https://codingwithpatrik.dev/posts/rest-api-in-gin/#creating-a-user)
        - [Explanation of the `registerUserHandler` Method](https://codingwithpatrik.dev/posts/rest-api-in-gin/#explanation-of-the-registeruserhandler-method)
    - [Swagger](https://codingwithpatrik.dev/posts/rest-api-in-gin/#swagger)
        - [Trying our api with swagger](https://codingwithpatrik.dev/posts/rest-api-in-gin/#trying-our-api-with-swagger)
    - [Conclusion](https://codingwithpatrik.dev/posts/rest-api-in-gin/#conclusion)

## Setting up the project

1.       
    
    **Enable live reload**:
    
    install air
    
    [https://github.com/air-verse/air](https://github.com/air-verse/air)
    
    If you are using zsh, you can add the following to your `.zshrc` file(for window set `Environment Path` of /go/bin ):
    
    ```
    export PATH=$PATH:$HOME/go/bin
    ```
    
    Create a `.air.toml` file in the root of the project. With the following content:
    
    ```
    root = "."
    testdata_dir = "testdata"
    tmp_dir = "tmp"
    [build]
    args_bin = []
    bin = "./tmp/main"
    cmd = "go build -o ./tmp/main ./cmd/api"
    delay = 1000
    exclude_dir = ["assets", "tmp", "vendor", "testdata"]
    exclude_file = []
    exclude_regex = ["_test.go"]
    exclude_unchanged = false
    follow_symlink = false
    full_bin = ""
    include_dir = []
    include_ext = ["go", "tpl", "tmpl", "html"]
    include_file = []
    kill_delay = "0s"
    log = "build-errors.log"
    poll = false
    poll_interval = 0
    post_cmd = []
    pre_cmd = []
    rerun = false
    rerun_delay = 500
    send_interrupt = false
    stop_on_error = false
    
    [color]
    app = ""
    build = "yellow"
    main = "magenta"
    runner = "green"
    watcher = "cyan"
    
    [log]
    main_only = false
    silent = false
    time = false
    
    [misc]
    clean_on_exit = false
    
    [proxy]
    app_port = 0
    enabled = false
    proxy_port = 0
    
    [screen]
    clear_on_rebuild = false
    keep_scroll = true
    
    ```
    

we will now be able to live reload the application with `air`.

1. **Initialize a new Go module**:

```
go mod init rest-api-in-gin
```

1. **Project Structure Setup**: 
    - Create a `cmd` directory at the root of your project. Inside `cmd`, add an `api` directory and place a `main.go` file within it.
    - At the root level, create an `internal` directory. Within `internal`, add a `database` directory.
    - Within the `internal` directory, create a `env` directory.
    - Within the `cmd` directory, create a `migrate` directory. Inside `migrate`, add a `main.go` file and a `migrations` directory.

Your project structure should look like this:

```
rest-api-in-gin
├── cmd
│   ├── api
│   │   ├── main.go
│   ├── migrate
│   │   ├── main.go
│   │   └── migrations
├── internal
│   ├── database
│   ├── env
```

## Database Tables Overview

Here is an overview of the tables we will be creating:

### Users Table

| Column | Description |
| --- | --- |
| **id** | Primary key, auto-incremented, unique identifier for each user. |
| **email** | Unique and required, email address of the user. |
| **name** | Required, the full name of the user. |
| **password** | Required, should be stored securely, used for user authentication. |

### Events Table

| Column | Description |
| --- | --- |
| **id** | Primary key, auto-incremented, unique identifier for each event. |
| **owner_id** | Foreign key referencing `users`, links an event to a user. |
| **name** | Required, the name of the event. |
| **description** | Required, a brief description of the event. |
| **date** | Required, the date when the event is scheduled to occur. |
| **location** | Required, the venue or place where the event will take place. |

### Attendees Table

| Column | Description |
| --- | --- |
| **id** | Primary key, auto-incremented, unique identifier for each attendee record. |
| **user_id** | Foreign key referencing `users`, links an attendee to a user. |
| **event_id** | Foreign key referencing `events`, links an attendee to an event. |

The `attendees` table links users to events, ensuring that each user and event exists. If a user or event is deleted, related attendee records are also removed automatically.

## Migrations 

This project uses golang-migrate for database migrations. First, install the migrate CLI:

- Prerequistive : c compiler and add environment variable  
- Golang migrate [https://github.com/golang-migrate/migrate/blob/master/cmd/migrate/README.md](https://github.com/golang-migrate/migrate/blob/master/cmd/migrate/README.md)

Add the following code to the `cmd/migrate/main.go` file:

```
package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a migration direction: 'up' or 'down'")
	}

	direction := os.Args[1]

	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	instance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatal(err)
	}

	fSrc, err := (&file.File{}).Open("cmd/migrate/migrations")
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithInstance("file", fSrc, "sqlite3", instance)
	if err != nil {
		log.Fatal(err)
	}

	switch direction {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
	default:
		log.Fatal("Invalid direction. Use 'up' or 'down'.")
	}
}
```

Here is a breakdown of the migration code:

- We Check for a migration direction (`up` or `down`) from command-line arguments.
- Connect to the SQLite database (`data.db`).
- Create a migration instance using the database connection and migration files.
- If the direction is “up,” apply all pending migrations to update the schema.
- If the direction is “down,” roll back the most recent migration.
- Log errors for invalid directions or failed migration operations, ignoring `ErrNoChange`.

Lets create our migrations by running the following commands:

```
migrate create -ext sql -dir ./cmd/migrate/migrations -seq create_users_table
migrate create -ext sql -dir ./cmd/migrate/migrations -seq create_events_table
migrate create -ext sql -dir ./cmd/migrate/migrations -seq create_attendees_table
```

This will create 6 files in the `cmd/migrate/migrations` folder. one up and one down for each migration.

Open up the `000001_create_users_table.up.sql` file and add the following code to the file:

```
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    password TEXT NOT NULL
);
```

Next open up the `000002_create_events_table.up.sql` file and add the following code to the file:

```
CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    date DATETIME NOT NULL,
    location TEXT NOT NULL,
    FOREIGN KEY (owner_id) REFERENCES users (id) ON DELETE CASCADE
);
```

An event has an owner id that references the user id, this will be used to restrict the events that a user can delete and update. If a user is deleted, all events created by that user will also be deleted.

The last migration file is the `000003_create_attendees_table.up.sql` file. Open it up and add the following code to the file:

```
CREATE TABLE IF NOT EXISTS attendees (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    event_id INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE
);
```

The `attendees` table links users to events, ensuring that each user and event exists. If a user or event is deleted, related attendee records are also removed automatically.

To every down file we need to add the following code:

```
-- 000001_create_users_table.down.sql
DROP TABLE IF EXISTS users;
```

```
-- 000002_create_events_table.down.sql
DROP TABLE IF EXISTS events;
```

```
-- 000003_create_attendees_table.down.sql
DROP TABLE IF EXISTS attendees;
```

We can now run the migrations by running the following command:

```
go run ./cmd/migrate/main.go up
```

This should now created a `data.db` file in the root of the project. We can view the database in a GUI like TablePlus. It would look something like this:

![](https://codingwithpatrik.dev/_astro/go-gin-api-tableplus.0Upo2JUj_1Ahk6z.webp.webp)

Go Gin API TablePlus

## Connecting our api app to the database

Open up the `main.go` file in the `cmd/api` folder and add the following code to the file:

```
package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
}
```

We open a connection to the database and check for errors.

We then use the `defer` keyword to close the database connection when the main function exits.

## Models

We will be creating 3 models:

1. `User`
2. `Event`
3. `Attendee`

Start by creating a `models.go` file in the `internal/database` folder.

With the following code:

```
package database

import "database/sql"

type Models struct {
	Users     UserModel
	Events    EventModel
	Attendees AttendeeModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:     UserModel{DB: db},
		Events:    EventModel{DB: db},
		Attendees: AttendeeModel{DB: db},
	}
}
```

Here we are creating a `Models` struct with 3 fields: `Users`, `Events`, and `Attendees`.

We are also creating a `NewModels` function that takes a `*sql.DB` instance as an argument and passes it to the `UserModel`, `EventModel`, and `AttendeeModel` structs.

Next we will create the `UserModel` struct. Create a `users.go` file in the `internal/database` folder.

```
package database

import "database/sql"

type UserModel struct {
	DB *sql.DB
}

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"-"`
}
```

- 
    
    The `UserModel` struct contains a `DB` field, which is a pointer to a `sql.DB` instance.
    
- 
    
    The `User` struct includes four fields: `Id`, `Email`, `Password`, and `Name`.
    
- 
    
    `json` tags are used to define how the struct fields are converted to and from JSON, ensuring proper data serialization and deserialization.
    
- 
    
    The `Password` field is marked with a `-` in the `json` tag, instructing the JSON package to exclude it from JSON responses, making sure we don’t expose the password in the response.
    

The next model we will create is the `EventModel` struct. Create a `events.go` file in the `internal/database` folder.

```
package database

import "database/sql"

type EventModel struct {
	DB *sql.DB
}

type Event struct {
	Id          int    `json:"id"`
	OwnerId     int    `json:"ownerId" binding:"required"`
	Name        string `json:"name" binding:"required,min=3"`
	Description string `json:"description" binding:"required,min=10"`
	Date        string `json:"date" binding:"required,datetime=2006-01-02"`
	Location    string `json:"location" binding:"required,min=3"`
}
```

- 
    
    The `Event` struct includes five fields: `Id`, `OwnerId`, `Name`, `Description`, `Date`, and `Location`.
    
- 
    
    We set binding tags and some validation rules. These will used later when creating an event and binding the request body to the `Event` struct. This is done by the Gin framework.
    
- 
    
    For now set a binding tag on the `OwnerId` field. Later we will remove it and instead use the current logged in user.
    

After that we will create the `AttendeeModel` struct. Create a `attendees.go` file in the `internal/database` folder.

```
package database

import "database/sql"

type AttendeeModel struct {
	DB *sql.DB
}

type Attendee struct {
	Id       int    `json:"id"`
	UserId   int    `json:"userId"`
	EventId  int    `json:"eventId"`
}
```

- 
    
    The `Attendee` struct includes three fields: `Id`, `UserId`, and `EventId`.
    
- 
    
    An attendee is a user that has signed up for an event. An event can have many attendees and an attendee can attend many events.
    

## Setting Up the Gin Server

1.    
    
    Create a `routes.go` file in the `cmd/api` folder.
    
    This file will define the routes for your Gin server.
    
    ```
    package main
    
    import (
    	   "net/http"
        "github.com/gin-gonic/gin"
    )
    
    func (app *application) routes() http.Handler {
        g := gin.Default()
        return g
    }
    ```
    
    We create a function `routes` that initializes a new Gin server instance using `gin.Default()`, which sets up some default middleware (like logging and recovery). Currently, it just returns the Gin instance. We will add some routes to this instance later.
    
2.    
    
    Create a `server.go` file in the `cmd/api` folder.
    
    This file will handle starting the HTTP server.
    
    ```
    package main
    
    import (
        "fmt"
        "log"
        "net/http"
        "time"
    )
    
    func serve(app *application) error {
    	server := &http.Server{
    		Addr:         fmt.Sprintf(":%d", app.port),
    		Handler:      app.routes(),
    		IdleTimeout:  time.Minute,
    		ReadTimeout:  10 * time.Second,
    		WriteTimeout: 30 * time.Second,
    	}
    
    	log.Printf("Starting server on port %d", app.port)
    
    	return server.ListenAndServe()
    }
    ```
    
    The `serve` function sets up an HTTP server with specific configurations like address, handler, and timeouts. It uses the `routes` function to get the handler (Gin instance) for the server. The server is started with `ListenAndServe`, if there is an error it will log the error and exit the program.
    
3.    
    
    Create a `env.go` file in the `internal/env` folder.
    
    Add the following code to the `env.go` file:
    
    ```
    package env
    
    import (
        "os"
        "strconv"
    )
    
    func GetEnvString(key, defaultValue string) string {
        if value, exists := os.LookupEnv(key); exists {
            return value
        }
        return defaultValue
    }
    
    func GetEnvInt(key string, defaultValue int) int {
        if value, exists := os.LookupEnv(key); exists {
            if intValue, err := strconv.Atoi(value); err == nil {
                return intValue
            }
        }
        return defaultValue
    }
    ```
    
    The `GetEnvString` and `GetEnvInt` functions are used to get the value of an environment variable. If the environment variable is not set, the function returns the default value.
    
4.     
    
    Put It All Together in the `main.go` in the `cmd/api` folder:
    
    ```
    package main
    
    import (
        "database/sql"
        "log"
        "rest-api-in-gin/internal/database"
        "rest-api-in-gin/internal/env"
        – "github.com/mattn/go-sqlite3"
        _ "github.com/joho/godotenv/autoload" // Automatically loads environment variables
    )
    
    type application struct {
        port   int
        jwtSecret string
        models database.Models
    }
    
    func main() {
    
        db, err := sql.Open("sqlite3", "./data.db")
        if err != nil {
            log.Fatal(err)
        }
        defer db.Close()
    
        models := database.NewModels(db)
    
        app := &application{
            port: env.GetEnvInt("PORT", 8080),
            jwtSecret: env.GetEnvString("JWT_SECRET", "some-secret-1213123"),
            models: models,
        }
    
        if err := serve(app); err != nil {
            log.Fatal(err)
        }
    }
    ```
    
    Here we load environment variables, initialize the database connection, create an `application` struct and start the server using the `serve` function.
    
    The application struct will be used to pass the dependencies around without having global variables.
    
    We then start the server using the `serve` function.
    

Now we can start the server by running the following command:

```
air
```

You should see the following output:

```
Starting server on port 8080
```

This means that the server is running and listening for incoming requests on port 8080.

## CRUD for Events

Currently we have no routes so let’s add some.

1.    
    
    **Set Up Event Routes:**
    
    Add routes to handle HTTP requests for event operations in your `routes.go` file.
    
    ```
    func (app *application) routes() http.Handler {
        g := gin.Default()
        v1 := g.Group("/api/v1")
        {
            v1.POST("/events", app.createEvent)
            v1.GET("/events", app.getAllEvents)
            v1.GET("/events/:id", app.getEvent)
            v1.PUT("/events/:id", app.updateEvent)
            v1.DELETE("/events/:id", app.deleteEvent)
        }
    
        return g
    }
    ```
    
    We define a route group `/api/v1` to version our API. Within this group, we map HTTP methods and paths to the corresponding handler functions for event operations. This structure helps organize routes and makes it easier to manage API versions.
    
2.                 
    
    **Implement Event Handlers:**
    
    Create a `events.go` file in the `cmd/api` folder and add the following methods.
    
    **Create Event**
    
    ```
    package main
    
    import (
    	"net/http"
    	"rest-api-in-gin/internal/database"
    	"strconv"
    
    	"github.com/gin-gonic/gin"
    )
    
    func (app *application) createEvent(c *gin.Context) {
    	var event database.Event
    	if err := c.ShouldBindJSON(&event); err != nil {
    		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    		return
    	}
    	err := app.models.Events.Insert(&event)
    	if err != nil {
    		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
    		return
    	}
    	c.JSON(http.StatusCreated, event)
    }
    ```
    
    This handler manages the creation of a new event. It binds the incoming JSON request body to an `Event` struct, validates the data, and calls the `Insert` method on the `EventModel` to add the event to the database. If successful, it returns a `201 Created` status with the created event data.
    
    **Get All Events**
    
    ```
    func (app *application) getAllEvents(c *gin.Context) {
        events, err := app.models.Events.GetAll()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
            return
        }
        c.JSON(http.StatusOK, events)
    }
    ```
    
    This handler retrieves all events. It calls the `GetAll` method on the `EventModel` to fetch all events from the database. If successful, it returns a `200 OK` status with the list of events.
    
    **Get Event**
    
    ```
    func (app *application) getEvent(c *gin.Context) {
    	id, err := strconv.Atoi(c.Param("id"))
    	if err != nil {
    		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
    		return
    	}
    	event, err := app.models.Events.Get(id)
    
    	if event == nil {
    		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
    		return
    	}
    
    	if err != nil {
    		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
    		return
    	}
    	c.JSON(http.StatusOK, event)
    }
    ```
    
    This handler retrieves a specific event by its ID. It extracts the event ID from the URL parameters, validates it, and calls the `Get` method on the `EventModel` to fetch the event from the database. If the event is found, it returns a `200 OK` status with the event data else it returns a `404 Not Found` status.
    
    **Update Event**
    
    ```
    func (app *application) updateEvent(c *gin.Context) {
    	id, err := strconv.Atoi(c.Param("id"))
    	if err != nil {
    		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
    		return
    	}
    
    	existingEvent, err := app.models.Events.Get(id)
    
    	if err != nil {
    		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
    		return
    	}
    
    	if existingEvent == nil {
    		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
    		return
    	}
    
    	updateEvent := &database.Event{}
    
    	if err := c.ShouldBindJSON(&updateEvent); err != nil {
    		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    		return
    	}
    
    	updateEvent.Id = id
    
    	if err := app.models.Events.Update(updateEvent); err != nil {
    		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
    		return
    	}
    
    	c.JSON(http.StatusOK, updateEvent)
    }
    ```
    
    This handler updates an existing event. It extracts and validates the event ID from the URL parameters, checks if the event exists, binds the incoming JSON request body to an `Event` struct, and calls the `Update` method on the `EventModel` to update the event in the database. If successful, it returns a `200 OK` status with the updated event data.
    
    **Delete Event**
    
    ```
    func (app *application) deleteEvent(c *gin.Context) {
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
            return
        }
        if err := app.models.Events.Delete(id); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
            return
        }
        c.JSON(http.StatusNoContent, nil)
    }
    ```
    
    This handler deletes a specific event by its ID. It extracts and validates the event ID from the URL parameters and calls the `Delete` method on the `EventModel` to remove the event from the database. If successful, it returns a `204 No Content` status.
    
3.                    
    
    **Implement Event Model Methods:**
    
    Define the methods for the `EventModel` to interact with the database. Open up the `events.go` file in the `database` folder. Update the imports and below the `Event` struct add the following methods.
    
    **Update imports**
    
    ```
    package database
    
    import (
        "database/sql"
        "context"
        "time"
    )
    ```
    
    **Insert Method**
    
    ```
    func (m EventModel) Insert(event *Event) error {
    	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    	defer cancel()
    
    	query := "INSERT INTO events (owner_id, name, description, date, location) VALUES ($1, $2, $3, $4, $5) RETURNING id"
    
    	err := m.DB.QueryRowContext(ctx, query, event.OwnerId, event.Name, event.Description, event.Date, event.Location).Scan(&event.Id)
    	if err != nil {
    		return err
    	}
    
    	return nil
    }
    ```
    
    This function inserts a new event into the `events` table. It uses `QueryRowContext`, which executes the query with a context that includes a 3-second timeout, ensuring the operation doesn’t hang indefinitely. If there is no error we add the id to the event and return `nil`.
    
    **GetAll Method**
    
    ```
     func (m EventModel) GetAll() ([]*Event, error) {
    	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    	defer cancel()
    
    	query := "SELECT * FROM events"
    
    	rows, err := m.DB.QueryContext(ctx, query)
    	if err != nil {
    		return nil, err
    	}
    	defer rows.Close()
    
    	events := []*Event{}
    
    	for rows.Next() {
    		var event Event
    		err := rows.Scan(&event.Id, &event.OwnerId, &event.Name, &event.Description, &event.Date, &event.Location)
    		if err != nil {
    			return nil, err
    		}
    		events = append(events, &event)
    	}
    
    	if err = rows.Err(); err != nil {
    		return nil, err
    	}
    
    	return events, nil
    }
    ```
    
    We retrieve all records from the `events` table. We then iterate over the result set and append each event to the `events` slice. If the query fails, we return an error.
    
    **Get Method**
    
    ```
    func (m EventModel) Get(id int) (*Event, error) {
    	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    	defer cancel()
    
    	query := "SELECT * FROM events WHERE id = $1"
    
    	row := m.DB.QueryRowContext(ctx, query, id)
    
    	var event Event
    
    	err := row.Scan(&event.Id, &event.OwnerId, &event.Name, &event.Description, &event.Date, &event.Location)
    	if err != nil {
    		if err == sql.ErrNoRows {
    			return nil, nil
    		}
    		return nil, err
    	}
    
    	return &event, nil
    }
    ```
    
    This function retrieves a specific record from the `events` table where the `id` matches the provided value. It maps the result to the `Event` struct fields. We check if the event is not found and return nil if it is not found.
    
    **Update Method**
    
    ```
    func (m EventModel) Update(event *Event) error {
    	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    	defer cancel()
    
    	query := "UPDATE events SET name = $1, description = $2, date = $3, location = $4 WHERE id = $5"
    
    	_, err := m.DB.ExecContext(ctx, query, event.Name, event.Description, event.Date, event.Location, event.Id)
    	if err != nil {
    		return err
    	}
    	return nil
    }
    ```
    
    This function updates an existing record in the `events` table. It uses the `SET` clause to specify the columns to be updated and their new values. It ensures only the record with the specified `id` is updated. If the update fails, it returns an error.
    
    **Delete Method**
    
    ```
    func (m EventModel) Delete(id int) error {
    	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    	defer cancel()
    
    	query := "DELETE FROM events WHERE id = $1"
    
    	_, err := m.DB.ExecContext(ctx, query, id)
    	if err != nil {
    		return err
    	}
    	return nil
    }
    
    ```
    
    Removes a record from the `events` table where the `id` matches the provided value. Returns an error if the deletion fails.
    
    You now have a complete CRUD functionality for events. Let’s test it.
    

## Creating a User

To be able to connect events with attendees and have events have an owner we need to create a user.

Start adding a new route to the `routes.go` file.

```
func (app *application) routes() http.Handler {
      ... rest of the routes
      v1.POST("/auth/register", app.registerUser)
}
```

We will group the routes under `/auth` and use the `POST` `register` method to register a new user, later we will add a `login` route.

Create a new handlers file called `auth.go` in the `cmd/api` folder and add the following code:

```
package main

import (
	"net/http"
	"rest-api-in-gin/internal/database"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required,min=2"`
}

func (app *application) registerUser(c *gin.Context) {
	var register registerRequest
	if err := c.ShouldBindJSON(&register); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(register.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}
	register.Password = string(hashedPassword)
	user := database.User{
		Email:    register.Email,
		Password: register.Password,
		Name:     register.Name,
	}
	err = app.models.Users.Insert(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

```

### Explanation of the `registerUserHandler` Method

The `registerUserHandler` function is responsible for handling user registration requests. Here’s a breakdown of its functionality:

1. 
    
    **Data Binding and Validation**: The function begins by binding the incoming JSON request body to a `registerRequest` struct. This ensures that the data is properly formatted and meets the required criteria, such as a valid email format and minimum password length.
    
2. 
    
    **Password Hashing**: To enhance security, the user’s password is hashed using the `bcrypt` library. This step is crucial as it ensures that the password is not stored in plain text in the database.
    
3. 
    
    **User Creation**: A new `User` instance is created with the provided email, hashed password, and name. This instance is then inserted into the database.
    
4. 
    
    **Response Handling**: If the user is successfully created, the function responds with a `201 Created` status and the user data. If any errors occur during the process, appropriate error messages are returned to the client.
    

Open `user.go` in the `database` package and update the imports and add the insert method to the `UserModel` struct.

```
import (
	"context"
	"database/sql"
	"time"
)
```

```
func (m *UserModel) Insert(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO users (email, password, name) VALUES ($1, $2, $3) RETURNING id`
	err := m.DB.QueryRowContext(ctx, stmt, user.Email, user.Password, user.Name).Scan(&user.Id)
	if err != nil {
		return err
	}
	return nil
}
```

Here we insert the user into the database and return an error if there is one.

## Testing Events and Users with Curl

We will use [Curl](https://curl.se/) a command line tool to test http requests.

Here are [all curl commands](https://raw.githubusercontent.com/patni1992/Rest-api-in-go-gin/refs/heads/main/curl.txt) that you can paste into your terminal. To test the events.

Let’s start by creating a new user.

```
curl -X POST http://localhost:8080/api/v1/auth/register \
-H "Content-Type: application/json" \
-d '{
  "email": "test@example.com",
  "password": "password",
  "name": "Test User"
}'
```

After you have created a user you can create an event.

```
curl -X POST http://localhost:8080/api/v1/events \
-H "Content-Type: application/json" \
-d '{
  "name": "Go Conference",
  "ownerId": 1,
  "description": "A conference about Go programming",
  "date": "2025-05-20",
  "location": "San Francisco"
}' \
-w "\nHTTP Status: %{http_code}\n"
```

If everything went well you should see the following output:

```
{
  "id": 1,
  "ownerId": 1,
  "name": "Go Conference",
  "description": "A conference about Go programming",
  "date": "2025-05-20",
  "location": "San Francisco"
}
HTTP Status: 201 Created
```

We can also retrieve all events.

```
curl -X GET http://localhost:8080/api/v1/events \
-H "Content-Type: application/json" \
-w "\nHTTP Status: %{http_code}\n"
```

To retrieve a specific event we can use the following command:

```
curl -X GET http://localhost:8080/api/v1/events/1 \
-H "Content-Type: application/json" \
-w "\nHTTP Status: %{http_code}\n"
```

Lets update the event.

```
curl -X PUT http://localhost:8080/api/v1/events/1 \
-H "Content-Type: application/json" \
-d '{
  "name": "Go Conference",
  "ownerId": 1,
  "description": "A conference about Go programming",
  "date": "2025-05-20",
  "location": "New York"
}' \
-w "\nHTTP Status: %{http_code}\n"
```

If you try to retrieve the event again you should see that the location has been updated.

The last thing we can do is delete the event.

```
curl -X DELETE http://localhost:8080/api/v1/events/1 \
-H "Content-Type: application/json" \
-w "\nHTTP Status: %{http_code}\n"
```

If you try to retrieve the event again you should get back a 404 not found error.

## Connecting Events with Attendees

We want users to be able to attend events.

1.   
    
    **Set Up Attendee Routes:**
    
    Add routes to handle HTTP requests for attendee operations in your `routes.go` file.
    
    ```
    func (app *application) routes() http.Handler {
        g := gin.Default()
        v1 := g.Group("/api/v1")
        {
            // ... rest of the routes ...
            v1.POST("/events/:id/attendees/:userId", app.addAttendeeToEvent)
            v1.GET("/events/:id/attendees", app.getAttendeesForEvent)
        }
    
        return g
    }
    ```
    
2.       
    
    **Implement Attendee Handlers:**
    
    Open up `events.go` in the `cmd/api` folder and add the following code:
    
    ```
    func (app *application) addAttendeeToEvent(c *gin.Context) {
    	eventId, err := strconv.Atoi(c.Param("id"))
    	if err != nil {
    		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
    		return
    	}
    
    	userId, err := strconv.Atoi(c.Param("userId"))
    	if err != nil {
    		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
    		return
    	}
    
    	event, err := app.models.Events.Get(eventId)
    	if err != nil {
    		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
    		return
    	}
    	if event == nil {
    		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
    		return
    	}
    
    	userToAdd, err := app.models.Users.Get(userId)
    	if err != nil {
    		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
    		return
    	}
    	if userToAdd == nil {
    		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
    		return
    	}
    
    	existingAttendee, err := app.models.Attendees.GetByEventAndAttendee(event.Id, userToAdd.Id)
    	if err != nil {
    		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve attendee"})
    		return
    	}
    	if existingAttendee != nil {
    		c.JSON(http.StatusConflict, gin.H{"error": "Attendee already exists"})
    		return
    	}
    
    	attendee := database.Attendee{
    		EventId: event.Id,
    		UserId:  userToAdd.Id,
    	}
    
    	_, err = app.models.Attendees.Insert(&attendee)
    	if err != nil {
    		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add attendee"})
    		return
    	}
    
    	c.JSON(http.StatusCreated, attendee)
    }
    ```
    
    - `c.Param("id")` and `c.Param("userId")` are used to extract URL parameters for the event and user IDs, respectively.
    - The function checks if the event and user exist in the database. If not, it returns a `404 Not Found` response.
    - It verifies if the attendee already exists for the event. If so, it returns a `409 Conflict` response.
    - The `Insert` method is called to add the attendee to the database if all checks pass.
    - If any operation fails, appropriate HTTP error responses are returned.
    
    Add a new handler function to get the attendees for an event.
    
    ```
    func (app *application) getAttendeesForEvent(c *gin.Context) {
    	id, err := strconv.Atoi(c.Param("id"))
    	if err != nil {
    		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
    		return
    	}
    
    	users, err := app.models.Attendees.GetAttendeesByEvent(id)
    	if err != nil {
    		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    		return
    	}
    
    	c.JSON(http.StatusOK, users)
    }
    ```
    
    - The method extracts the event ID from the URL parameters using `c.Param("id")` and converts it to an integer. If the conversion fails, it returns a `400 Bad Request` response indicating an invalid event ID.
    - It calls the `GetAttendeesByEvent` method from the `Attendees` model to fetch a list of users attending the specified event.
    - If an error occurs during data retrieval, a `500 Internal Server Error` response is returned with the error message.
    - If the data retrieval is successful, a `200 OK` response is returned along with the list of attendees.
3.                
    
    **Implement Attendee Model Methods:**
    
    Define the methods for the `AttendeeModel` in `attendees.go`.
    
    **Insert Method**
    
    Open `attendees.go` in the `database` folder and add the following code:
    
    Update the imports.
    
    ```
    import (
    "context"
    "database/sql"
    "time"
    )
    ```
    
    Add the insert method to the `AttendeeModel` struct.
    
    **Insert Method**
    
    ```
    package database
    
    func (m *AttendeeModel) Insert(attendee *Attendee) (*Attendee, error) {
    	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    	defer cancel()
    
    	query := `INSERT INTO attendees (event_id, user_id) VALUES ($1, $2) RETURNING id`
    	err := m.DB.QueryRowContext(ctx, query, attendee.EventId, attendee.UserId).Scan(&attendee.Id)
    
    	if err != nil {
    		return nil, err
    	}
    
    	return attendee, nil
    }
    
    ```
    
    Here we insert the attendee into the database with the provided user ID, event ID and return an error if there is one.
    
    **GetByEventAndAttendee Method**
    
    ```
    func (m *AttendeeModel) GetByEventAndAttendee(eventId, userId int) (*Attendee, error) {
    	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    	defer cancel()
    
    	query := `SELECT * FROM attendees WHERE event_id = $1 AND user_id = $2`
    	var attendee Attendee
    	err := m.DB.QueryRowContext(ctx, query, eventId, userId).Scan(&attendee.Id, &attendee.UserId, &attendee.EventId)
    	if err != nil {
    		if err == sql.ErrNoRows {
    			return nil, nil
    		}
    		return nil, err
    	}
    	return &attendee, nil
    }
    
    ```
    
    This method retrieves an attendee record from the database based on the provided event ID and user ID.
    
    **GetAttendeesByEvent Method**
    
    ```
    func (m AttendeeModel) GetAttendeesByEvent(eventId int) ([]User, error) {
    	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    	defer cancel()
    
    	query := `
         SELECT u.id, u.name, u.email
         FROM users u
         JOIN attendees a ON u.id = a.user_id
         WHERE a.event_id = $1
     `
    	rows, err := m.DB.QueryContext(ctx, query, eventId)
    	if err != nil {
    		return nil, err
    	}
    	defer rows.Close()
    
    	var users []User
    	for rows.Next() {
    		var user User
    		err := rows.Scan(&user.Id, &user.Name, &user.Email)
    		if err != nil {
    			return nil, err
    		}
    		users = append(users, user)
    	}
    	return users, nil
    }
    ```
    
    This method retrieves a list of users attending a specific event by joining the `users` and `attendees` tables.
    
4.    
    
    **Add get user by id method**
    
    Add the following method to the `UserModel` struct in `users.go`.
    
    ```
    func (m *UserModel) Get(id int) (*User, error) {
    	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    	defer cancel()
    
    	query := `SELECT * FROM users WHERE id = $1`
    
    	var user User
    	err := m.DB.QueryRowContext(ctx, query, id).Scan(&user.Id, &user.Email, &user.Name, &user.Password)
    	if err != nil {
    		if err == sql.ErrNoRows {
    			return nil, nil
    		}
    		return nil, err
    	}
    	return &user, nil
    }
    ```
    
    This method retrieves a user by their ID from the database.
    

## Delete Attendee from Event & Get Events for User

We can now add attendess to an event and retrieve the attendees for an event. However it would be nice if we could delete an attendee from an event and retrieve the events for a user.

1.   
    
    Routes Addition
    
    Add the following routes to handle new functionalities in your `routes.go` file.
    
    ```
    v1.DELETE("/events/:id/attendees/:userId", app.deleteAttendeeFromEvent)
    v1.GET("/attendees/:id/events", app.getEventsByAttendee)
    ```
    
2.     
    
    Handler Functions
    
    Implement the following handler functions in the `events.go` in the `cmd/api` folder.
    
    **Remove Attendee from Event:**
    
    ```
        func (app *application) deleteAttendeeFromEvent(c *gin.Context) {
        id, err := strconv.Atoi(c.Param("id"))
        if err != nil {
        	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
        	return
        }
    
        userId, err := strconv.Atoi(c.Param("userId"))
        if err != nil {
        	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        	return
        }
    
        err = app.models.Attendees.Delete(userId, id)
        if err != nil {
        	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete attendee"})
        	return
        }
    
        c.JSON(http.StatusNoContent, nil)
    
    }
    ```
    
    - 
        
        Extracts `id` (attendee ID) and `eventId` from the URL parameters.
        
    - 
        
        Validates the IDs and returns a `400 Bad Request` if they are invalid.
        
    - 
        
        Calls the `Delete` method on the `AttendeeModel` to remove the attendee.
        
    - 
        
        Returns a `204 No Content` status if the operation is successful, indicating that the request was successful but there is no content to send back.
        
        **Get Events for Attendee:**
        
        ```
        func (app *application) getEventsByAttendee(c *gin.Context) {
        	id, err := strconv.Atoi(c.Param("id"))
        	if err != nil {
        		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attendee ID"})
        		return
        	}
        
        	events, err := app.models.Events.GetByAttendee(id)
        	if err != nil {
        		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        		return
        	}
        	c.JSON(http.StatusOK, events)
        }
        ```
        
        - This function retrieves all events an attendee is attending.
        - It extracts `id` (attendee ID) from the URL parameters.
        - Validates the ID and returns a `400 Bad Request` if it is invalid.
        - Calls the `GetByAttendee` method on the `EventModel` to fetch the events.
        - Returns a `200 OK` status with the list of events if the operation is successful.
3.        
    
    Database Methods
    
    **Delete Method:**
    
    Open `attendees.go` in the `database` folder and add this method:
    
    ```
    func (m *AttendeeModel) Delete(userId, eventId int) error {
    	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    	defer cancel()
    
    	query := `DELETE FROM attendees WHERE user_id = $1 AND event_id = $2`
    	_, err := m.DB.ExecContext(ctx, query, userId, eventId)
    	if err != nil {
    		return err
    	}
    	return nil
    }
    ```
    
    This method deletes an attendee from an event with the provided user ID and event ID.
    
    **Get Events for Attendee:**
    
    ```
    func (m EventModel) GetByAttendee(attendeeId int) ([]Event, error) {
    	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    	defer cancel()
    
    	query := `
    		SELECT e.id, e.owner_id, e.name, e.description, e.date, e.location
    		FROM events e
    		JOIN attendees a ON e.id = a.event_id
    		WHERE a.user_id = $1
    	`
    	rows, err := m.DB.QueryContext(ctx, query, attendeeId)
    	if err != nil {
    		return nil, err
    	}
    	defer rows.Close()
    
    	var events []Event
    	for rows.Next() {
    		var event Event
    		err := rows.Scan(&event.Id, &event.OwnerId, &event.Name, &event.Description, &event.Date, &event.Location)
    		if err != nil {
    			return nil, err
    		}
    		events = append(events, event)
    	}
    	return events, nil
    }
    ```
    
    This method retrieves all events a user is attending with the provided attendee ID, joining the `events` and `attendees` tables to get the relevant data.
    

## Testing the API

We created a user before that we can use to test the API. Lets create a new event.

```
curl -X POST http://localhost:8080/api/v1/events -H "Content-Type: application/json" -d '{"name": "Test Event", "ownerId": 1, "description": "This is a test event", "date": "2025-10-01", "location": "Test Location"}' -w "\nHTTP Status: %{http_code}\n"
```

We can now add the user to the event. Take the id from the user and the event id. (Your ids may be different check the response from the previous requests)

```
curl -X POST http://localhost:8080/api/v1/events/1/attendees/1  -H "Content-Type: application/json" -w "\nHTTP Status: %{http_code}\n"
```

We should now get back an attendee this means the user has been added to the event. If we try the same request again we will get `{"error":"Attendee already exists"}`.

This will add the user to the event.

Lets get all the attendees for the event.

```
curl http://localhost:8080/api/v1/events/1/attendees
```

## Adding Authentication

Currently, anyone can create, update, and delete events. It would be nice if we could restrict these operations to only be performed by authenticated users.

1.   
    
    **Routes Addition**
    
    Add a new route in `routes.go` to handle the authentication.
    
    ```
    v1.POST("/auth/login", app.login)
    ```
    
2.        
    
    **Handler Functions**
    
    Add the following code to `auth.go` in the `cmd/api` folder to handle login.
    
    **Update Imports**
    
    ```
    import (
        "fmt"
        "net/http"
        "rest-api-in-gin/internal/database"
        "time"
    
        "github.com/gin-gonic/gin"
        "github.com/golang-jwt/jwt"
        "golang.org/x/crypto/bcrypt"
    )
    ```
    
    **Create Login Request, Response, and Handler**
    
    ```
    type loginRequest struct {
        Email    string `json:"email" binding:"required,email"`
        Password string `json:"password" binding:"required,min=8"`
    }
    
    type loginResponse struct {
        Token string `json:"token"`
    }
    
    func (app *application) login(c *gin.Context) {
    
        var auth loginRequest
        if err := c.ShouldBindJSON(&auth); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
    
        existingUser, err := app.models.Users.GetByEmail(auth.Email)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }
    
        err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(auth.Password))
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
            return
        }
    
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
            "userId": existingUser.Id,
            "exp":    time.Now().Add(time.Hour * 72).Unix(), // Token expires in 72 hours
        })
    
        tokenString, err := token.SignedString([]byte(app.jwtSecret))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
            return
        }
    
        c.JSON(http.StatusOK, loginResponse{Token: tokenString})
    }
    ```
    
    **Explanation of the Login Function**
    
    - The function begins by binding the incoming JSON request body to a `loginRequest` struct. This ensures that the data is properly formatted and meets the required criteria, such as a valid email format and a minimum password length of 8 characters.
    - It checks if the user exists in the database by calling the `GetByEmail` method on the `UserModel`. If the user is not found, a `404 Not Found` response is returned.
    - The function uses the `bcrypt` library to compare the provided password with the stored hashed password. If the passwords do not match, a `401 Unauthorized` response is returned.
    - Upon successful authentication, a JWT token is generated using the `jwt` library. The token includes the user’s ID and an expiration time (e.g., 72 hours from the time of issuance).
    - The generated token is returned to the client in a `200 OK` response. This token can then be used by the client to access protected routes.
3.    
    
    **Database Methods**
    
    Add `getByEmail` to the `UserModel`. Open `users.go` in the `database` and replace the `Get` method with the following code.
    
    ```
    func (m *UserModel) getUser(query string, args ...interface{}) (*User, error) {
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()
    
        var user User
        err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Id, &user.Email, &user.Name, &user.Password)
        if err != nil {
            if err == sql.ErrNoRows {
                return nil, nil
            }
            return nil, err
        }
        return &user, nil
    }
    
    func (m *UserModel) Get(id int) (*User, error) {
        query := `SELECT * FROM users WHERE id = $1`
        return m.getUser(query, id)
    }
    
    func (m *UserModel) GetByEmail(email string) (*User, error) {
        query := `SELECT * FROM users WHERE email = $1`
        return m.getUser(query, email)
    }
    ```
    
    Here we did some refactoring and created a new method called `getUser`, notice the `...interface{}` in the method signature. This allows us to pass in multiple arguments to the method. Then we have the `Get` and `GetByEmail` methods that we can use to get a user by id or email. This refactoring reduces code duplication and centralizes the logic for querying and handling user data.
    
4.      
    
    **Middleware**
    
    We can now add middleware to our API to restrict access to certain routes.
    
    Add a new file called `middleware.go` in `cmd/api` and add the following code:
    
    ```
    package main
    
    import (
        "net/http"
        "strings"
    
        "github.com/gin-gonic/gin"
        "github.com/golang-jwt/jwt"
    )
    
    func (app *application) AuthMiddleware() gin.HandlerFunc {
        return func(c *gin.Context) {
            authHeader := c.GetHeader("Authorization")
            if authHeader == "" {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
                c.Abort()
                return
            }
    
            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            if tokenString == authHeader {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token is required"})
                c.Abort()
                return
            }
    
            token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, jwt.ErrSignatureInvalid
                }
                return []byte(app.jwtSecret), nil
            })
    
            if err != nil || !token.Valid {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
                c.Abort()
                return
            }
    
            claims, ok := token.Claims.(jwt.MapClaims)
            if !ok {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
                c.Abort()
                return
            }
    
            userId := claims["userId"].(float64)
    
            user, err := app.models.Users.Get(int(userId))
            if err != nil {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
                c.Abort()
                return
            }
    
            c.Set("user", user)
    
            c.Next()
        }
    }
    ```
    
    **Explanation of the Middleware**
    
    - **Retrieve the Authorization Header**: The middleware starts by reading the `Authorization` header from the incoming request. This header should contain the JWT. If the header is missing, the middleware responds with a `401 Unauthorized` status and stops further processing by calling `c.Abort()`.
    - **Extract the Bearer Token**: The JWT is expected to be in the format `Bearer {token}`. The middleware removes the `Bearer`  prefix to extract the actual token. If the token is not in the expected format, it responds with a `401 Unauthorized` status and aborts the request.
    - **Parse and Validate the JWT**: The middleware uses the `jwt.Parse` function to decode and validate the token. It checks that the token’s signing method is HMAC, which is a common method for signing JWTs. The `jwtSecret` is used to verify the token’s signature, ensuring it hasn’t been tampered with.
    - **Handle Invalid Tokens**: If the token is invalid or an error occurs during parsing, the middleware responds with a `401 Unauthorized` status and aborts the request.
    - **Extract User Information**: If the token is valid, the middleware extracts the user ID from the token’s claims and retrieves the corresponding user from the database. The user is then set in the request context using `c.Set("user", user)`. This allows other handlers in the chain to access the authenticated user.
    - **Allow the Request to Proceed**: If the token is valid, the middleware calls `c.Next()`, allowing the request to proceed to the next handler in the chain.
5.    
    
    **Protect Routes**
    
    We can now add the middleware to our routes. Start by creating a new group of routes that we want to protect, then add the middleware to the group.
    
    Your code should now look like this:
    
    ```
    package main
    
    import (
        "net/http"
    
        "github.com/gin-gonic/gin"
    )
    
    func (app *application) routes() http.Handler {
    
        g := gin.Default()
        v1 := g.Group("/api/v1")
        {
            v1.GET("/events", app.getAllEvents)
            v1.GET("/events/:id", app.getEvent)
            v1.GET("/events/:id/attendees", app.getAttendeesForEvent)
            v1.GET("/attendees/:id/events", app.getEventsByAttendee)
    
            v1.POST("/register", app.registerUser)
            v1.POST("/login", app.login)
        }
    
        authGroup := v1.Group("/")
        authGroup.Use(app.AuthMiddleware())
        {
            authGroup.POST("/events", app.createEvent)
            authGroup.PUT("/events/:id", app.updateEvent)
            authGroup.DELETE("/events/:id", app.deleteEvent)
            authGroup.POST("/events/:id/attendees/:userId", app.addAttendeeToEvent)
            authGroup.DELETE("/events/:id/attendees/:userId", app.deleteAttendeeFromEvent)
        }
    
        return g
    }
    ```
    
6.            
    
    **Testing the API**
    
    We can now test the API by trying to create a new event without a valid token.
    
    ```
    curl -X POST http://localhost:8080/api/v1/events -H "Content-Type: application/json" -d '{"name": "Test Event", "ownerId": 1, "description": "This is a test event", "date": "2025-01-01", "location": "Test Location"}' -w "\nHTTP Status: %{http_code}\n"
    ```
    
    This should return a `401 Unauthorized` status. With the message `{"error":"Authorization header is required"}`
    
    **Login and get a token**
    
    We can now login and get a token.
    
    ```
    curl -X POST http://localhost:8080/api/v1/login -H "Content-Type: application/json" -d '{"email": "test@example.com", "password": "password"}' -w "\nHTTP Status: %{http_code}\n"
    ```
    
    This will return a token that we can use to authenticate our requests.
    
    **Use the Token**
    
    We can now use the token to create a new event. To add the token to the request we need to add it to the `Authorization` header. The format should be `-H "Authorization: Bearer {token}"`.
    
    ```
    curl -X POST http://localhost:8080/api/v1/events -H "Content-Type: application/json" -H "Authorization: Bearer {token}" -d '{"name": "Test Event", "ownerId": 1, "description": "This is a test event", "date": "2025-01-01", "location": "Test Location"}' -w "\nHTTP Status: %{http_code}\n"
    ```
    
    should now succeed and return a `201 Created` status.
    

## Adding Authorization

Currently a user can delete and update any event. We want to restrict this to only allow the user to do it if they are the owner of the event.

First we need to add a helper function to get the user from the context. Create a new file called `context.go` in `cmd/api` and add the following code:

```
package main

import (
	"rest-api-in-gin/internal/database"

	"github.com/gin-gonic/gin"
)

func (app *application) GetUserFromContext(c *gin.Context) *database.User {
	contextUser, exists := c.Get("user")
	if !exists {
		return &database.User{}
	}

	user, ok := contextUser.(*database.User)
	if !ok {
		return &database.User{}
	}

	return user
}
```

Here we are getting the user from the context and returning it. If the user is not found we return an empty user.

Let’s start with the handler for updating an event, open up `events.go` in `cmd/api` and replace the `updateEvent` method with the following code.

```
func (app *application) updateEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	user := app.GetUserFromContext(c)
	existingEvent, err := app.models.Events.Get(id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}

	if existingEvent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if existingEvent.OwnerId != user.Id {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this event"})
		return
	}

	updateEvent := &database.Event{
		Id: id,
	}

	if err := c.ShouldBindJSON(&updateEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := app.models.Events.Update(updateEvent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	c.JSON(http.StatusOK, updateEvent)
}
```

We get the user from the context and check if the user is the owner of the event. If they are not we return a `403 Forbidden` status.

Lets do the same for deleting an event, adding an attendee and deleting an attendee from an event.

**Delete Event**

```
func (app *application) deleteEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	user := app.GetUserFromContext(c)
	existingEvent, err := app.models.Events.Get(id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}
	if existingEvent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if user.Id != existingEvent.OwnerId {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	if err := app.models.Events.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
```

**Add Attendee to Event**

```
func (app *application) addAttendeeToEvent(c *gin.Context) {
	eventId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	event, err := app.models.Events.Get(eventId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}

	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	userToAdd, err := app.models.Users.Get(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	user := app.GetUserFromContext(c)

	if user.Id != event.OwnerId {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	if userToAdd == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	existingAttendee, err := app.models.Attendees.GetByEventAndAttendee(event.Id, userToAdd.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve attendee"})
		return
	}
	if existingAttendee != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Attendee already exists"})
		return
	}

	attendee := database.Attendee{
		EventId: event.Id,
		UserId:  userToAdd.Id,
	}

	_, err = app.models.Attendees.Insert(&attendee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add attendee"})
		return
	}

	c.JSON(http.StatusCreated, attendee)
}
```

**Delete Attendee from Event**

```
func (app *application) deleteAttendeeFromEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	event, err := app.models.Events.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}

	user := app.GetUserFromContext(c)

	if user.Id != event.OwnerId {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	err = app.models.Attendees.Delete(userId, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete attendee"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
```

We can also use the same method when creating an event. To set the owner of the event. Remmeber before we were setting the owner to the user id from the request.

**Create Event**

```
func (app *application) createEvent(c *gin.Context) {
	var event database.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := app.GetUserFromContext(c)
	event.OwnerId = user.Id

	err := app.models.Events.Insert(&event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, event)
}
```

We get the user from the context and set the owner of the event to the user id.

We also need to update our event model so binding for the owner id is not required. Open up `events.go` in `internal/database` and remove the `OwnerId` field from the `Event` struct.

The event model should now look like this:

```
type Event struct {
	Id          int    `json:"id"`
	Name        string `json:"name" binding:"required,min=3"`
	Description string `json:"description" binding:"required,min=10"`
	Date        string `json:"date" binding:"required,datetime=2006-01-02"`
	Location    string `json:"location" binding:"required,min=3"`
	OwnerId     int    `json:"ownerId"`
}
```

## Swagger

Swagger is an API documentation tool that allows you to generate interactive API documentation from your code.

To add Swagger to our API we can use the `gin-swagger` package.

Add the following to your `main.go` file in `cmd/api` import the docs and add the swagger comments.

```
import (
	"database/sql"
	"log"
	_ "rest-api-in-gin/docs"
	"rest-api-in-gin/internal/database"
	"rest-api-in-gin/internal/env"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

// @title Go Gin Rest API
// @version 1.0
// @description A rest API in Go using Gin framework.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your bearer token in the format **Bearer &lt;token&gt;**

// Apply the security definition to your endpoints
// @security BearerAuth

type application struct {
	port      int
	jwtSecret string
	models    database.Models
}

// -- rest of the code --

```

Open up `routes.go` and add the following route to serve the swagger.json file.

```
import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (app *application) routes() http.Handler {
	g := gin.Default()
	v1 := g.Group("/api/v1")

    ...rest of the routes...

	g.GET("/swagger/*any", func(c *gin.Context) {
		if c.Request.RequestURI == "/swagger/" {
			c.Redirect(302, "/swagger/index.html")
			return
		}
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("http://localhost:8080/swagger/doc.json"))(c)
	})

	return g
}

```

This code snippet sets up a route to serve the Swagger documentation. It also redirects the root `/swagger/` path to the Swagger UI. So it will be available at `http://localhost:8080/swagger/`.

Now we just need to document our handlers.

`events.go` will now look like this:

```
package main

import (
	"net/http"
	"rest-api-in-gin/internal/database"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

// GetEvents returns all events
//
//	@Summary		Returns all events
//	@Description	Returns all events
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	[]database.Event
//	@Router			/api/v1/events [get]
func (app *application) getAllEvents(c *gin.Context) {
	events, err := app.models.Events.GetAll()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retreive events"})
	}

	c.JSON(http.StatusOK, events)
}

// GetEvent returns a single event
//
//	@Summary		Returns a single event
//	@Description	Returns a single event
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Event ID"
//	@Success		200	{object}	database.Event
//	@Router			/api/v1/events/{id} [get]
func (app *application) getEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
	}

	event, err := app.models.Events.Get(id)

	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retreive event"})
	}

	c.JSON(http.StatusOK, event)
}

// CreateEvent creates a new event
//
//	@Summary		Creates a new event
//	@Description	Creates a new event
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Param			event	body		database.Event	true	"Event"
//	@Success		201		{object}	database.Event
//	@Router			/api/v1/events [post]
//	@Security		BearerAuth
func (app *application) createEvent(c *gin.Context) {
	var event database.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := app.GetUserFromContext(c)
	event.OwnerId = user.Id

	err := app.models.Events.Insert(&event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// UpdateEvent updates an existing event
//
//	@Summary		Updates an existing event
//	@Description	Updates an existing event
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Event ID"
//	@Param			event	body		database.Event	true	"Event"
//	@Success		200	{object}	database.Event
//	@Router			/api/v1/events/{id} [put]
//	@Security		BearerAuth
func (app *application) updateEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	user := app.GetUserFromContext(c)
	existingEvent, err := app.models.Events.Get(id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retreive event"})
		return
	}

	if existingEvent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if existingEvent.OwnerId != user.Id {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this event"})
		return
	}

	updatedEvent := &database.Event{}

	if err := c.ShouldBindJSON(updatedEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedEvent.Id = id

	if err := app.models.Events.Update(updatedEvent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	c.JSON(http.StatusOK, updatedEvent)
}

// DeleteEvent deletes an existing event
//
//	@Summary		Deletes an existing event
//	@Description	Deletes an existing event
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Event ID"
//	@Success		204
//	@Router			/api/v1/events/{id} [delete]
//	@Security		BearerAuth
func (app *application) deleteEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid event ID"})
	}

	user := app.GetUserFromContext(c)
	existingEvent, err := app.models.Events.Get(id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to retreive event"})
		return
	}

	if existingEvent == nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Event not found"})
		return
	}

	if existingEvent.OwnerId != user.Id {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this event"})
		return
	}

	if err := app.models.Events.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetAttendeesForEvent returns all attendees for a given event
//
//	@Summary		Returns all attendees for a given event
//	@Description	Returns all attendees for a given event
//	@Tags			attendees
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Event ID"
//	@Success		200	{object}	[]database.User
//	@Router			/api/v1/events/{id}/attendees [get]
func (app *application) getAttendeesForEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event id"})
		return
	}

	users, err := app.models.Attendees.GetAttendeesByEvent(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to to retreive attendees for events"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// AddAttendeeToEvent adds an attendee to an event
// @Summary		Adds an attendee to an event
// @Description	Adds an attendee to an event
// @Tags			attendees
// @Accept			json
// @Produce		json
// @Param			id	path		int	true	"Event ID"
// @Param			userId	path		int	true	"User ID"
// @Success		201		{object}	database.Attendee
// @Router			/api/v1/events/{id}/attendees/{userId} [post]
// @Security		BearerAuth
func (app *application) addAttendeeToEvent(c *gin.Context) {
	eventId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event Id"})
		return
	}

	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user Id"})
		return
	}

	event, err := app.models.Events.Get(eventId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retreive event"})
		return
	}
	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
	}

	userToAdd, err := app.models.Users.Get(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retreive user"})
		return
	}

	if userToAdd == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	}

	user := app.GetUserFromContext(c)

	if event.OwnerId != user.Id {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to add an attendee"})
		return
	}

	existingAttendee, err := app.models.Attendees.GetByEventAndAttendee(event.Id, userToAdd.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retreive attendee"})
		return
	}
	if existingAttendee != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Attendee already exists"})
		return
	}

	attendee := database.Attendee{
		EventId: event.Id,
		UserId:  userToAdd.Id,
	}

	_, err = app.models.Attendees.Insert(&attendee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add  attendee"})
		return
	}

	c.JSON(http.StatusCreated, attendee)

}

// GetEventsByAttendee returns all events for a given attendee
//
//	@Summary		Returns all events for a given attendee
//	@Description	Returns all events for a given attendee
//	@Tags			attendees
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Attendee ID"
//	@Success		200	{object}	[]database.Event
//	@Router			/api/v1/attendees/{id}/events [get]
func (app *application) getEventsByAttendee(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attendee id"})
		return
	}
	events, err := app.models.Attendees.GetEventsByAttendee(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get events"})
		return
	}

	c.JSON(http.StatusOK, events)
}

// DeleteAttendeeFromEvent deletes an attendee from an event
// @Summary		Deletes an attendee from an event
// @Description	Deletes an attendee from an event
// @Tags			attendees
// @Accept			json
// @Produce		json
// @Param			id	path		int	true	"Event ID"
// @Param			userId	path		int	true	"User ID"
// @Success		204
// @Router			/api/v1/events/{id}/attendees/{userId} [delete]
// @Security		BearerAuth
func (app *application) deleteAttendeeFromEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event id"})
		return
	}

	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}

	event, err := app.models.Events.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	user := app.GetUserFromContext(c)
	if event.OwnerId != user.Id {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete an attendeeFromEvent"})
		return
	}

	err = app.models.Attendees.Delete(userId, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete attendee"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
```

The `auth.go` file will now look like this:

```
package main

import (
	"net/http"
	"rest-api-in-gin/internal/database"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required,min=2"`
}

// RegisterUser registers a new user
// @Summary		Registers a new user
// @Description	Registers a new user
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			user	body		registerRequest	true	"User"
// @Success		201	{object}	database.User
// @Router			/api/v1/auth/register [post]
func (app *application) registerUser(c *gin.Context) {
	var register registerRequest

	if err := c.ShouldBindJSON(&register); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(register.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Something went wrong"})
		return
	}

	register.Password = string(hashedPassword)
	user := database.User{
		Email:    register.Email,
		Password: register.Password,
		Name:     register.Name,
	}

	err = app.models.Users.Insert(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login logs in a user
//
//	@Summary		Logs in a user
//	@Description	Logs in a user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			user	body	loginRequest	true	"User"
//	@Success		200	{object}	loginResponse
//	@Router			/api/v1/auth/login [post]
func (app *application) login(c *gin.Context) {
	var auth loginRequest
	if err := c.ShouldBindJSON(&auth); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingUser, err := app.models.Users.GetByEmail(auth.Email)
	if existingUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(auth.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": existingUser.Id,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(app.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{Token: tokenString})

}
```

Below are the Swagger tags used in the code and their purposes:

- `@Summary`: A short description of the API endpoint.
- `@Description`: A detailed description of the API endpoint.
- `@Tags`: The category of the API endpoint.
- `@Accept`: The media type the API accepts.
- `@Produce`: The media type the API produces.
- `@Param`: The parameters of the API endpoint.
- `@Success`: The expected response of the API endpoint.
- `@Router`: The HTTP method and path of the API endpoint.
- `@Security`: The security scheme used for the API endpoint.

cmd: ```go install github.com/swaggo/swag/cmd/swag@latest```   
Run `swag init --dir cmd/api --parseDependency --parseInternal --parseDepth 1` to generate the Swagger documentation.

If you open up localhost:8080/swagger/index.html in your browser, you will see the Swagger UI with all the API endpoints.

![](https://codingwithpatrik.dev/_astro/go-gin-swagger.GAEm3Tps_QWSCt.webp.webp)

Swagger Overview

### Trying our api with swagger [docs]("https://github.com/swaggo/swag/blob/master/README.md#declarative-comments-format)

Go to the login endpoint and click the button `Try it out` Fill in the email and password and click `Execute`.

![](https://codingwithpatrik.dev/_astro/go-swagger-login.ZhaYPaZv_Z1N7caG.webp.webp)

Swagger Login

You will get a token in the response.

![](https://codingwithpatrik.dev/_astro/go-swagger-login-response.N6U6IsyR_2d8jzb.webp.webp)

Swagger Login Response

Copy the token and scroll upp to The Authorization button and click it.

Write `Bearer` in the `Value` field and paste the token after it and click `Authorize`.

![](https://codingwithpatrik.dev/_astro/go-swagger-authorization.rB52q978_1IT9IR.webp.webp)

Swagger Authorization

You are now logged in and can try any endpoint that requires authorization.

## Conclusion

In this tutorial, we successfully built a REST API in Go using the Gin framework. Our project included event management system that allows for the creation and management of users, events, and attendees. We implemented JWT-based authentication to secure our API, ensuring that only authorized users can perform certain actions.

Additionally, we enhanced our API with Swagger documentation, providing clear and interactive descriptions of our API endpoints.

We organized our code by creating models, handlers, routes, and middleware, each responsible for specific aspects of the application, ensuring a clean and maintainable codebase.
