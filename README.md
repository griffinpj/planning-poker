# Sprint Planning Poker

A real-time web application for agile sprint planning poker sessions built with Go, HTMX, and Tailwind CSS.

## Features

- **Session Management**: Create and join planning sessions with unique URLs
- **Real-time Updates**: Server-Sent Events (SSE) for live collaboration
- **Voting System**: Fibonacci sequence cards (0, 1, 2, 3, 5, 8, 13, 21, 34) and special cards (☕, ?)
- **Ticket Management**: Add, edit, and organize tickets for estimation
- **Emoji Reactions**: Send animated emoji reactions to team members
- **Responsive Design**: Works on desktop, tablet, and mobile devices
- **Session-based Authentication**: No persistent accounts required

## Technology Stack

- **Backend**: Go with Chi router
- **Database**: SQLite with migrations via Goose
- **Frontend**: HTMX for dynamic interactions
- **Styling**: Tailwind CSS with Material Design principles
- **Real-time**: Server-Sent Events (SSE)

## Prerequisites

- Go 1.21 or higher
- SQLite

## Quick Start

1. **Clone the repository** (or use the existing code)

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Run the application**:
   ```bash
   go run cmd/server/main.go
   ```

4. **Open your browser** and navigate to `http://localhost:8080`

## Project Structure

```
poker-planning/
├── cmd/server/          # Application entry point
├── internal/
│   ├── database/        # Database connection and migrations
│   ├── handlers/        # HTTP request handlers
│   ├── models/          # Data models
│   ├── services/        # Business logic
│   └── utils/           # Utility functions
├── migrations/          # Database migration files
├── static/              # Static assets (CSS, JS)
├── templates/           # HTML templates
└── README.md
```

## API Endpoints

### Main Routes
- `GET /` - Home page
- `POST /set-username` - Set user display name

### Session Routes
- `POST /session/create` - Create new session
- `GET /session/{id}` - Join/view session
- `GET /session/{id}/events` - SSE endpoint for real-time updates

### Session Management
- `POST /session/{id}/tickets` - Create ticket
- `DELETE /session/{id}/tickets/{ticketId}` - Delete ticket
- `POST /session/{id}/start-voting` - Start voting round
- `POST /session/{id}/end-voting` - End voting and reveal results
- `POST /session/{id}/next-ticket` - Advance to next ticket
- `POST /session/{id}/vote` - Submit vote
- `POST /session/{id}/emoji` - Send emoji reaction

## Usage

### Creating a Session

1. Enter your name when prompted
2. Click "Create New Session"
3. Enter a session name
4. Share the session URL with your team

### Planning Process

1. **Add Tickets**: Session owner can add tickets to estimate
2. **Start Voting**: Owner starts voting for current ticket
3. **Vote**: All participants select estimation cards
4. **Reveal Results**: Owner ends voting to show all votes
5. **Next Ticket**: Move to the next ticket when ready

### Voting Cards

- **Fibonacci Numbers**: 0, 1, 2, 3, 5, 8, 13, 21, 34
- **Special Cards**: 
  - ☕ (Coffee break - need more discussion)
  - ? (Unknown - insufficient information)

### Keyboard Shortcuts

- Number keys `1-9`: Select voting cards
- `Space`: Start/end voting (session owner only)

## Configuration

The application uses environment variables and sensible defaults:

- **Port**: 8080 (hardcoded in main.go)
- **Database**: SQLite file `poker.db` in working directory
- **Session Duration**: 6 hours with auto-renewal on activity

## Database

The application uses SQLite with automatic migrations. The database file (`poker.db`) is created automatically on first run.

### Tables

- `users` - Session-based user accounts
- `sessions` - Planning sessions
- `tickets` - Items to estimate
- `votes` - User votes on tickets
- `participants` - Session membership
- `recent_emojis` - User emoji history

## Real-time Features

The application uses Server-Sent Events (SSE) for real-time updates:

- User join/leave notifications
- Vote submissions
- Voting start/end events
- Ticket changes
- Emoji reactions with physics animations

## Security Features

- Input validation and sanitization
- HTML escaping to prevent XSS
- Session-based authentication
- CSRF protection for state-changing operations
- Rate limiting considerations for emoji reactions

## Browser Support

- Modern browsers with SSE support
- Mobile browsers (iOS Safari, Android Chrome)
- Keyboard accessibility
- Screen reader compatible

## Development

### Running in Development

```bash
# Run with automatic restart on changes
go run cmd/server/main.go

# Or build and run
go build -o poker-planning cmd/server/main.go
./poker-planning
```

### Database Migrations

Migrations are handled automatically by Goose on application startup. Migration files are located in `internal/database/migrations/`.

### Adding New Features

1. Add database migrations if needed
2. Update models in `internal/models/`
3. Implement business logic in `internal/services/`
4. Add HTTP handlers in `internal/handlers/`
5. Update templates and static assets
6. Add routes in `cmd/server/main.go`

## License

This project is provided as-is for educational and demonstration purposes.

## Contributing

This is a demonstration project. Feel free to use it as a starting point for your own planning poker application.