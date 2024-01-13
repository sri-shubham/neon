# Neon: Light Up Your API with Golang Magic ✨

[![Go Report Card](https://goreportcard.com/badge/github.com/sri-shubham/neon)](https://goreportcard.com/report/github.com/sri-shubham/neon)
[![GitHub issues](https://img.shields.io/github/issues/sri-shubham/neon)](https://github.com/sri-shubham/neon/issues)
[![GitHub stars](https://img.shields.io/github/stars/sri-shubham/neon)](https://github.com/sri-shubham/neon/stargazers)

Hey fellow developers! 👋 Welcome to Neon, the Rest framework that adds a touch of magic to your API development. With Neon, handling HTTP routes is as enchanting as a wizard's spell. Let's dive into the spellbook and explore the wonders of Neon! 🌟

## Key Features ✨

**Struct Tag Magic:**
Neon makes HTTP handler metadata a breeze with struct tags. Think of them as little notes to your code, making it look less like a jungle and more like a well-organized garden.

**Annotation-Like Annotations:**
Golang might not have native annotations, but Neon says, "Who needs 'em?" Struct tags step up to the plate, mimicking those fancy annotations and making your code look swanky and readable.

## Getting Started 🚀

1. **Installation:**
   ```bash
   go get -u github.com/sri-shubham/neon
   ```

2. **Create Your Magic Server:**
   ```go
   package main

   import (
       "fmt"
       "github.com/sri-shubham/neon"
   )

   func main() {
       app := neon.New()
       // Add your spells here ✨
       fmt.Println(app.Run())
   }
   ```

3. **Defining Spells (Services and Handlers):**
   ```go
   package main

   import (
       "fmt"
       "net/http"
       "github.com/sri-shubham/neon"
   )

   type UserService struct {
       neon.Module `base:"/user" v:"1"`
       getUser     neon.Get
   }

   func (s UserService) GetUser(w http.ResponseWriter, r *http.Request) {
       fmt.Fprint(w, "Hello, Neon Magic!")
   }
   ```

4. **Run Your Magic Server:**
   ```bash
   go run test/main.go
   ```

5. **Explore the Glow:**
   Neon adds a glow to your routes:
   - Base route for UserService: `/user`
   - Version: `v1`
   - Route for `getUser`: `/user/v1`

## Join the Coding Wizardry! 🧙‍♂️

Ready to illuminate your API development with Neon? Join the coding wizardry and explore the full potential of Neon on [GitHub](https://github.com/sri-shubham/neon). Let the coding magic begin! 🚀✨
