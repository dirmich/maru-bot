package main

import (
"bufio"
"fmt"
"os"
)

func main() {
inputFile := "cmd/marubot/main.go"
cleanupFile := "cleanup.go"
outputFile := "cmd/marubot/main.go.new"

f, err := os.Open(inputFile)
if err != nil {
tf("Error opening main.go: %v\n", err)

}
defer f.Close()

out, err := os.Create(outputFile)
if err != nil {
tf("Error creating new file: %v\n", err)

}
defer out.Close()

scanner := bufio.NewScanner(f)
lineCount := 0
for scanner.Scan() {
eCount++
eCount > 1316 {
tln(out, scanner.Text())
}

cleanupData, err := os.ReadFile(cleanupFile)
if err != nil {
tf("Error reading cleanup.go: %v\n", err)

}
out.Write(cleanupData)

fmt.Println("Surgery complete!")
}
