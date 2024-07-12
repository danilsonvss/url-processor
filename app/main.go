package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var rdb *redis.Client

func main() {
	// Configuração do cliente Redis
	rdb = redis.NewClient(&redis.Options{
		Addr: "redis:6379", // Nome do serviço do Redis conforme definido no docker-compose.yml
		DB:   0,
	})

	// Configuração do servidor HTTP com Gin
	router := gin.Default()
	router.POST("/urls", saveURLHandler)
	router.GET("/urls", listURLsHandler)

	// Iniciar o processamento das URLs em uma goroutine separada
	go processURLs()

	// Iniciar o servidor HTTP
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Falha ao iniciar o servidor HTTP: %v", err)
	}
}

func saveURLHandler(c *gin.Context) {
	url := c.PostForm("url")

	// Verificar se a URL foi fornecida
	if url == "" {
		log.Printf("URL não fornecida")
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL não fornecida"})
		return
	}

	// Salva a URL no Redis
	err := rdb.LPush(context.Background(), "urls", url).Err()
	if err != nil {
		log.Printf("Erro ao salvar URL no Redis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar URL"})
		return
	}

	log.Printf("URL salva no Redis: %s", url)
	c.JSON(http.StatusOK, gin.H{"message": "URL salva com sucesso"})
}

func listURLsHandler(c *gin.Context) {
	// Obter as 10 últimas URLs da lista 'urls' no Redis
	urls, err := rdb.LRange(context.Background(), "urls", 0, 9).Result()
	if err != nil {
		log.Printf("Erro ao obter URLs do Redis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao obter URLs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"urls": urls})
}

func processURLs() {
	for {
		// LPop para pegar a URL mais antiga
		result, err := rdb.LPop(context.Background(), "urls").Result()
		if err != nil {
			if err != redis.Nil {
				log.Printf("Erro ao obter URL do Redis: %v", err)
			}
			time.Sleep(time.Second) // Espera um segundo antes de tentar novamente
			continue
		}

		// Processar a URL em uma goroutine separada
		go processURL(result)
	}
}

func processURL(url string) {
	// Simulação de processamento da URL
	log.Printf("Chamando URL: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Erro ao chamar a URL %s: %v", url, err)
		// Você pode implementar lógica de retry aqui se necessário
		return
	}
	defer resp.Body.Close()

	log.Printf("URL %s chamada com sucesso, status: %s", url, resp.Status)
	// Aqui você pode processar a resposta da chamada se necessário
}
