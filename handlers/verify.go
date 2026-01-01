package handlers

import (
	"mailnull/api/internal/verifier"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VerifyRequest struct {
	Email  string   `json:"email"`
	Emails []string `json:"emails"`
}

type VerifyResponse struct {
	Results []verifier.Result `json:"results"`
}

type VerifyHandler struct {
	pool *verifier.WorkerPool
}

func NewVerifyHandler(pool *verifier.WorkerPool) *VerifyHandler {
	return &VerifyHandler{pool: pool}
}

func (h *VerifyHandler) Verify(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var emails []string
	isSingleEmail := false

	if req.Email != "" {
		emails = append(emails, req.Email)
		isSingleEmail = true
	}
	if len(req.Emails) > 0 {
		emails = append(emails, req.Emails...)
		isSingleEmail = false
	}

	if len(emails) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No email provided"})
		return
	}

	resultChan := make(chan verifier.Result, len(emails))

	for _, email := range emails {
		h.pool.Submit(email, resultChan)
	}

	results := make([]verifier.Result, 0, len(emails))
	for i := 0; i < len(emails); i++ {
		res := <-resultChan
		results = append(results, res)
	}

	close(resultChan)

	if isSingleEmail {
		c.JSON(http.StatusOK, results[0])
	} else {
		c.JSON(http.StatusOK, VerifyResponse{
			Results: results,
		})
	}
}
