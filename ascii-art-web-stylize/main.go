package main

import (
	"fmt"
	"html/template" // Package for working with HTML templates
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// UserInput represents user input data.
type UserInput struct {
	UserText   string   // User's input text
	BannerType string   // Selected banner type
	OutputArr  []string // Array of generated ASCII art
}

var templates *template.Template

func init() {
	// Parse the HTML templates during initialization
	templates = template.Must(template.ParseGlob("templates/*.html"))
}

// WelcomeHandler handles requests to the root URL ("/").
func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// If the URL path is not "/", handle it as a 404 error
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		renderTemplate(w, r, "index.html", nil)
	case http.MethodPost:
		processForm(w, r)
	}
}

func processForm(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	userText := r.Form.Get("userText")     // Get the user's input text from the form
	userBanner := r.Form.Get("bannerType") // Get the selected banner type from the form
	
	if userText == "" {
		// If userText is empty, handle it as a 400 error
		errorHandler(w, r, http.StatusBadRequest)
		return
	}

	var bannerFile string
	switch userBanner {
	case "Standard":
		bannerFile = "standard.txt"
	case "Shadow":
		bannerFile = "shadow.txt"
	case "Thinkertoy":
		bannerFile = "thinkertoy.txt"
	default:
		bannerFile = "standard.txt"
	}

	// Read the banner file based on the selected banner type
	file, err := os.Open(bannerFile)
	if err != nil {
		// Handle the error (perhaps log it) and return an internal server error
		errorHandler(w, r, http.StatusInternalServerError)
		return
	}
	defer file.Close()
	
	data, err := io.ReadAll(file)
	if err != nil {
		// Handle the error (perhaps log it) and return an internal server error
		errorHandler(w, r, http.StatusInternalServerError)
		return
	}
	
	// Split the banner file into lines
	lines := strings.Split(string(data), "\n")
	// Split the user's input text into lines
	inputLines := strings.Split(userText, "\n")
	
	var asciiArr []string
	// Generate ASCII art for each line in the user's input text
	for _, line := range inputLines {
		words := strings.FieldsFunc(line, strSplit)
		for _, word := range words {
			if word == "" {
				continue
			}
			asciiArr = append(asciiArr, printWord(word, lines)...)
		}
	}

	myUser := UserInput{
		UserText:   userText,   // Set the UserText field of the UserInput struct to the userText variable
		BannerType: userBanner, // Set the BannerType field of the UserInput struct to the userBanner variable
		OutputArr:  asciiArr,   // Set the OutputArr field of the UserInput struct to the asciiArr variable
	}

	renderTemplate(w, r, "index.html", myUser) // Render the "index.html" template with myUser as the data
}

func renderTemplate(w http.ResponseWriter, r *http.Request, tmplName string, data interface{}) {
    tmpl, err := template.ParseFiles("templates/" + tmplName)
    if err != nil {
        // Log the error
        log.Println("Error parsing template:", err)

        // Render the error template with a user-friendly error message
        errorHandler(w, r, http.StatusInternalServerError)
        return
    }

    err = tmpl.Execute(w, data)
    if err != nil {
        // Log the error and return a simple error message
        log.Println("Error executing template:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}


func main() {

	http.Handle("/style/", http.StripPrefix("/style/", http.FileServer(http.Dir("style"))))
	
	// Register the WelcomeHandler function to handle requests to the root path "/"
	http.HandleFunc("/", WelcomeHandler)

	// Print a message indicating that the server is listening on port 8000
	fmt.Println("Server started at http://localhost:8000/")

	// Start the HTTP server on port 8000
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		// If an error occurs while starting the server, log the error and exit
		log.Fatal("Error starting server:", err)
	}
}

func printWord(word string, lines []string) []string {
	
	var strArray []string

	for j := 1; j < 9; j++ {
		str := ""
		// Iterate over each letter in the word
		for _, letter := range word {
			val := int(letter)
			line := (val - 32) * 9
		
			if line+j >= len(lines) || line+j < 0 {
				// Handle this scenario gracefully, log an error, or skip the letter
				continue
			}
			
			str += lines[line+j]
		}
		
		// Append the row of ASCII art to the string array
		strArray = append(strArray, str)
	}
	// Return the generated ASCII art
	return strArray
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
    // Set the HTTP response status code
    w.WriteHeader(status)

    // Define the template filename based on the status code
    templateFile := "error.html"

    // Create data to pass to the template
    data := struct {
        StatusCode int
        StatusText string
        Description string
    }{
        StatusCode: status,
        StatusText: http.StatusText(status),
        Description: getErrorDescription(status),
    }

    // Render the corresponding error template
    renderTemplate(w, r, templateFile, data)
}

func getErrorDescription(status int) string {
    // Provide custom error descriptions based on status code
    switch status {
    case http.StatusBadRequest:
        return "Sorry, the request is invalid or incomplete."
    case http.StatusNotFound:
        return "Sorry, the page you are looking for might be missing or the URL is incorrect."
    case http.StatusInternalServerError:
        return "Sorry, something went wrong on our end. We are working to fix the issue."
    default:
        return "An unexpected error occurred."
    }
}


//handle space and enter
func strSplit(r rune) bool {
	return r == '\n' 
}