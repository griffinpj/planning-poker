# Sprint Planning Poker Architecture

## Overview

This is a real-time web application for agile sprint planning poker built with Go backend and HTMX frontend, using WebSockets for live collaboration.

## Technology Stack

### Backend
- **Go** - Core application language
- **Chi Router** - HTTP routing and middleware
- **SQLite** - Database with sqlc for type-safe queries
- **Goose** - Database migrations
- **Gorilla WebSocket** - Real-time communication

### Frontend
- **HTMX** - Dynamic HTML updates without JavaScript frameworks
- **Tailwind CSS** - Styling with Material Design principles
- **Native WebSocket API** - Real-time updates
- **Go html/template** - Server-side templating

## Application Architecture

### 1. Session-Based User Management
- **No persistent accounts** - users identified by session cookies
- **6-hour session lifetime** - auto-renewed on activity
- **Username modal** - appears on first visit
- Session ID stored in HTTP-only cookie for security

### 2. Real-Time Communication Flow

```
User Action → HTTP Request → Handler → Service → Database
     ↓
WebSocket Broadcast → All Connected Users → UI Updates
```

#### WebSocket Message Types:
- `user-joined`, `user-left` - User presence updates
- `voting-started`, `voting-ended` - Voting phase changes
- `vote-cast` - Individual vote submissions
- `ticket-changed`, `ticket-created` - Ticket management
- `session-ended` - Session termination

### 3. Voting Workflow

1. **Session Creation**: Owner creates session with unique URL
2. **User Joining**: Participants join via shared URL
3. **Ticket Management**: Owner adds/edits tickets for estimation
4. **Voting Phases**:
   - **Pre-voting**: Tickets visible, no voting active
   - **Active voting**: Cards enabled, votes hidden, averages hidden
   - **Results**: All votes revealed, averages shown, vote changes allowed
5. **Progression**: Owner controls ticket selection and voting phases

### 4. Data Models

```go
User {
    ID (session_id), Username, CreatedAt, LastSeen
}

Session {
    ID, Name, OwnerID, CurrentTicketID, IsVotingActive
    Participants []User, Tickets []Ticket
}

Ticket {
    ID, SessionID, Title, Description, FinalEstimate, Position
    Votes []Vote
}

Vote {
    ID, TicketID, UserID, VoteValue, CreatedAt
}
```

### 5. Frontend Architecture

#### Template Structure:
- **base.html** - Layout with WebSocket connection logic
- **session.html** - Main voting interface
- **home.html** - Landing page

#### HTMX Integration:
- **Partial updates** - `/session/{id}/partial` endpoint for content refresh
- **Form submissions** - Ticket creation, voting without page reloads  
- **WebSocket triggers** - Content updates via `htmx.ajax()` calls

#### Card Highlighting Logic:
- **Template-driven** - `{{if and $.UserVote (eq . $.UserVote.VoteValue)}}border-blue-500{{end}}`
- **Persistent** - Survives WebSocket content refreshes
- **Immediate feedback** - Shows selection instantly

### 6. Security & Validation

#### Input Sanitization:
- **Server-side** - All form inputs trimmed and validated
- **XSS Prevention** - Go templates auto-escape HTML
- **CSRF Protection** - Session-based validation

#### Access Control:
- **Session ownership** - Only owners can manage sessions/tickets
- **Participant validation** - Users must be session participants for WebSocket access
- **Vote authorization** - Users can only vote on active tickets

### 7. Database Schema

```sql
users(id, username, created_at, last_seen)
sessions(id, name, owner_id, current_ticket_id, is_voting_active)
tickets(id, session_id, title, description, final_estimate, position)
votes(id, ticket_id, user_id, vote_value)
participants(session_id, user_id, joined_at)
```

### 8. Key Features

#### Voting Cards:
- **Fibonacci sequence**: 0,1,2,3,5,8,13,21,34,55,89,144
- **Special cards**: ☕ (coffee break), ? (unknown)
- **Average calculation**: Excludes special cards, includes numeric votes only

#### Real-Time Updates:
- **Live participant list** - Shows online users
- **Vote indicators** - Real-time voting status
- **Automatic refresh** - Content updates without manual reload
- **Error resilience** - Handles network interruptions gracefully

#### UI/UX:
- **Responsive design** - Mobile, tablet, desktop support
- **Visual feedback** - Card highlighting, vote status, progress indicators
- **Information hiding** - Averages hidden during active voting
- **Owner controls** - Start/stop voting, ticket management, session control

## Request Flow Examples

### Voting Process:
1. User clicks card → `POST /session/{id}/vote`
2. Handler validates and saves vote → Database
3. WebSocket broadcast → `vote-cast` message
4. All clients refresh content → `GET /session/{id}/partial`
5. Template renders with updated state
6. Card highlighting preserved via template logic

### Ticket Creation:
1. Owner submits form → `POST /session/{id}/tickets` (HTMX)
2. Handler creates ticket → Database
3. WebSocket broadcast → `ticket-created` message
4. All clients refresh → Updated sidebar with new ticket
5. Modal closes automatically

## Performance Considerations

- **WebSocket connection pooling** - One connection per user per session
- **Selective updates** - Only refresh content that changed
- **Template caching** - Parsed once at startup
- **SQLite efficiency** - Appropriate indexes, type-safe queries via sqlc
- **Minimal JavaScript** - Relies on HTMX and templates for most interactivity

## Deployment Architecture

```
Client Browser ←→ Go HTTP Server ←→ SQLite Database
        ↕                ↕
   WebSocket         Static Files
   Connection        (CSS/JS/Images)
```

The application is designed as a single binary with embedded templates and static files, making deployment simple and self-contained.