package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"poker-planning/internal/database"
	"poker-planning/internal/handlers"
	"poker-planning/internal/services"
	"poker-planning/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Get port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Get database path from environment variable or default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "poker.db"
	}

	db, err := database.NewDB(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	userService := services.NewUserService(db.DB)
	sessionService := services.NewSessionService(db.DB)
	votingService := services.NewVotingService(db.DB)
	ticketService := services.NewTicketService(db.DB)
	wsService := services.NewWSService()
	go wsService.Run() // Start the WebSocket service

	h := handlers.NewHandler(userService, sessionService, votingService, ticketService, wsService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(utils.RecoverFromPanic)
	r.Use(middleware.Compress(5))
	r.Use(middleware.Timeout(30 * time.Second)) // Add timeout middleware
	r.Use(handlers.SessionMiddleware(userService))

	r.Get("/", h.Home)
	r.Post("/set-username", h.SetUsername)
	
	r.Route("/session", func(r chi.Router) {
		r.Post("/create", h.CreateSession)
		r.Get("/{sessionID}", h.GetSession)
		r.Get("/{sessionID}/partial", h.GetSessionPartial)
		r.Post("/{sessionID}/join", h.JoinSession)
		r.Post("/{sessionID}/tickets", h.CreateTicket)
		r.Delete("/{sessionID}/tickets/{ticketID}", h.DeleteTicket)
		r.Post("/{sessionID}/start-voting", h.StartVoting)
		r.Post("/{sessionID}/end-voting", h.EndVoting)
		r.Post("/{sessionID}/next-ticket", h.NextTicket)
		r.Post("/{sessionID}/select-ticket/{ticketID}", h.SelectTicket)
		r.Post("/{sessionID}/vote", h.SubmitVote)
		r.Get("/{sessionID}/ws", h.WebSocketHandler)
		r.Post("/{sessionID}/leave", h.LeaveSession)
		r.Delete("/{sessionID}", h.DeleteSession)
		r.Post("/{sessionID}/review", h.ReviewSession)
		r.Get("/{sessionID}/summary", h.GetSessionSummary)
	})

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}