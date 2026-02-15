# Conference Backend API

A Go backend for a conference website with Google OAuth authentication and Firebase integration.

## Features

- **Google OAuth Authentication** via Firebase Auth
- **Firebase Firestore** for data persistence
- **Conference Registration System** - Create, read, update, and delete registrations
- **JWT Token Verification** - Secure API endpoints with Firebase ID tokens
- **CORS Support** - Configurable cross-origin resource sharing
- **Graceful Shutdown** - Proper server shutdown handling

## Prerequisites

- Go 1.21 or later
- Firebase project with Authentication and Firestore enabled
- Google OAuth credentials configured in Firebase

## Firebase Setup

1. Go to the [Firebase Console](https://console.firebase.google.com/)
2. Create a new project or select an existing one
3. Enable **Authentication**:
   - Go to Authentication > Sign-in method
   - Enable **Google** as a sign-in provider
   - Configure OAuth consent screen and add your domain
4. Enable **Firestore Database**:
   - Go to Firestore Database
   - Create a database (start in production or test mode)
5. Generate a **Service Account Key**:
   - Go to Project Settings > Service Accounts
   - Click "Generate new private key"
   - Save the JSON file as `firebase-service-account.json` in the project root

## Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd backend-ITC
   ```

2. Copy the environment example file:
   ```bash
   cp .env.example .env
   ```

3. Configure your `.env` file with your Firebase credentials and settings.

4. Place your Firebase service account JSON file in the project root (or update the path in `.env`).

5. Install dependencies:
   ```bash
   go mod tidy
   ```

## Running the Server

### Development
```bash
go run cmd/server/main.go
```

### Production
```bash
go build -o server cmd/server/main.go
./server
```

## API Endpoints

### Health Check
- `GET /health` - Check if the server is running

### Authentication
- `POST /api/v1/auth/google` - Authenticate with Google (Firebase ID token)
- `POST /api/v1/auth/verify` - Verify an existing token
- `POST /api/v1/auth/logout` - Logout (client-side cleanup)

### User (Protected)
- `GET /api/v1/me` - Get current user profile

### Registrations (Protected)
- `POST /api/v1/registrations` - Create a new registration
- `GET /api/v1/registrations/me` - Get current user's registration
- `PUT /api/v1/registrations/me` - Update current user's registration
- `DELETE /api/v1/registrations/me` - Delete current user's registration

### Admin (Protected)
- `GET /api/v1/admin/registrations` - Get all registrations

## Frontend Integration

### 1. Initialize Firebase in your frontend

```javascript
import { initializeApp } from 'firebase/app';
import { getAuth, signInWithPopup, GoogleAuthProvider } from 'firebase/auth';

const firebaseConfig = {
  apiKey: "YOUR_API_KEY",
  authDomain: "YOUR_PROJECT.firebaseapp.com",
  projectId: "YOUR_PROJECT_ID",
  // ... other config
};

const app = initializeApp(firebaseConfig);
const auth = getAuth(app);
```

### 2. Sign in with Google

```javascript
const provider = new GoogleAuthProvider();

async function signInWithGoogle() {
  try {
    const result = await signInWithPopup(auth, provider);
    const idToken = await result.user.getIdToken();
    
    // Send token to your backend
    const response = await fetch('http://localhost:8080/api/v1/auth/google', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ idToken }),
    });
    
    const data = await response.json();
    console.log('Logged in:', data);
  } catch (error) {
    console.error('Error:', error);
  }
}
```

### 3. Make authenticated requests

```javascript
async function makeAuthenticatedRequest(endpoint) {
  const user = auth.currentUser;
  if (!user) throw new Error('Not authenticated');
  
  const idToken = await user.getIdToken();
  
  const response = await fetch(`http://localhost:8080/api/v1${endpoint}`, {
    headers: {
      'Authorization': `Bearer ${idToken}`,
      'Content-Type': 'application/json',
    },
  });
  
  return response.json();
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Port to run the server on | `8080` |
| `SERVER_HOST` | Host to bind to | `0.0.0.0` |
| `FIREBASE_CREDENTIALS_FILE` | Path to Firebase service account JSON | `firebase-service-account.json` |
| `FIREBASE_PROJECT_ID` | Firebase project ID | - |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | - |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | - |
| `GOOGLE_REDIRECT_URL` | OAuth redirect URL | `http://localhost:8080/auth/google/callback` |
| `SESSION_SECRET` | Secret for session encryption | - |
| `ENVIRONMENT` | `development` or `production` | `development` |
| `FRONTEND_URL` | Frontend URL for CORS | `http://localhost:3000` |

## Project Structure

```
backend-ITC/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── firebase/
│   │   └── firebase.go      # Firebase client initialization
│   ├── handlers/
│   │   ├── auth.go          # Authentication handlers
│   │   └── registration.go  # Registration handlers
│   ├── middleware/
│   │   └── auth.go          # Authentication middleware
│   ├── models/
│   │   └── user.go          # Data models
│   └── router/
│       └── router.go        # Route definitions
├── .env.example             # Example environment file
├── go.mod                   # Go module definition
└── README.md                # This file
```

## Firestore Collections

### `users`
Stores user profiles linked to Firebase Auth.

### `registrations`
Stores conference registration data.

## Security Notes

1. **Never commit** your `firebase-service-account.json` or `.env` file
2. Use **environment variables** in production
3. Enable **Firestore security rules** to protect your data
4. Set **CORS origins** appropriately for production

## License

MIT