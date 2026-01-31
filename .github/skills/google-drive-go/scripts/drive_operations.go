package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"google.golang.org/api/drive/v3"
)

// ListFiles lists files from Google Drive with pagination
func ListFiles(srv *drive.Service, pageSize int64, query string) error {
	pageToken := ""
	for {
		call := srv.Files.List().
			PageSize(pageSize).
			Fields("nextPageToken, files(id, name, mimeType, size, createdTime)")

		if query != "" {
			call = call.Q(query)
		}

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		r, err := call.Do()
		if err != nil {
			return fmt.Errorf("unable to retrieve files: %v", err)
		}

		for _, file := range r.Files {
			fmt.Printf("%-40s %s\n", file.Name, file.Id)
		}

		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return nil
}

// UploadFile uploads a file to Google Drive
func UploadFile(srv *drive.Service, localPath, driveName string, parentFolderID string) (*drive.File, error) {
	file := &drive.File{
		Name: driveName,
	}

	if parentFolderID != "" {
		file.Parents = []string{parentFolderID}
	}

	content, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %v", err)
	}
	defer content.Close()

	createdFile, err := srv.Files.Create(file).Media(content).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create file: %v", err)
	}

	return createdFile, nil
}

// DownloadFile downloads a file from Google Drive
func DownloadFile(srv *drive.Service, fileID, localPath string) error {
	// Get file metadata
	file, err := srv.Files.Get(fileID).Do()
	if err != nil {
		return fmt.Errorf("unable to get file metadata: %v", err)
	}

	// Download content
	resp, err := srv.Files.Get(fileID).Download()
	if err != nil {
		return fmt.Errorf("unable to download file: %v", err)
	}
	defer resp.Body.Close()

	// Save to local file
	outPath := localPath
	if outPath == "" {
		outPath = file.Name
	}

	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("unable to create local file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("unable to save file: %v", err)
	}

	return nil
}

// CreateFolder creates a folder in Google Drive
func CreateFolder(srv *drive.Service, name string, parentFolderID string) (*drive.File, error) {
	folder := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
	}

	if parentFolderID != "" {
		folder.Parents = []string{parentFolderID}
	}

	createdFolder, err := srv.Files.Create(folder).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create folder: %v", err)
	}

	return createdFolder, nil
}

// SearchFiles searches for files matching a query
func SearchFiles(srv *drive.Service, query string) ([]*drive.File, error) {
	r, err := srv.Files.List().
		Q(query).
		Fields("files(id, name, mimeType, size, createdTime)").
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to search files: %v", err)
	}

	return r.Files, nil
}

// ShareFile shares a file with a user
func ShareFile(srv *drive.Service, fileID, email, role string) error {
	permission := &drive.Permission{
		Type:         "user",
		Role:         role, // reader, writer, commenter
		EmailAddress: email,
	}

	_, err := srv.Permissions.Create(fileID, permission).Do()
	if err != nil {
		return fmt.Errorf("unable to share file: %v", err)
	}

	return nil
}

// DeleteFile moves a file to trash
func DeleteFile(srv *drive.Service, fileID string) error {
	err := srv.Files.Delete(fileID).Do()
	if err != nil {
		return fmt.Errorf("unable to delete file: %v", err)
	}
	return nil
}

// Example usage
func main() {
	ctx := context.Background()
	// Assume srv is your authenticated Drive service
	// srv := getDriveService(ctx)

	log.Println("Google Drive API helper functions")
	log.Println("This file provides reusable functions for common Drive operations")

	// Example: List files
	// err := ListFiles(srv, 10, "")

	// Example: Upload file
	// file, err := UploadFile(srv, "local.txt", "remote.txt", "")

	// Example: Download file
	// err := DownloadFile(srv, "file-id", "downloaded.txt")

	// Example: Create folder
	// folder, err := CreateFolder(srv, "My Folder", "")

	// Example: Search files
	// files, err := SearchFiles(srv, "name contains 'report'")

	// Example: Share file
	// err := ShareFile(srv, "file-id", "user@example.com", "reader")

	_ = ctx
}
