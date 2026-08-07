package main 

import (

	"log"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default() //это функция которая создает новый экземпляр Gin для работы с маршрутами и запросами

	r.GET("/health", func(ctx *gin.Context) { //этот маршрут возвращает JSON с статусом "ok" при GET запросе на "/health"
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	if err := r.Run(":9001"); err != nil { //этот код запускает сервер Gin на порту 9001 и ловит ошибки, если они возникают
		log.Fatal(err)
	}
}