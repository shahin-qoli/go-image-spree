package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter the folder path: ")
		folderPath, _ := reader.ReadString('\n')
		fmt.Print("Enter the Token: ")
		token, _ := reader.ReadString('\n')
		fmt.Print("Enter the Suffix: ")
		suffix, _ := reader.ReadString('\n')

		folderPath = strings.TrimSpace(folderPath)
		token = strings.TrimSpace(token)
		suffix = strings.TrimSpace(suffix)

		dir, err := os.Open(folderPath)
		if err != nil {
			fmt.Println("Error reading files", err)
			continue // Retry loop
		}

		files, err := dir.Readdir(-1)
		dir.Close() // Close the directory after reading

		if err != nil {
			fmt.Println("Error reading files", err)
			continue // Retry loop
		}

		for _, file := range files {
			filePath := filepath.Join(folderPath, file.Name())
			fileObj, err := os.Open(filePath)
			if err != nil {
				fmt.Println("Error reading files to upload", err)
				continue // Skip to next file
			}

			sku := file.Name()
			sku = strings.Split(sku, ".")[0]
			var isProduct bool
			if strings.HasSuffix(sku, suffix) {
				isProduct = true
				sku = strings.TrimSuffix(sku, suffix)
			} else {
				isProduct = false
			}

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("image", filepath.Base(filePath))
			if err != nil {
				fileObj.Close() // Close the file before skipping
				fmt.Println("Error creating form file", err)
				continue // Skip to next file
			}

			_, err = io.Copy(part, fileObj)
			fileObj.Close() // Close the file after reading

			if err != nil {
				fmt.Println("Error copying file content", err)
				continue // Skip to next file
			}

			err = writer.Close()
			if err != nil {
				fmt.Println("Error closing writer", err)
				continue // Skip to next file
			}

			fmt.Println("SKU:", sku)
			apiURL := fmt.Sprintf("http://94.101.181.98:4500/api/v2/storefront/products_brx/create_product_image?is_product=%v&sku=%v", isProduct, sku)
			req, err := http.NewRequest("POST", apiURL, body)
			if err != nil {
				fmt.Println("Error creating request", err)
				continue // Retry loop
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
			req.Header.Set("Content-Type", writer.FormDataContentType())

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Error sending request", err)
				continue // Retry loop
			}

			responseBody, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close() // Close the response body after reading

			fmt.Println(responseBody)
			if err != nil {
				fmt.Println("Error reading response body", err)
				continue // Retry loop
			}
			fmt.Printf("**** Result for SKU: %v ****\n", sku)
			fmt.Printf("Response Status: %s\n", resp.Status)
			fmt.Printf("Response Body: %s\n", responseBody)
		}
	}
}
