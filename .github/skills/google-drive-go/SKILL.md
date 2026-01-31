---
name: google-drive-go
description: Expert guidance for using the Google Drive API with Go, including OAuth2 authentication, file operations (list, upload, download, search), folder management, permissions, and best practices. Use when building Go applications that interact with Google Drive, implementing file storage/sync features, managing Drive permissions, or any Google Drive integration tasks in Go.
---

# Google Drive Go API

This skill provides guidance for working with Google Drive API v3 in Go applications.

## Quick Start

### 1. Setup Authentication

Google Drive API uses OAuth2 for authentication. Set up credentials:

```go
import (
    "context"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/drive/v3"
    "google.golang.org/api/option"
)

// Read credentials from JSON file (from Google Cloud Console)
b, err := os.ReadFile("credentials.json")
if err != nil {
    log.Fatalf("Unable to read client secret file: %v", err)
}

// Configure OAuth2 with required scopes
config, err := google.ConfigFromJSON(b, drive.DriveScope)
if err != nil {
    log.Fatalf("Unable to parse client secret file: %v", err)
}

// Get authenticated client
client := getClient(config)

// Create Drive service
srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
```

### 2. Common OAuth2 Scopes

- `drive.DriveScope` - Full access to Drive
- `drive.DriveFileScope` - Access to files created by the app
- `drive.DriveMetadataReadonlyScope` - Read-only metadata access
- `drive.DriveReadonlyScope` - Read-only access

### 3. Authentication Helper Functions

See `references/auth.md` for complete OAuth2 implementation including token storage and retrieval.

## Common Operations

### List Files

```go
r, err := srv.Files.List().
    PageSize(10).
    Fields("nextPageToken, files(id, name, mimeType, size)").
    Do()
if err != nil {
    log.Fatalf("Unable to retrieve files: %v", err)
}

for _, file := range r.Files {
    fmt.Printf("%s (%s)\n", file.Name, file.Id)
}
```

### Upload File

```go
file := &drive.File{
    Name: "myfile.txt",
    MimeType: "text/plain",
}

content, err := os.Open("local-file.txt")
if err != nil {
    log.Fatalf("Unable to read file: %v", err)
}
defer content.Close()

createdFile, err := srv.Files.Create(file).Media(content).Do()
if err != nil {
    log.Fatalf("Unable to create file: %v", err)
}
fmt.Printf("File ID: %s\n", createdFile.Id)
```

### Download File

```go
// Get file metadata
file, err := srv.Files.Get(fileId).Do()

// Download content
resp, err := srv.Files.Get(fileId).Download()
if err != nil {
    log.Fatalf("Unable to download file: %v", err)
}
defer resp.Body.Close()

// Save to local file
out, err := os.Create(file.Name)
if err != nil {
    log.Fatalf("Unable to create local file: %v", err)
}
defer out.Close()

io.Copy(out, resp.Body)
```

### Search Files

```go
query := "name contains 'report' and mimeType='application/pdf'"
r, err := srv.Files.List().
    Q(query).
    Fields("files(id, name, createdTime)").
    Do()
```

Common query patterns:
- `name = 'filename.txt'` - Exact name match
- `name contains 'keyword'` - Partial match
- `mimeType = 'application/pdf'` - Filter by type
- `'parent-folder-id' in parents` - Files in folder
- `trashed = false` - Exclude trashed files

### Create Folder

```go
folder := &drive.File{
    Name:     "My Folder",
    MimeType: "application/vnd.google-apps.folder",
}

createdFolder, err := srv.Files.Create(folder).Do()
if err != nil {
    log.Fatalf("Unable to create folder: %v", err)
}
```

### Manage Permissions

```go
permission := &drive.Permission{
    Type: "user",
    Role: "writer",
    EmailAddress: "user@example.com",
}

_, err := srv.Permissions.Create(fileId, permission).Do()
if err != nil {
    log.Fatalf("Unable to share file: %v", err)
}
```

Permission roles: `owner`, `organizer`, `fileOrganizer`, `writer`, `commenter`, `reader`

## Advanced Topics

### Pagination

Handle large result sets with pagination:

```go
pageToken := ""
for {
    r, err := srv.Files.List().
        PageSize(100).
        PageToken(pageToken).
        Fields("nextPageToken, files(id, name)").
        Do()
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    for _, file := range r.Files {
        // Process file
    }
    
    pageToken = r.NextPageToken
    if pageToken == "" {
        break
    }
}
```

### File Metadata Fields

Use the `Fields` parameter to specify which fields to return:

```go
Fields("files(id, name, mimeType, size, createdTime, modifiedTime, webViewLink)")
```

Common fields: `id`, `name`, `mimeType`, `size`, `createdTime`, `modifiedTime`, `parents`, `webViewLink`, `iconLink`, `thumbnailLink`

### Export Google Docs

Export Google Workspace files to different formats:

```go
// Export Google Doc as PDF
resp, err := srv.Files.Export(fileId, "application/pdf").Download()

// Export Google Sheet as Excel
resp, err := srv.Files.Export(fileId, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet").Download()
```

## Best Practices

1. **Use appropriate scopes** - Request minimum necessary permissions
2. **Handle rate limits** - Implement exponential backoff for retries
3. **Batch operations** - Use batch requests for multiple operations
4. **Cache tokens** - Store OAuth2 tokens to avoid repeated authorization
5. **Use fields parameter** - Only request needed fields to reduce bandwidth
6. **Handle pagination** - Always implement pagination for list operations
7. **Check mime types** - Validate file types before operations
8. **Error handling** - Properly handle API errors and network issues

## Common MIME Types

- Plain text: `text/plain`
- PDF: `application/pdf`
- JPEG: `image/jpeg`
- PNG: `image/png`
- Google Docs: `application/vnd.google-apps.document`
- Google Sheets: `application/vnd.google-apps.spreadsheet`
- Google Slides: `application/vnd.google-apps.presentation`
- Folder: `application/vnd.google-apps.folder`

## Resources

- **Authentication details**: See `references/auth.md` for complete OAuth2 implementation
- **API documentation**: https://developers.google.com/drive/api/v3/reference
- **Go client library**: https://pkg.go.dev/google.golang.org/api/drive/v3

## Installation

```bash
go get google.golang.org/api/drive/v3
go get golang.org/x/oauth2/google
```
