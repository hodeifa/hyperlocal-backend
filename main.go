   package main

   import (
       "fmt"
       "net/http"
   )

   func main() {
       http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
           fmt.Fprintf(w, "Halo dari Codespace Go!, Mantap jiwa bosque")
       })
       fmt.Println("Server jalan di :8080")
       http.ListenAndServe(":8080", nil)
   }