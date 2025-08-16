package neon

import (
	"fmt"

	"github.com/fatih/color"
)

var yellow = color.New(color.FgYellow).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var colors = [](func(a ...interface{}) string){yellow, red, blue, green}

func printLogo() {
	logo := `
	$$\   $$\                               
	$$$\  $$ |                              
	$$$$\ $$ | $$$$$$\   $$$$$$\  $$$$$$$\  
	$$ $$\$$ |$$  __$$\ $$  __$$\ $$  __$$\ 
	$$ \$$$$ |$$$$$$$$ |$$ /  $$ |$$ |  $$ |
	$$ |\$$$ |$$   ____|$$ |  $$ |$$ |  $$ |
	$$ | \$$ |\$$$$$$$\ \$$$$$$  |$$ |  $$ |
	\__|  \__| \_______| \______/ \__|  \__|
`
	c := -1
	for i, ch := range logo {
		if i%10 == 0 {
			c = (c + 1) % len(colors)
		}
		fmt.Print(colors[c](string(ch)))
	}
	fmt.Println()
	color.Unset()
}

func printInfo(app *App) {
	app.Logger.Info("Version", "version", ver)
	app.Logger.Info("Environment", "env", app.Env.String())
}
