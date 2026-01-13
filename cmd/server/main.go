package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mullayam/go-tcp-chat/config"
	"github.com/mullayam/go-tcp-chat/internal/auth"
	"github.com/mullayam/go-tcp-chat/internal/room"
	"github.com/mullayam/go-tcp-chat/internal/server"
	"github.com/mullayam/go-tcp-chat/internal/session"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println("Starting TCP Chat Server...")
	log.Printf("Configuration loaded:")
	log.Printf("  - TCP Port: %s", cfg.TCPPort)
	log.Printf("  - SMTP Host: %s:%d", cfg.SMTPHost, cfg.SMTPPort)
	log.Printf("  - SMTP Email: %s", cfg.SMTPEmail)
	log.Printf("  - OTP Expiration: %d minutes", cfg.OTPExpirationMinutes)
	log.Printf("  - OTP Max Retries: %d", cfg.OTPMaxRetries)
	log.Printf("  - Username Length: %d-%d characters", cfg.UsernameMinLength, cfg.UsernameMaxLength)

	// Initialize managers
	sessionMgr := session.NewManager(cfg.UsernameMinLength, cfg.UsernameMaxLength)
	roomMgr := room.NewManager()
	otpService := auth.NewOTPService(cfg.OTPExpirationMinutes, cfg.OTPMaxRetries)
	emailService := auth.NewEmailService(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPEmail, cfg.SMTPPassword)

	// Create TCP server
	tcpServer := server.NewTCPServer(
		cfg.TCPPort,
		sessionMgr,
		roomMgr,
		otpService,
		emailService,
	)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nShutting down server...")
		if err := tcpServer.Stop(); err != nil {
			log.Printf("Error stopping server: %v", err)
		}
		os.Exit(0)
	}()

	// Start server
	if err := tcpServer.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
