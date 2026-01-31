# OAuth2 Authentication for Google Drive API

Complete implementation of OAuth2 authentication flow for Google Drive API in Go.

## Full Authentication Implementation

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/drive/v3"
)

// getClient retrieves an authenticated HTTP client
func getClient(config *oauth2.Config) *http.Client {
    tokFile := "token.json"
    tok, err := tokenFromFile(tokFile)
    if err != nil {
        tok = getTokenFromWeb(config)
        saveToken(tokFile, tok)
    }
    return config.Client(context.Background(), tok)
}

// getTokenFromWeb requests a token from the web
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
    authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
    fmt.Printf("Go to the following link in your browser then type the "+
        "authorization code: \n%v\n", authURL)

    var authCode string
    if _, err := fmt.Scan(&authCode); err != nil {
        log.Fatalf("Unable to read authorization code: %v", err)
    }

    tok, err := config.Exchange(context.TODO(), authCode)
    if err != nil {
        log.Fatalf("Unable to retrieve token from web: %v", err)
    }
    return tok
}

// tokenFromFile retrieves a token from a local file
func tokenFromFile(file string) (*oauth2.Token, error) {
    f, err := os.Open(file)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    tok := &oauth2.Token{}
    err = json.NewDecoder(f).Decode(tok)
    return tok, err
}

// saveToken saves a token to a file path
func saveToken(path string, token *oauth2.Token) {
    fmt.Printf("Saving credential file to: %s\n", path)
    f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
    if err != nil {
        log.Fatalf("Unable to cache oauth token: %v", err)
    }
    defer f.Close()
    json.NewEncoder(f).Encode(token)
}
```

## Setup Steps

### 1. Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable Google Drive API for the project

### 2. Create OAuth2 Credentials

1. Go to "Credentials" in Google Cloud Console
2. Click "Create Credentials" > "OAuth client ID"
3. Select "Desktop app" as application type
4. Download the credentials JSON file
5. Save as `credentials.json` in your project

### 3. credentials.json Format

```json
{
  "installed": {
    "client_id": "your-client-id.apps.googleusercontent.com",
    "project_id": "your-project-id",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_secret": "your-client-secret",
    "redirect_uris": ["http://localhost"]
  }
}
```

## OAuth2 Scopes

Choose the appropriate scope based on your application needs:

```go
// Full access to all files
config, err := google.ConfigFromJSON(b, drive.DriveScope)

// Access only to files created by the app
config, err := google.ConfigFromJSON(b, drive.DriveFileScope)

// Read-only access to file metadata
config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)

// Read-only access to files and metadata
config, err := google.ConfigFromJSON(b, drive.DriveReadonlyScope)

// Multiple scopes
config, err := google.ConfigFromJSON(b, 
    drive.DriveScope,
    "https://www.googleapis.com/auth/drive.appdata")
```

## Token Management

### Token Storage

The `token.json` file stores:
- Access token (short-lived, ~1 hour)
- Refresh token (long-lived, used to get new access tokens)
- Expiry time
- Token type

### Token Refresh

The OAuth2 library automatically refreshes expired tokens when using `config.Client()`:

```go
client := config.Client(context.Background(), tok)
// Client automatically refreshes token when needed
```

### Manual Token Refresh

```go
tokenSource := config.TokenSource(context.Background(), tok)
newToken, err := tokenSource.Token()
if err != nil {
    log.Fatalf("Unable to refresh token: %v", err)
}
// Save new token
saveToken("token.json", newToken)
```

## Complete Example

```go
package main

import (
    "context"
    "log"
    "os"

    "golang.org/x/oauth2/google"
    "google.golang.org/api/drive/v3"
    "google.golang.org/api/option"
)

func main() {
    ctx := context.Background()
    
    // Read credentials
    b, err := os.ReadFile("credentials.json")
    if err != nil {
        log.Fatalf("Unable to read client secret file: %v", err)
    }

    // Configure OAuth2
    config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
    if err != nil {
        log.Fatalf("Unable to parse client secret file to config: %v", err)
    }
    
    // Get authenticated client
    client := getClient(config)

    // Create Drive service
    srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
    if err != nil {
        log.Fatalf("Unable to retrieve Drive client: %v", err)
    }

    // Use the service
    r, err := srv.Files.List().PageSize(10).
        Fields("nextPageToken, files(id, name)").Do()
    if err != nil {
        log.Fatalf("Unable to retrieve files: %v", err)
    }
    
    for _, file := range r.Files {
        log.Printf("%s (%s)\n", file.Name, file.Id)
    }
}
```

## Troubleshooting

### Invalid Grant Error

If you get "invalid_grant" error:
1. Delete `token.json`
2. Run the application again to re-authorize
3. Make sure system clock is synchronized

### Scope Changes

When you change scopes in your code:
1. Delete existing `token.json`
2. Re-run authorization flow
3. User will need to grant new permissions

### Token Security

- Never commit `credentials.json` or `token.json` to version control
- Add to `.gitignore`:
  ```
  credentials.json
  token.json
  ```
- Store credentials securely in production (environment variables, secret managers)
