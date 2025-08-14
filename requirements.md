# Sprint Planning Poker Web App Requirements

## 1. Functional Requirements

### 1.1 User Management
- **Session-based authentication** (no persistent accounts)
- **Username modal** appears on first visit to set display name
- Username stored in session cookie (6-hour expiration, renewed on activity)
- Users identified by session ID + chosen username

### 1.2 Planning Session Management
- **Create session**: Any user can start a new planning session
- **Unique session URL**: Shareable link for inviting team members
- **Session persistence**: 6-hour lifetime with auto-renewal on activity
- **Owner controls**:
  - Add/edit/remove tickets
  - Start/stop voting rounds
  - Advance to next ticket
  - Transfer ownership to another participant
- **Participant list**: Real-time display of all connected users

### 1.3 Voting Cards
- **Fibonacci sequence**: 0, 1, 2, 3, 5, 8, 13, 21, 34
- **Special cards**: ☕ (coffee break), ? (unknown)
- **Card selection**: Click to select during active voting
- **Hidden votes**: Cards remain face-down until voting ends
- **Vote changing**: Allowed during review phase

### 1.4 Ticket Management
- **Simple ticket creation**: Title/description text input
- **Ticket queue**: List of tickets to estimate
- **Current ticket display**: Prominent display of active ticket
- **Progress indicator**: Show position in ticket queue (e.g., "3 of 7")

### 1.5 Voting Workflow
1. **Pre-voting**: Session owner adds tickets
2. **Voting phase**: 
   - Owner starts voting
   - All participants select cards
   - Real-time indicator of who has/hasn't voted
3. **Reveal phase**:
   - All cards flip simultaneously
   - Display histogram/bar chart of vote distribution
   - Allow vote changes
4. **Next ticket**: Owner advances when ready

### 1.6 Emoji Reactions
- **Trigger**: Hover over user avatar/name to show emoji selector
- **Quick picker**: 5 most recently used emojis
- **Full picker**: "+" button opens comprehensive emoji selector
- **Animation**: Physics-based trajectory from sender to recipient
- **Visual feedback**: Emoji "impacts" target user with particle effect

## 2. Technical Requirements

### 2.1 Backend Stack
- **Language**: Go
- **Router**: Chi (github.com/go-chi/chi/v5)
- **Templates**: Go html/template with HTMX integration
- **Database**: SQLite with sqlc for type-safe queries
- **Migrations**: Goose for schema management
- **Real-time**: Server-Sent Events (SSE) for low-latency updates

### 2.2 Frontend Stack
- **Interactivity**: HTMX for dynamic updates
- **Styling**: Tailwind CSS with Material Design principles
- **Animations**: CSS transitions + JavaScript for emoji physics
- **Icons**: Material Icons or Heroicons

### 2.3 Database Schema
```sql
-- Users (session-based)
CREATE TABLE users (
    id TEXT PRIMARY KEY, -- session_id
    username TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Planning sessions
CREATE TABLE sessions (
    id TEXT PRIMARY KEY, -- UUID
    name TEXT NOT NULL,
    owner_id TEXT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tickets
CREATE TABLE tickets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    final_estimate INTEGER,
    position INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Votes
CREATE TABLE votes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ticket_id INTEGER NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id),
    vote_value TEXT NOT NULL, -- stores fibonacci number or special values
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ticket_id, user_id)
);

-- Session participants
CREATE TABLE participants (
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (session_id, user_id)
);

-- Recent emojis
CREATE TABLE recent_emojis (
    user_id TEXT NOT NULL REFERENCES users(id),
    emoji TEXT NOT NULL,
    used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, emoji)
);
```

### 2.4 Real-time Communication
- **SSE endpoints** for:
  - Session state updates
  - Vote submissions
  - User join/leave events
  - Emoji reactions
- **Message types**:
  - ```user-joined```, ```user-left```
  - ```voting-started```, ```voting-ended```
  - ```vote-cast```, ```vote-changed```
  - ```emoji-sent```
  - ```ticket-changed```
  - ```owner-transferred```

## 3. Non-Functional Requirements

### 3.1 Performance
- **Latency**: < 100ms for all user interactions
- **SSE delivery**: < 50ms for message broadcast
- **Concurrent users**: Support 50+ users per session
- **Database queries**: Optimized with appropriate indexes

### 3.2 User Experience
- **Responsive design**: Mobile, tablet, and desktop support
- **Accessibility**: WCAG 2.1 AA compliance
- **Loading states**: Skeleton screens and progress indicators
- **Error handling**: User-friendly error messages
- **Offline detection**: Alert when connection lost

### 3.3 Security
- **Session validation**: Verify session ownership for admin actions
- **Input sanitization**: Prevent XSS in ticket titles/descriptions
- **Rate limiting**: Prevent emoji spam
- **CSRF protection**: Token validation for state-changing operations

## 4. UI/UX Requirements

### 4.1 Layout Components
- **Header**: App name, session name, user info
- **Sidebar**: Participant list with online indicators
- **Main area**: Current ticket and voting interface
- **Ticket queue**: Collapsible list of upcoming tickets
- **Results panel**: Histogram and individual votes

### 4.2 Visual Design
- **Color scheme**: Material Design 3 color system
- **Typography**: Clear hierarchy with system fonts
- **Spacing**: Consistent 8px grid system
- **Animations**: Smooth transitions (200-300ms)
- **Dark mode**: Optional toggle

### 4.3 Mobile Considerations
- **Touch targets**: Minimum 44x44px
- **Swipe gestures**: Navigate between tickets
- **Simplified layout**: Stack components vertically
- **Emoji picker**: Full-screen on mobile

## 5. Development Phases

### Phase 1: Core Infrastructure
- Database schema and migrations
- Session management
- Basic UI layout
- SSE setup

### Phase 2: Voting Functionality
- Card selection UI
- Vote submission/reveal
- Histogram visualization
- Real-time updates

### Phase 3: Enhanced Features
- Emoji reactions with physics
- Owner controls and transfer
- Ticket management
- Polish and optimizations

Would you like me to start implementing any specific component first, or would you prefer to see the complete project structure and setup instructions?


